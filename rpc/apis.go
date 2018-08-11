package rpc

import (
	"context"
	"fmt"

	"github.com/iost-official/Go-IOS-Protocol/common"
	"github.com/iost-official/Go-IOS-Protocol/core/new_block"
	"github.com/iost-official/Go-IOS-Protocol/core/new_blockcache"
	"github.com/iost-official/Go-IOS-Protocol/core/new_tx"
	"github.com/iost-official/Go-IOS-Protocol/core/new_txpool"
)

//go:generate mockgen -destination mock_rpc/mock_rpc.go -package rpc_mock github.com/iost-official/Go-IOS-Protocol/rpc CliServer

type RpcServer struct {
}

// newRpcServer
func newRpcServer() *RpcServer {
	s := &RpcServer{}
	return s
}

func (s *RpcServer) GetHeight(ctx context.Context, void *VoidReq) (*HeightRes, error) {
	return &HeightRes{
		Height: block.BChain.Length(),
	}
}

func GetTxByHash(ctx context.Context, hash *HashReq) (*TxRaw, error) {
	if hash == nil {
		return nil, fmt.Errorf("argument cannot be nil pointer")
	}
	txHash := hash.Hash

	trx, err := txdb.Get(txHash)
	if err != nil {
		return nil, err
	}
	txRaw := trx.toTxRaw()
	return txRaw, nil
}

func GetBlockByHash(ctx context.Context, blkHashReq *BlockByHashReq) (*BlockInfo, error) {
	if blkHashReq == nil {
		return nil, fmt.Errorf("argument cannot be nil pointer")
	}

	hash := blkHashReq.hash
	complete := blkHashReq.complete

	blk := bchain.GetBlockByHash(hash)
	if blk == nil {
		blk := bc.GetBlockByHash(hash)
	}
	if blk == nil {
		return nil, fmt.Errorf("cant find the block")
	}
	blkInfo := &BlockInfo{
		Head:   blk.Head,
		Txs:    make([]*new_tx.TxRaw, 0),
		Txhash: make([][]byte, 0),
	}
	for _, trx := range blk.Txs {
		if complete {
			blkInfo.Txs = append(blkInfo.Txs, trx.ToTxRaw)
		} else {
			blkInfo.Txhash = append(blkInfo.Txhash, trx.Hash())
		}
	}
	return blkInfo, nil
}

func GetBlockByNum(ctx context.Context, blkNumReq *BlockByNumReq) (*BlockInfo, error) {
	if blkNumReq == nil {
		return nil, fmt.Errorf("argument cannot be nil pointer")
	}

	num := blkNumReq.Num
	complete := blkNumReq.complete
	if bchain == nil {
		panic(fmt.Errorf("block.BChain cannot be nil"))
	}
	blk := bchain.GetBlockByNumber(num)
	if blk == nil {
		blk := bc.GetBlockByNumber(num)
	}
	if blk == nil {
		return nil, fmt.Errorf("cant find the block")
	}
	blkInfo := &BlockInfo{
		Head:   blk.Head,
		Txs:    make([]*new_tx.TxRaw, 0),
		Txhash: make([][]byte, 0),
	}
	for _, trx := range blk.Txs {
		if complete {
			blkInfo.Txs = append(blkInfo.Txs, trx.ToTxRaw)
		} else {
			blkInfo.Txhash = append(blkInfo.Txhash, trx.Hash())
		}
	}
	return blkInfo, nil
}
func GetBalance(ctx context.Context, key *GetBalanceReq) (*GetBalanceRes, error) {
	if key == nil {
		return nil, fmt.Errorf("argument cannot be nil pointer")
	}
	pub := key.pubkey
	balance, err := mvccdb.Get("", "i-"+pub)
	if err != nil {
		return nil, err
	}
	return &GetBalanceRes{
		balance: balance,
	}, nil
}

/*
func GetState(ctx context.Context,key *GetStateReq) (*GetStateRes, error) {
	if key == nil {
		return nil, fmt.Errorf("argument cannot be nil pointer")
	}
	pub := key.pubkey
	val,err:=mvccdb.Get("","i-"+pub)
	if err!=nil{
		return nil,err
	}
	return &GetStateRes{
		value:val,
	},nil
}
*/
func SendRawTx(ctx context.Context, rawTx *RawTxReq) (*SendRawTxRes, error) {
	if rawTx == nil {
		return nil, fmt.Errorf("argument cannot be nil pointer")
	}
	var trx tx.Tx
	err := trx.Decode(rawTx.data)
	if err != nil {
		return nil, err
	}
	// add servi
	tx.RecordTx(trx, tx.Data.Self())

	ret := new_txpool.TxPoolS.AddTx(trx)
	if ret != new_txpool.Success {
		return nil, fmt.Errorf("tx err:%v", ret)
	}
	res := SendRawTxRes{}
	res.hash = trx.Hash()
	return &res, nil
}

/*
func EstimateGas(ctx context.Context,rawTx *RawTxReq) (*GasRes, error){

}
*/
