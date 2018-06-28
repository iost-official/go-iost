package rpc

import (
	"context"
	"fmt"
	"reflect"

	"github.com/iost-official/prototype/account"
	"github.com/iost-official/prototype/common"
	"github.com/iost-official/prototype/consensus"
	"github.com/iost-official/prototype/core/block"
	"github.com/iost-official/prototype/core/message"
	"github.com/iost-official/prototype/core/state"
	"github.com/iost-official/prototype/core/tx"
	"github.com/iost-official/prototype/core/txpool"
	"github.com/iost-official/prototype/network"
	"github.com/iost-official/prototype/vm"
	"github.com/iost-official/prototype/vm/lua"
)

//go:generate mockgen -destination mock_rpc/mock_rpc.go -package rpc_mock github.com/iost-official/prototype/rpc CliServer

type BInfo struct {
	Head  block.BlockHead
	TxCnt int
}
type RpcServer struct {
}

// newRpcServer
func newRpcServer() *RpcServer {
	s := &RpcServer{}
	return s
}

func (s *RpcServer) Transfer(ctx context.Context, txinfo *TransInfo) (*PublishRet, error) {
	ret := PublishRet{Code: -1}
	seckey := txinfo.Seckey
	nonce := txinfo.Nonce
	code := txinfo.Contract
	acc, err := account.NewAccount(common.Base58Decode(seckey))
	if err != nil {
		return &ret, fmt.Errorf("NewAccount:%v", err)
	}

	var contract vm.Contract
	parser, _ := lua.NewDocCommentParser(code)
	contract, err = parser.Parse()
	if err != nil {
		return &ret, fmt.Errorf("Parse:%v", err)
	}
	mtx := tx.NewTx(nonce, contract)
	sig, err := tx.SignContract(mtx, acc)
	if !mtx.VerifySigner(sig) {
		return &ret, fmt.Errorf("VerifySigner:%v", err)
	}
	mtx.Signs = append(mtx.Signs, sig)

	if err != nil {
		return &ret, fmt.Errorf("SignContract:%v", err)
	}

	stx, err := tx.SignTx(mtx, acc)
	if err != nil {
		return &ret, fmt.Errorf("SignTx:%v", err)

	}

	err = stx.VerifySelf() //verify Publisher and Signers
	if err != nil {
		return &ret, fmt.Errorf("VerifySelf:%v", err)

	}

	// add servi
	tx.RecordTx(stx, tx.Data.Self())

	//broadcast the tx
	router := network.Route
	if router == nil {
		panic(fmt.Errorf("network.Router shouldn't be nil"))
	}
	broadTx := message.Message{
		Body:    stx.Encode(),
		ReqType: int32(network.ReqPublishTx),
	}
	router.Broadcast(broadTx)
	Cons := consensus.Cons
	if Cons == nil {
		panic(fmt.Errorf("Consensus is nil"))
	}
	txpool.TxPoolS.AddTransaction(&broadTx)
	ret.Code = 0
	ret.Hash = stx.Hash()
	return &ret, nil
}
func (s *RpcServer) PublishTx(ctx context.Context, _tx *Transaction) (*PublishRet, error) {
	fmt.Println("publish")
	ret := PublishRet{Code: -1}
	var tx1 tx.Tx
	if _tx == nil {
		return &ret, fmt.Errorf("argument cannot be nil pointer")
	}
	err := tx1.Decode(_tx.Tx)
	if err != nil {
		return &ret, err
	}

	err = tx1.VerifySelf() //verify Publisher and Signers
	if err != nil {
		return &ret, err
	}

	// add servi
	tx.RecordTx(tx1, tx.Data.Self())

	//broadcast the tx
	router := network.Route
	if router == nil {
		panic(fmt.Errorf("network.Router shouldn't be nil"))
	}
	broadTx := message.Message{
		Body:    tx1.Encode(),
		ReqType: int32(network.ReqPublishTx),
	}
	router.Broadcast(broadTx)
	Cons := consensus.Cons
	if Cons == nil {
		panic(fmt.Errorf("Consensus is nil"))
	}
	txpool.TxPoolS.AddTransaction(&broadTx)
	ret.Code = 0
	ret.Hash = tx1.Hash()
	return &ret, nil
}
func (s *RpcServer) GetTransaction(ctx context.Context, txkey *TransactionKey) (*Transaction, error) {
	if txkey == nil {
		return nil, fmt.Errorf("argument cannot be nil pointer")
	}
	Nonce := txkey.Nonce

	txDb := tx.TxDb
	if txDb == nil {
		panic(fmt.Errorf("TxDb should be nil"))
	}
	tx, err := txDb.(*tx.TxPoolDb).GetByPN(Nonce, txkey.Publisher)
	if err != nil {
		return nil, err
	}

	return &Transaction{Tx: tx.Encode()}, nil
}

