package rpc

import (
	"context"
	"fmt"
	"net"
	"strconv"
	"strings"
	"time"

	"google.golang.org/grpc"

	"github.com/golang/protobuf/ptypes/empty"
	"github.com/iost-official/Go-IOS-Protocol/common"
	"github.com/iost-official/Go-IOS-Protocol/core/block"
	"github.com/iost-official/Go-IOS-Protocol/core/blockcache"
	"github.com/iost-official/Go-IOS-Protocol/core/event"
	"github.com/iost-official/Go-IOS-Protocol/core/global"
	"github.com/iost-official/Go-IOS-Protocol/core/tx"
	"github.com/iost-official/Go-IOS-Protocol/core/txpool"
	"github.com/iost-official/Go-IOS-Protocol/db"
	"github.com/iost-official/Go-IOS-Protocol/ilog"
	"github.com/iost-official/Go-IOS-Protocol/vm/database"
)

//go:generate mockgen -destination mock_rpc/mock_rpc.go -package rpc_mock github.com/iost-official/Go-IOS-Protocol/new_rpc ApisServer

// RPCServer is the class of RPC server
type RPCServer struct {
	bc      blockcache.BlockCache
	txdb    tx.TxDB
	txpool  txpool.TxPool
	bchain  block.Chain
	forkDB  db.MVCCDB
	visitor *database.Visitor
	port    int
}

// newRPCServer
func NewRPCServer(tp txpool.TxPool, bcache blockcache.BlockCache, _global global.BaseVariable) *RPCServer {
	forkDb := _global.StateDB().Fork()
	return &RPCServer{
		txdb:    _global.TxDB(),
		txpool:  tp,
		bchain:  _global.BlockChain(),
		bc:      bcache,
		forkDB:  forkDb,
		visitor: database.NewVisitor(0, forkDb),
		port:    _global.Config().RPC.GRPCPort,
	}
}

// Start ...
func (s *RPCServer) Start() error {
	port := strconv.Itoa(s.port)
	if !strings.HasPrefix(port, ":") {
		port = ":" + port
	}

	lis, err := net.Listen("tcp4", port)
	if err != nil {
		return fmt.Errorf("failed to listen: %v", err)
	}

	server := grpc.NewServer()
	if s == nil {
		return fmt.Errorf("failed to rpc NewServer")
	}

	RegisterApisServer(server, s)
	go server.Serve(lis)
	ilog.Info("RPCServer Start")
	return nil
}

// Stop ...
func (s *RPCServer) Stop() {
	return
}

// GetHeight ...
func (s *RPCServer) GetHeight(ctx context.Context, empty *empty.Empty) (*HeightRes, error) {
	return &HeightRes{
		Height: s.bchain.Length() - 1,
	}, nil
}

// GetTxByHash ...
func (s *RPCServer) GetTxByHash(ctx context.Context, hash *HashReq) (*TxRes, error) {
	if hash == nil {
		return nil, fmt.Errorf("argument cannot be nil pointer")
	}
	txHash := hash.Hash
	txHashBytes := common.Base58Decode(txHash)

	trx, err := s.txdb.GetTx(txHashBytes)

	if err != nil {
		return nil, err
	}

	return &TxRes{
		TxRaw: trx.ToTxRaw(),
		Hash:  trx.Hash(),
	}, nil
}

// GetTxReceiptByHash ...
func (s *RPCServer) GetTxReceiptByHash(ctx context.Context, hash *HashReq) (*TxReceiptRes, error) {
	if hash == nil {
		return nil, fmt.Errorf("argument cannot be nil pointer")
	}
	receiptHash := hash.Hash
	receiptHashBytes := common.Base58Decode(receiptHash)

	receipt, err := s.txdb.GetReceipt(receiptHashBytes)
	if err != nil {
		return nil, err
	}

	return &TxReceiptRes{
		TxReceiptRaw: receipt.ToTxReceiptRaw(),
		Hash:         receiptHashBytes,
	}, nil
}

// GetTxReceiptByTxHash...
func (s *RPCServer) GetTxReceiptByTxHash(ctx context.Context, hash *HashReq) (*TxReceiptRes, error) {
	if hash == nil {
		return nil, fmt.Errorf("argument cannot be nil pointer")
	}
	txHash := hash.Hash
	txHashBytes := common.Base58Decode(txHash)

	receipt, err := s.txdb.GetReceiptByTxHash(txHashBytes)
	if err != nil {
		return nil, err
	}

	return &TxReceiptRes{
		TxReceiptRaw: receipt.ToTxReceiptRaw(),
		Hash:         receipt.Hash(),
	}, nil
}

