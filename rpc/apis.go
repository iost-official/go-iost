package rpc

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/bitly/go-simplejson"
	"github.com/golang/protobuf/ptypes/empty"
	"google.golang.org/grpc"

	"github.com/iost-official/go-iost/common"
	"github.com/iost-official/go-iost/consensus/cverifier"
	"github.com/iost-official/go-iost/consensus/pob"
	"github.com/iost-official/go-iost/core/block"
	"github.com/iost-official/go-iost/core/blockcache"
	"github.com/iost-official/go-iost/core/contract"
	"github.com/iost-official/go-iost/core/event"
	"github.com/iost-official/go-iost/core/global"
	"github.com/iost-official/go-iost/core/tx"
	"github.com/iost-official/go-iost/core/tx/pb"
	"github.com/iost-official/go-iost/core/txpool"
	"github.com/iost-official/go-iost/db"
	"github.com/iost-official/go-iost/ilog"
	"github.com/iost-official/go-iost/p2p"
	"github.com/iost-official/go-iost/verifier"
	"github.com/iost-official/go-iost/vm/database"
	"github.com/iost-official/go-iost/vm/host"
)

//go:generate mockgen -destination mock_rpc/mock_rpc.go -package rpc_mock github.com/iost-official/go-iost/new_rpc ApisServer

// GRPCServer GRPC rpc server
type GRPCServer struct {
	bc         blockcache.BlockCache
	p2pService p2p.Service
	txpool     txpool.TxPool
	bchain     block.Chain
	forkDB     db.MVCCDB
	visitor    *database.Visitor
	port       int
	bv         global.BaseVariable
}

// NewRPCServer create GRPC rpc server
func NewRPCServer(tp txpool.TxPool, bcache blockcache.BlockCache, _global global.BaseVariable, p2pService p2p.Service) *GRPCServer {
	forkDb := _global.StateDB().Fork()
	return &GRPCServer{
		p2pService: p2pService,
		txpool:     tp,
		bchain:     _global.BlockChain(),
		bc:         bcache,
		forkDB:     forkDb,
		visitor:    database.NewVisitor(0, forkDb),
		port:       _global.Config().RPC.GRPCPort,
		bv:         _global,
	}
}