func (s *RpcServer) GetTransactionByHash(ctx context.Context, txhash *TransactionHash) (*Transaction, error) {
	fmt.Println("GetTransaction begin")
	if txhash == nil {
		return nil, fmt.Errorf("argument cannot be nil pointer")
	}
	txHash := txhash.Hash

	txDb := tx.TxDb
	if txDb == nil {
		panic(fmt.Errorf("TxDb should be nil"))
	}
	tx, err := txDb.(*tx.TxPoolDb).Get(txHash)
	if err != nil {
		return nil, err
	}

	return &Transaction{Tx: tx.Encode()}, nil
}

func (s *RpcServer) GetBalance(ctx context.Context, iak *Key) (*Value, error) {
	fmt.Println("GetBalance begin")
	if iak == nil {
		return nil, fmt.Errorf("argument cannot be nil pointer")
	}
	ia := iak.S
	val0, err := state.StdPool.GetHM("iost", state.Key(ia))
	if err != nil {
		return nil, err
	}
	val, ok := val0.(*state.VFloat)
	if !ok {
		return nil, fmt.Errorf("RPC : pool type error: should VFloat, acture %v; in iost.%v",
			reflect.TypeOf(val0).String(), vm.IOSTAccount(ia))
	}
	balance := val.EncodeString()

	return &Value{Sv: balance}, nil
}

func (s *RpcServer) GetState(ctx context.Context, stkey *Key) (*Value, error) {
	fmt.Println("GetState begin")
	if stkey == nil {
		return nil, fmt.Errorf("argument cannot be nil pointer")
	}
	key := stkey.S

	stPool := state.StdPool
	if stPool == nil {
		panic(fmt.Errorf("state.StdPool shouldn't be nil"))
	}
	stValue, err := stPool.Get(state.Key(key))
	if err != nil {
		return nil, fmt.Errorf("GetState Error: [%v]", err)
	}

	return &Value{Sv: stValue.EncodeString()}, nil
}

func (s *RpcServer) GetBlock(ctx context.Context, bk *BlockKey) (*BlockInfo, error) {
	if bk == nil {
		return nil, fmt.Errorf("argument cannot be nil pointer")
	}

	bc := block.BChain
	if bc == nil {
		panic(fmt.Errorf("block.BChain cannot be nil"))
	}
	layer := bk.Layer
	curLen := bc.Length()
	if (layer < 0) || (uint64(layer) > curLen-1) {
		return nil, fmt.Errorf("out of bound")
	}
	block := bc.GetBlockByNumber(curLen - 1 - uint64(layer))
	if block == nil {
		return nil, fmt.Errorf("cannot get BlockInfo")
	}

	head := &Head{
		Version:    block.Head.Version,
		ParentHash: block.Head.ParentHash,
		TreeHash:   block.Head.TreeHash,
		BlockHash:  block.HeadHash(),
		Info:       block.Head.Info,
		Number:     block.Head.Number,
		Witness:    block.Head.Witness,
		Signature:  block.Head.Signature,
		Time:       block.Head.Time,
	}

	txList := make([]*TransactionKey, block.LenTx())
	for k, v := range block.Content {
		txList[k] = &TransactionKey{
			Publisher: v.Publisher.Pubkey,
			Nonce:     v.Nonce,
		}
	}

	return &BlockInfo{
		Head:   head,
		Txcnt:  int64(block.LenTx()),
		TxList: txList,
	}, nil
}

func (s *RpcServer) GetBlockByHeight(ctx context.Context, bk *BlockKey) (*BlockInfo, error) {
	if bk == nil {
		return nil, fmt.Errorf("argument cannot be nil pointer")
	}

	bc := block.BChain
	if bc == nil {
		panic(fmt.Errorf("block.BChain cannot be nil"))
	}
	height := bk.Layer
	curLen := bc.Length()
	if (height < 0) || (uint64(height) > curLen-1) {
		return nil, fmt.Errorf("out of bound")
	}
	block := bc.GetBlockByNumber(uint64(height))
	if block == nil {
		return nil, fmt.Errorf("cannot get BlockInfo")
	}

	head := &Head{
		Version:    block.Head.Version,
		ParentHash: block.Head.ParentHash,
		TreeHash:   block.Head.TreeHash,
		BlockHash:  block.HeadHash(),
		Info:       block.Head.Info,
		Number:     block.Head.Number,
		Witness:    block.Head.Witness,
		Signature:  block.Head.Signature,
		Time:       block.Head.Time,
	}

	txList := make([]*TransactionKey, block.LenTx())
	for k, v := range block.Content {
		txList[k] = &TransactionKey{
			Publisher: v.Publisher.Pubkey,
			Nonce:     v.Nonce,
		}
	}

	return &BlockInfo{
		Head:   head,
		Txcnt:  int64(block.LenTx()),
		TxList: txList,
	}, nil
}