// GetBlockByHash ...
func (s *RPCServer) GetBlockByHash(ctx context.Context, blkHashReq *BlockByHashReq) (*BlockInfo, error) {
	if blkHashReq == nil {
		return nil, fmt.Errorf("argument cannot be nil pointer")
	}

	hash := blkHashReq.Hash
	hashBytes := common.Base58Decode(hash)
	complete := blkHashReq.Complete

	blk, _ := s.bchain.GetBlockByHash(hashBytes)
	if blk == nil {
		blk, _ = s.bc.GetBlockByHash(hashBytes)
	}
	if blk == nil {
		return nil, fmt.Errorf("cant find the block")
	}
	blkInfo := &BlockInfo{
		Head:   blk.Head,
		Hash:   blk.HeadHash(),
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
	for _, receipt := range blk.Receipts {
		if complete {
			blkInfo.Receipts = append(blkInfo.Receipts, receipt.ToTxReceiptRaw())
		} else {
			blkInfo.ReceiptHash = append(blkInfo.ReceiptHash, receipt.Hash())
		}
	}
	return blkInfo, nil
}

// GetBlockByNum ...
func (s *RPCServer) GetBlockByNum(ctx context.Context, blkNumReq *BlockByNumReq) (*BlockInfo, error) {
	if blkNumReq == nil {
		return nil, fmt.Errorf("argument cannot be nil pointer")
	}

	num := blkNumReq.Num
	complete := blkNumReq.Complete

	blk, _ := s.bchain.GetBlockByNumber(num)
	if blk == nil {
		blk, _ = s.bc.GetBlockByNumber(num)
	}
	if blk == nil {
		return nil, fmt.Errorf("cant find the block")
	}
	blkInfo := &BlockInfo{
		Head:   blk.Head,
		Hash:   blk.HeadHash(),
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

// GetState ...
func (s *RPCServer) GetState(ctx context.Context, key *GetStateReq) (*GetStateRes, error) {
	if key == nil {
		return nil, fmt.Errorf("argument cannot be nil pointer")
	}
	s.forkDB.Checkout(string(s.bc.LinkedRoot().Block.HeadHash()))
	return &GetStateRes{
		Value: s.visitor.BasicHandler.Get(key.Key),
	}, nil
}

// GetBalance ...
func (s *RPCServer) GetBalance(ctx context.Context, key *GetBalanceReq) (*GetBalanceRes, error) {
	if key == nil {
		return nil, fmt.Errorf("argument cannot be nil pointer")
	}
	if key.UseLongestChain {
		s.forkDB.Checkout(string(s.bc.Head().Block.HeadHash())) // long
	} else {
		s.forkDB.Checkout(string(s.bc.LinkedRoot().Block.HeadHash())) // confirm
	}
	return &GetBalanceRes{
		Balance: s.visitor.Balance(key.ID),
	}, nil
}

// SendRawTx ...
func (s *RPCServer) SendRawTx(ctx context.Context, rawTx *RawTxReq) (*SendRawTxRes, error) {
	ilog.Info("RPC received rawTx")
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
	ilog.Infof("the Tx is:\n%+v\n", trx)
	ret := s.txpool.AddTx(&trx)
	switch ret {
	case txpool.TimeError:
		return nil, fmt.Errorf("tx err:%v", "TimeError")
	case txpool.VerifyError:
		return nil, fmt.Errorf("tx err:%v", "VerifyError")
	case txpool.DupError:
		return nil, fmt.Errorf("tx err:%v", "DupError")
	case txpool.GasPriceError:
		return nil, fmt.Errorf("tx err:%v", "GasPriceError")
	default:
	}
	res := SendRawTxRes{}
	res.Hash = string(trx.Hash())
	return &res, nil
}

// EstimateGas ...
func (s *RPCServer) EstimateGas(ctx context.Context, rawTx *RawTxReq) (*GasRes, error) {
	return nil, nil
}

// Subscribe ...
func (s *RPCServer) Subscribe(req *SubscribeReq, res Apis_SubscribeServer) error {
	ec := event.GetEventCollectorInstance()
	sub := event.NewSubscription(100, req.Topics)
	ec.Subscribe(sub)
	defer ec.Unsubscribe(sub)

	timerChan := time.NewTicker(time.Minute).C
forloop:
	for {
		select {
		case <-timerChan:
			ilog.Debugf("timeup in subscribe send")
			break forloop
		case ev := <-sub.ReadChan():
			err := res.Send(&SubscribeRes{Ev: ev})
			if err != nil {
				return err
			}
		default:
		}
	}
	return nil
}