// Start start GRPC server
func (s *GRPCServer) Start() error {
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

// Stop stop GRPC server
func (s *GRPCServer) Stop() {
	return
}

// GetNodeInfo return the node info
func (s *GRPCServer) GetNodeInfo(ctx context.Context, empty *empty.Empty) (*NodeInfoRes, error) {
	netService, ok := s.p2pService.(*p2p.NetService)
	if !ok {
		return nil, fmt.Errorf("internal error: netService type conversion failed")
	}
	res := &NodeInfoRes{}
	res.Network = &NetworkInfo{}
	res.Network.PeerInfo = make([]*PeerInfo, 0)
	for _, p := range netService.GetAllNeighbors() {
		res.Network.PeerInfo = append(res.Network.PeerInfo, &PeerInfo{ID: p.ID(), Addr: p.Addr()})
	}
	res.Network.PeerCount = (int32)(len(res.Network.PeerInfo))
	res.Network.ID = s.p2pService.ID()
	res.GitHash = global.GitHash
	res.BuildTime = global.BuildTime
	res.Mode = s.bv.Mode().String()
	return res, nil
}

// GetChainInfo return the chain info
func (s *GRPCServer) GetChainInfo(ctx context.Context, empty *empty.Empty) (*ChainInfoRes, error) {
	return &ChainInfoRes{
		NetType:              s.bv.Config().Version.NetName,
		ProtocolVersion:      s.bv.Config().Version.ProtocolVersion,
		Height:               s.bchain.Length() - 1,
		WitnessList:          pob.GetStaticProperty().WitnessList,
		HeadBlock:            toBlockInfo(s.bc.Head().Block, false),
		LatestConfirmedBlock: toBlockInfo(s.bc.LinkedRoot().Block, false),
	}, nil
}

// GetTxByHash get tx by transaction hash
func (s *GRPCServer) GetTxByHash(ctx context.Context, hash *HashReq) (*TxRes, error) {
	if hash == nil {
		return nil, fmt.Errorf("argument cannot be nil pointer")
	}
	txHash := hash.Hash
	txHashBytes := common.Base58Decode(txHash)

	trx, err := s.bchain.GetTx(txHashBytes)

	if err != nil {
		return nil, err
	}

	return &TxRes{
		Tx:   trx.ToPb(),
		Hash: common.Base58Encode(trx.Hash()),
	}, nil
}

// GetTxReceiptByHash get receipt by receipt hash
func (s *GRPCServer) GetTxReceiptByHash(ctx context.Context, hash *HashReq) (*TxReceiptRes, error) {
	if hash == nil {
		return nil, fmt.Errorf("argument cannot be nil pointer")
	}
	receiptHash := hash.Hash
	receiptHashBytes := common.Base58Decode(receiptHash)

	receipt, err := s.bchain.GetReceipt(receiptHashBytes)
	if err != nil {
		return nil, err
	}

	return &TxReceiptRes{
		TxReceipt: receipt.ToPb(),
		Hash:      common.Base58Encode(receiptHashBytes),
	}, nil
}

// GetTxReceiptByTxHash get receipt by transaction hash
func (s *GRPCServer) GetTxReceiptByTxHash(ctx context.Context, hash *HashReq) (*TxReceiptRes, error) {
	if hash == nil {
		return nil, fmt.Errorf("argument cannot be nil pointer")
	}
	txHash := hash.Hash
	txHashBytes := common.Base58Decode(txHash)

	receipt, err := s.bchain.GetReceiptByTxHash(txHashBytes)
	if err != nil {
		return nil, err
	}

	return &TxReceiptRes{
		TxReceipt: receipt.ToPb(),
		Hash:      common.Base58Encode(receipt.Hash()),
	}, nil
}

func toBlockInfo(blk *block.Block, complete bool) *BlockInfo {
	blkInfo := &BlockInfo{
		Head:   blk.Head.ToPb(),
		Hash:   common.Base58Encode(blk.HeadHash()),
		Txs:    make([]*txpb.Tx, 0),
		Txhash: make([]string, 0),
	}
	for _, trx := range blk.Txs {
		if complete {
			blkInfo.Txs = append(blkInfo.Txs, trx.ToPb())
		}
		blkInfo.Txhash = append(blkInfo.Txhash, common.Base58Encode(trx.Hash()))
	}
	for _, receipt := range blk.Receipts {
		if complete {
			blkInfo.Receipts = append(blkInfo.Receipts, receipt.ToPb())
		}
		blkInfo.ReceiptHash = append(blkInfo.ReceiptHash, common.Base58Encode(receipt.Hash()))
	}
	return blkInfo
}

// GetBlockByHash get block by block head hash
func (s *GRPCServer) GetBlockByHash(ctx context.Context, blkHashReq *BlockByHashReq) (*BlockInfo, error) {
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
	blkInfo := toBlockInfo(blk, complete)
	return blkInfo, nil
}

// GetBlockByNum get block by block number
func (s *GRPCServer) GetBlockByNum(ctx context.Context, blkNumReq *BlockByNumReq) (*BlockInfo, error) {
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
	blkInfo := toBlockInfo(blk, complete)
	return blkInfo, nil
}

// GetContractStorage get contract storage from state db
func (s *GRPCServer) GetContractStorage(ctx context.Context, req *GetContractStorageReq) (*GetContractStorageRes, error) {
	if req == nil {
		return nil, fmt.Errorf("argument cannot be nil pointer")
	}
	if req.ContractID == "" {
		return nil, fmt.Errorf("contract id cannot be empty")
	}
	s.forkDB.Checkout(string(s.bc.LinkedRoot().Block.HeadHash()))
	var value string

	k := req.ContractID
	if req.Owner != "" {
		k = k + "@" + req.Owner
	}
	k = k + database.Separator + req.Key
	if req.Field == "" {
		value = s.visitor.BasicHandler.Get(k)
	} else {
		value = s.visitor.MapHandler.MGet(k, req.Field)
	}
	result := database.Unmarshal(value)
	var data string
	if result == nil || reflect.TypeOf(result).Kind() != reflect.String {
		bytes, err := json.Marshal(result)
		if err != nil {
			return nil, fmt.Errorf("cannot unmarshal %v", value)
		}
		data = string(bytes)
	} else {
		data = result.(string)
	}
	return &GetContractStorageRes{JsonStr: data}, nil
}

// GetContract return a contract by contract id
func (s *GRPCServer) GetContract(ctx context.Context, key *GetContractReq) (*GetContractRes, error) {
	if key == nil {
		return nil, fmt.Errorf("argument cannot be nil pointer")
	}
	if key.ContractID == "" {
		return nil, fmt.Errorf("argument cannot be empty string")
	}
	if !strings.HasPrefix(key.ContractID, "Contract") {
		return nil, fmt.Errorf("contract id should start with \"Contract\"")
	}

	txHashBytes := common.Base58Decode(key.ContractID[len("Contract"):])
	trx, err := s.bchain.GetTx(txHashBytes)

	if err != nil {
		return nil, err
	}
	// assume only one 'SetCode' action
	txActionName := trx.Actions[0].ActionName
	if trx.Actions[0].Contract != "system.iost" || txActionName != "SetCode" && txActionName != "UpdateCode" {
		return nil, fmt.Errorf("not a SetCode or Update transaction")
	}
	js, err := simplejson.NewJson([]byte(trx.Actions[0].Data))
	if err != nil {
		return nil, err
	}
	contractStr, err := js.GetIndex(0).String()
	if err != nil {
		return nil, err
	}
	c := &contract.Contract{}
	err = c.B64Decode(contractStr)
	if err != nil {
		return nil, err
	}
	return &GetContractRes{Value: c}, nil
	//return &GetContractRes{Value: s.visitor.Contract(key.Key)}, nil

}

// GetAccountInfo get account balance and gas etc
func (s *GRPCServer) GetAccountInfo(ctx context.Context, key *GetAccountReq) (*GetAccountRes, error) {
	if key == nil {
		return nil, fmt.Errorf("argument cannot be nil pointer")
	}
	if key.UseLongestChain {
		s.forkDB.Checkout(string(s.bc.Head().Block.HeadHash())) // long
	} else {
		s.forkDB.Checkout(string(s.bc.LinkedRoot().Block.HeadHash())) // confirm
	}

	accStr := database.MustUnmarshal(s.visitor.MGet("auth.iost-auth", key.ID))
	if accStr == nil {
		return nil, fmt.Errorf("non exist user %v", key.ID)
	}

	ram := &RAMInfo{}
	ram.Available = s.visitor.TokenBalance("ram", key.ID)
	balance := s.visitor.TokenBalanceFixed("iost", key.ID).ToString()

	gas := &GASInfo{}
	var h *host.Host
	c := host.NewContext(nil)
	h = host.NewHost(c, s.visitor, nil, nil)
	h.Context().Set("contract_name", "gas.iost")
	g := host.NewGasManager(h)
	v, _ := g.CurrentTotalGas(key.ID, s.bc.LinkedRoot().Head.Time)
	gas.CurrentTotal = v.ToString()
	v, _ = g.GasRate(key.ID)
	gas.IncreaseSpeed = v.ToString()
	v, _ = g.GasLimit(key.ID)
	gas.Limit = v.ToString()
	v, _ = g.GasPledge(key.ID, key.ID)
	gas.PledgedCoin = v.ToString()
	return &GetAccountRes{
		Balance:     balance,
		Gas:         gas,
		Ram:         ram,
		AccountJson: accStr.(string),
	}, nil
}

// SendTx send transaction to blockchain
func (s *GRPCServer) SendTx(ctx context.Context, txReq *TxReq) (*SendTxRes, error) {
	if txReq == nil {
		return nil, fmt.Errorf("argument cannot be nil pointer")
	}
	var trx tx.Tx
	trx.FromPb(txReq.Tx)
	// check trx is valid
	// check time
	if !trx.IsDefer() {
		now := time.Now().UnixNano()
		if trx.Time > now {
			return nil, fmt.Errorf("tx time is in future, %v > %v", trx.Time, now)
		}
		if trx.Expiration <= now {
			return nil, fmt.Errorf("tx already expired , %v <= %v", trx.Expiration, now)
		}
		if now-trx.Time > tx.MaxExpiration {
			return nil, fmt.Errorf("received tx too late, exceed max expiration time , %v - %v > %v", now, trx.Time, tx.MaxExpiration)
		}
	}
	// check gas
	if trx.GasPrice < 100 || trx.GasPrice > 10000 {
		return nil, errors.New("gas price illegal, should in [100, 10000]")
	}
	if trx.GasLimit < 500 {
		return nil, errors.New("gas limit illegal, should >= 500")
	}
	var h *host.Host
	c := host.NewContext(nil)
	h = host.NewHost(c, s.visitor, nil, nil)
	g := host.NewGasManager(h)
	gas, _ := g.CurrentTotalGas(trx.Publisher, s.bc.LinkedRoot().Head.Time)
	price := &common.Fixed{Value: trx.GasPrice, Decimal: 2}
	if gas.LessThan(price.Times(trx.GasLimit)) {
		return nil, fmt.Errorf("%v gas less than price * limit %v < %v * %v", trx.Publisher, gas.ToString(), price.ToString(), trx.GasLimit)
	}

	err := s.txpool.AddTx(&trx)
	if err != nil {
		return nil, err
	}
	res := SendTxRes{}
	res.Hash = common.Base58Encode(trx.Hash())
	return &res, nil
}

// ExecTx only exec the tx, but not put it onto chain
func (s *GRPCServer) ExecTx(ctx context.Context, txReq *TxReq) (*ExecTxRes, error) {
	if txReq == nil {
		return nil, fmt.Errorf("argument cannot be nil pointer")
	}
	var trx tx.Tx
	trx.FromPb(txReq.Tx)
	_, head := s.txpool.PendingTx()
	topBlock := head.Block
	blk := block.Block{
		Head: &block.BlockHead{
			Version:    0,
			ParentHash: topBlock.HeadHash(),
			Number:     topBlock.Head.Number + 1,
			Witness:    "",
			Time:       time.Now().UnixNano(),
		},
		Txs:      []*tx.Tx{},
		Receipts: []*tx.TxReceipt{},
	}
	v := verifier.Verifier{}
	receipt, err := v.Try(blk.Head, s.forkDB, &trx, cverifier.TxExecTimeLimit)
	if err != nil {
		return nil, fmt.Errorf("exec tx failed: %v", err)
	}
	return &ExecTxRes{TxReceipt: receipt.ToPb()}, nil
}

// Subscribe used for event
func (s *GRPCServer) Subscribe(req *SubscribeReq, res Apis_SubscribeServer) error {
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
