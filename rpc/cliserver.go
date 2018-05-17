package rpc

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/iost-official/prototype/common"
	"github.com/iost-official/prototype/core/block"
	"github.com/iost-official/prototype/core/message"
	"github.com/iost-official/prototype/core/state"
	"github.com/iost-official/prototype/core/tx"
	"github.com/iost-official/prototype/network"
	"github.com/iost-official/prototype/vm"
)

//go:generate mockgen -destination mock_rpc/mock_rpc.go -package rpc_mock github.com/iost-official/prototype/rpc CliServer

type BInfo struct {
	Head  block.BlockHead
	txCnt int
}
type HttpServer struct {
}

// newHttpServer 初始Http RPC结构体
func newHttpServer() *HttpServer{
	s:= &HttpServer{}
	return s
}

func (s *HttpServer) PublishTx(ctx context.Context, _tx *Transaction) (*Response, error) {

	var tx1 tx.Tx
	if _tx == nil {
		return &Response{Code: -1}, fmt.Errorf("argument cannot be nil pointer")
	}
	err := tx1.Decode(_tx.Tx)
	if err != nil {
		return &Response{Code: -1}, err
	}

	err = tx1.VerifySelf() //verify Publisher and Signers
	if err != nil {
		return &Response{Code: -1}, err
	}
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

	//add this tx to txpool
	tp, err := tx.TxPoolFactory("mem") //TODO:in fact,we should find the txpool_mem,not create a new txpool_mem
	if err != nil {
		panic(err)
	}
	tp.Add(&tx1)
	return &Response{Code: 0}, nil
}

func (s *HttpServer) GetTransaction(ctx context.Context, txkey *TransactionKey) (*Transaction, error) {

	if txkey == nil {
		return nil, fmt.Errorf("argument cannot be nil pointer")
	}
	var Pub common.Signature
	Pub.Decode([]byte(txkey.Publisher))
	Nonce := txkey.Nonce
	//check Publisher and Nonce in txkey

	txDb, err := tx.NewTxPoolDb() //TODO:in fact,we should find the txpool_db,not create a new txpool_db
	tx, err := txDb.(*tx.TxPoolDb).GetByPN(Nonce, Pub)
	if err != nil {
		return nil, err
	}

	return &Transaction{Tx: tx.Encode()}, nil
}

func (s *HttpServer) GetBalance(ctx context.Context, iak *Key) (*Value, error) {
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
		return nil, fmt.Errorf("pool type error: should VFloat, acture %v; in iost.%v",
			reflect.TypeOf(val0).String(), vm.IOSTAccount(ia))
	}
	balance := val.EncodeString()

	return &Value{Sv: balance}, nil
}

func (s *HttpServer) GetState(ctx context.Context, stkey *Key) (*Value, error) {
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

func (s *HttpServer) GetBlock(ctx context.Context, bk *BlockKey) (*BlockInfo, error) {
	if bk == nil {
		return nil, fmt.Errorf("argument cannot be nil pointer")
	}

	bc := block.BChain //we should get the instance of Chain,not to Create it again in the real version
	if bc == nil {
		panic(fmt.Errorf("block.BChain cannot be nil"))
	}
	layer := bk.Layer //I think bk.Layer should be uint64,because bc.Length() is uint64
	curLen := bc.Length()
	if (layer < 0) || (uint64(layer) > curLen-1) {
		return nil, fmt.Errorf("out of bound")
	}
	block := bc.GetBlockByNumber(curLen - 1 - uint64(layer))
	if block == nil {
		return nil, fmt.Errorf("cannot get BlockInfo")
	}
	//better to Encode BlockHead first?
	binfo := BInfo{
		Head:  block.Head,
		txCnt: block.LenTx(),
	}
	b, err := json.Marshal(binfo)
	if err != nil {
		return nil, fmt.Errorf("json.Marshal failed: [%v]", err)
	}
	return &BlockInfo{Json: string(b)}, nil
}
