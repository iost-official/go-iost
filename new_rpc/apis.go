package rpc

import (
	"context"
	"fmt"

	"github.com/iost-official/Go-IOS-Protocol/core/new_tx"
)

//go:generate mockgen -destination mock_rpc/mock_rpc.go -package rpc_mock github.com/iost-official/Go-IOS-Protocol/new_rpc ApisServer

type RpcServer struct {
}

// newRpcServer
func newRpcServer() *RpcServer {
	s := &RpcServer{}
	return s
}

func (s *RpcServer) GetHeight(ctx context.Context, void *VoidReq) (*HeightRes, error) {
	return &HeightRes{
		Height: bchain.Length(),
	}, nil
}

func (s *RpcServer) GetTxByHash(ctx context.Context, hash *HashReq) (*tx.TxRaw, error) {
	if hash == nil {
		return nil, fmt.Errorf("argument cannot be nil pointer")
	}
	txHash := hash.Hash

	trx, err := txdb.Get(txHash)
	if err != nil {
		return nil, err
	}
	txRaw := trx.ToTxRaw()
	return txRaw, nil
}

func (s *RpcServer) GetBlockByHash(ctx context.Context, blkHashReq *BlockByHashReq) (*BlockInfo, error) {
	if blkHashReq == nil {
		return nil, fmt.Errorf("argument cannot be nil pointer")
	}

	hash := blkHashReq.Hash
	complete := blkHashReq.Complete

	blk, _ := bchain.GetBlockByHash(hash)
	if blk == nil {
		blk, _ = bc.GetBlockByHash(hash)
	}
	if blk == nil {
		return nil, fmt.Errorf("cant find the block")
	}
	blkInfo := &BlockInfo{
		Head:   &blk.Head,
		Txs:    make([]*tx.TxRaw, 0),
		Txhash: make([][]byte, 0),
	}
	for _, trx := range blk.Txs {
		if complete {
			blkInfo.Txs = append(blkInfo.Txs, trx.ToTxRaw())
		} else {
			blkInfo.Txhash = append(blkInfo.Txhash, trx.Hash())
		}
	}
	return blkInfo, nil
}

func (s *RpcServer) GetBlockByNum(ctx context.Context, blkNumReq *BlockByNumReq) (*BlockInfo, error) {
	if blkNumReq == nil {
		return nil, fmt.Errorf("argument cannot be nil pointer")
	}

	num := blkNumReq.Num
	complete := blkNumReq.Complete

	blk, _ := bchain.GetBlockByNumber(num)
	if blk == nil {
		blk, _ = bc.GetBlockByNumber(num)
	}
	if blk == nil {
		return nil, fmt.Errorf("cant find the block")
	}
	blkInfo := &BlockInfo{
		Head:   &blk.Head,
		Txs:    make([]*tx.TxRaw, 0),
		Txhash: make([][]byte, 0),
	}
	for _, trx := range blk.Txs {
		if complete {
			blkInfo.Txs = append(blkInfo.Txs, trx.ToTxRaw())
		} else {
			blkInfo.Txhash = append(blkInfo.Txhash, trx.Hash())
		}
	}
	return blkInfo, nil
}

func (s *RpcServer) GetState(ctx context.Context, key *GetStateReq) (*GetStateRes, error) {
	if key == nil {
		return nil, fmt.Errorf("argument cannot be nil pointer")
	}
	return &GetStateRes{
		Value: visitor.BasicHandler.Get(key.Key),
	}, nil
}

func (s *RpcServer) GetBalance(ctx context.Context, key *GetBalanceReq) (*GetBalanceRes, error) {
	if key == nil {
		return nil, fmt.Errorf("argument cannot be nil pointer")
	}
	return &GetBalanceRes{
		Balance: visitor.BalanceHandler.Balance(key.Pubkey),
	}, nil 
}

func (s *RpcServer) SendRawTx(ctx context.Context, rawTx *RawTxReq) (*SendRawTxRes, error) {
	if rawTx == nil {
		return nil, fmt.Errorf("argument cannot be nil pointer")
	}
	var trx tx.Tx
	err := trx.Decode(rawTx.Data)
	if err != nil {
		return nil, err
	}
	// add servi
	//tx.RecordTx(trx, tx.Data.Self())
	/*
		ret := txpool.TxPoolS.AddTx(trx)
		if ret != txpool.Success {
			return nil, fmt.Errorf("tx err:%v", ret)
		}
	*/
	res := SendRawTxRes{}
	res.Hash = trx.Hash()
	return &res, nil
}

func (s *RpcServer) EstimateGas(ctx context.Context, rawTx *RawTxReq) (*GasRes, error) {
	return nil, nil
}
