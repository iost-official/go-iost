package rpc

import (
	"context"
	"error"
	"fmt"

	//	"github.com/iost-official/prototype/core/block"
	"github.com/iost-official/prototype/core/message"
	"github.com/iost-official/prototype/core/tx"
	"github.com/iost-official/prototype/network"
)

type HttpServer struct {
}

func (s *HttpServer) PublishTx(ctx context.Context, _tx *Transaction) (*Response, error) {
	/*
		var tx1 tx.Tx
		if _tx==nil{
			return &Response{Code: -1},fmt.Errorf("argument cannot be nil pointer")
		}
		err:=tx1.Decode(_tx.Tx)
		if err!=nil{
			return &Response{Code: -1},err
		}
		err = tx1.VerifySelf() //verify Publisher and Signers
		if err != nil {
			return &Response{Code: -1}, err
		}
		//broadcast the tx
		router, _ := network.RouterFactory("base")
		baseNet, _ := network.NewBaseNetwork(&(network.NetConifg{
			ListenAddr: "127.0.0.1"}))
		router.Init(baseNet, 12345)

		net:= router.(*RouterImpl).base.(*BaseNetwork)
		broadTx := message.Message{
			Body:    tx1.Encode(),
			ReqType: int32(network.ReqPublishTx),
			From:    net.localNode.String(), //?
		}
		router.Broadcast(broadTx)
		//add this tx to txpool
		tp, _ := tx.TxPoolFactory("mem")
		tp.Add(&tx1)
		return &Response{Code: 0}, nil
	*/
	return nil, nil
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
	/*
		bc,err:=NewBlockChain()
		if err!=nil{
			return nil,err
		}
		layer:=bk.Layer
		curLen=bc.Length()
		if (layer>curLen-1) || (layer<0){
			return nil,fmt.Errorf("out of bound")
		}
		bInfo:=bc.GetBlockByNumber(curLen-1-layer)
		if bInfo==nil{
			return nil,fmt.Errorf("cannot get BlockInfo")
		}
	*/
	return nil, nil
}
