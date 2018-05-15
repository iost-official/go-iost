package rpc

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/iost-official/prototype/core/block"
	"github.com/iost-official/prototype/core/message"
	"github.com/iost-official/prototype/core/tx"
	"github.com/iost-official/prototype/network"
)

type BInfo struct {
	Head  block.BlockHead
	txCnt int
}
type HttpServer struct {
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
	router, _ := network.RouterFactory("base")

	broadTx := message.Message{
		Body:    tx1.Encode(),
		ReqType: int32(network.ReqPublishTx),
	}
	router.Broadcast(broadTx)
	//add this tx to txpool
	tp, _ := tx.TxPoolFactory("mem") //TODO:in fact,we should find the txpool_mem,not create a new txpool_mem
	tp.Add(&tx1)
	return &Response{Code: 0}, nil
}

func (s *HttpServer) GetContract(ctx context.Context, tx *ContractKey) (*Contract, error) {

	return nil, nil
}

func (s *HttpServer) GetBalance(ctx context.Context, tx *Key) (*Value, error) {

	return nil, nil
}

func (s *HttpServer) GetState(ctx context.Context, tx *Key) (*Value, error) {

	return nil, nil
}

func (s *HttpServer) GetBlock(ctx context.Context, bk *BlockKey) (*BlockInfo, error) {
	bc, err := block.NewBlockChain() //we should get the instance of Chain,not to Create it again in the real version
	if err != nil {
		return nil, err
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
