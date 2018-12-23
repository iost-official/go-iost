package rpc

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"sort"
	"time"

	"github.com/iost-official/go-iost/common"
	"github.com/iost-official/go-iost/consensus/cverifier"
	"github.com/iost-official/go-iost/consensus/pob"
	"github.com/iost-official/go-iost/core/block"
	"github.com/iost-official/go-iost/core/blockcache"
	"github.com/iost-official/go-iost/core/event"
	"github.com/iost-official/go-iost/core/global"
	"github.com/iost-official/go-iost/core/tx"
	"github.com/iost-official/go-iost/core/txpool"
	"github.com/iost-official/go-iost/ilog"
	"github.com/iost-official/go-iost/p2p"
	"github.com/iost-official/go-iost/rpc/pb"
	"github.com/iost-official/go-iost/verifier"
	"github.com/iost-official/go-iost/vm/database"
	"github.com/iost-official/go-iost/vm/host"
)

//go:generate mockgen -destination mock_rpc/mock_api.go -package main github.com/iost-official/go-iost/rpc/pb ApiServiceServer

// APIService implements all rpc APIs.
type APIService struct {
	bc         blockcache.BlockCache
	p2pService p2p.Service
	txpool     txpool.TxPool
	blockchain block.Chain
	bv         global.BaseVariable
}

// NewAPIService returns a new APIService instance.
func NewAPIService(tp txpool.TxPool, bcache blockcache.BlockCache, bv global.BaseVariable, p2pService p2p.Service) *APIService {
	return &APIService{
		p2pService: p2pService,
		txpool:     tp,
		blockchain: bv.BlockChain(),
		bc:         bcache,
		bv:         bv,
	}
}

// GetNodeInfo returns information abount node.
func (as *APIService) GetNodeInfo(context.Context, *rpcpb.EmptyRequest) (*rpcpb.NodeInfoResponse, error) {
	res := &rpcpb.NodeInfoResponse{
		BuildTime: global.BuildTime,
		GitHash:   global.GitHash,
		Mode:      as.bv.Mode().String(),
		Network:   &rpcpb.NetworkInfo{},
	}
	p2pNeighbors := as.p2pService.GetAllNeighbors()
	networkInfo := &rpcpb.NetworkInfo{
		Id:        as.p2pService.ID(),
		PeerCount: int32(len(p2pNeighbors)),
	}
	for _, p := range p2pNeighbors {
		networkInfo.PeerInfo = append(networkInfo.PeerInfo, &rpcpb.PeerInfo{
			Id:   p.ID(),
			Addr: p.Addr(),
		})
	}
	res.Network = networkInfo
	return res, nil
}

// GetRAMInfo returns the chain info.
func (as *APIService) GetRAMInfo(context.Context, *rpcpb.EmptyRequest) (*rpcpb.RAMInfoResponse, error) {
	dbVisitor := as.getStateDBVisitor(true)
	return &rpcpb.RAMInfoResponse{
		AvailableRam: dbVisitor.LeftRAM(),
		UsedRam:      dbVisitor.UsedRAM(),
		TotalRam:     dbVisitor.TotalRAM(),
		SellPrice:    dbVisitor.SellPrice(),
		BuyPrice:     dbVisitor.BuyPrice(),
	}, nil
}

// GetChainInfo returns the chain info.
func (as *APIService) GetChainInfo(context.Context, *rpcpb.EmptyRequest) (*rpcpb.ChainInfoResponse, error) {
	headBlock := as.bc.Head().Block
	libBlock := as.bc.LinkedRoot().Block
	netName := "unknown"
	version := "unknown"
	if as.bv.Config().Version != nil {
		netName = as.bv.Config().Version.NetName
		version = as.bv.Config().Version.ProtocolVersion
	}
	return &rpcpb.ChainInfoResponse{
		NetName:         netName,
		ProtocolVersion: version,
		WitnessList:     pob.GetStaticProperty().WitnessList,
		HeadBlock:       headBlock.Head.Number,
		HeadBlockHash:   common.Base58Encode(headBlock.HeadHash()),
		LibBlock:        libBlock.Head.Number,
		LibBlockHash:    common.Base58Encode(libBlock.HeadHash()),
	}, nil
}

// GetTxByHash returns the transaction corresponding to the given hash.
func (as *APIService) GetTxByHash(ctx context.Context, req *rpcpb.TxHashRequest) (*rpcpb.TransactionResponse, error) {
	txHashBytes := common.Base58Decode(req.GetHash())
	status := rpcpb.TransactionResponse_PENDING
	var (
		t         *tx.Tx
		txReceipt *tx.TxReceipt
		err       error
	)
	t, err = as.txpool.GetFromPending(txHashBytes)
	if err != nil {
		status = rpcpb.TransactionResponse_PACKED
		t, txReceipt, err = as.txpool.GetFromChain(txHashBytes)
		if err != nil {
			status = rpcpb.TransactionResponse_IRREVERSIBLE
			t, err = as.blockchain.GetTx(txHashBytes)
			if err != nil {
				return nil, errors.New("tx not found")
			}
		}
	}

	return &rpcpb.TransactionResponse{
		Status:      status,
		Transaction: toPbTx(t, txReceipt),
	}, nil
}

// GetTxReceiptByTxHash returns transaction receipts corresponding to the given tx hash.
func (as *APIService) GetTxReceiptByTxHash(ctx context.Context, req *rpcpb.TxHashRequest) (*rpcpb.TxReceipt, error) {
	txHashBytes := common.Base58Decode(req.GetHash())
	receipt, err := as.blockchain.GetReceiptByTxHash(txHashBytes)
	if err != nil {
		return nil, err
	}
	return toPbTxReceipt(receipt), nil
}

// GetBlockByHash returns block corresponding to the given hash.
func (as *APIService) GetBlockByHash(ctx context.Context, req *rpcpb.GetBlockByHashRequest) (*rpcpb.BlockResponse, error) {
	hashBytes := common.Base58Decode(req.GetHash())
	var (
		blk *block.Block
		err error
	)
	status := rpcpb.BlockResponse_PENDING
	blk, err = as.bc.GetBlockByHash(hashBytes)
	if err != nil {
		status = rpcpb.BlockResponse_IRREVERSIBLE
		blk, err = as.blockchain.GetBlockByHash(hashBytes)
		if err != nil {
			return nil, err
		}
	}
	return &rpcpb.BlockResponse{
		Status: status,
		Block:  toPbBlock(blk, req.GetComplete()),
	}, nil
}

// GetBlockByNumber returns block corresponding to the given number.
func (as *APIService) GetBlockByNumber(ctx context.Context, req *rpcpb.GetBlockByNumberRequest) (*rpcpb.BlockResponse, error) {
	number := req.GetNumber()
	var (
		blk *block.Block
		err error
	)
	status := rpcpb.BlockResponse_PENDING
	blk, err = as.bc.GetBlockByNumber(number)
	if err != nil {
		status = rpcpb.BlockResponse_IRREVERSIBLE
		blk, err = as.blockchain.GetBlockByNumber(number)
		if err != nil {
			return nil, err
		}
	}
	return &rpcpb.BlockResponse{
		Status: status,
		Block:  toPbBlock(blk, req.GetComplete()),
	}, nil
}

// GetAccount returns account information corresponding to the given account name.
func (as *APIService) GetAccount(ctx context.Context, req *rpcpb.GetAccountRequest) (*rpcpb.Account, error) {
	dbVisitor := as.getStateDBVisitor(req.ByLongestChain)
	// pack basic account information
	acc, _ := host.ReadAuth(dbVisitor, req.GetName())
	if acc == nil {
		return nil, errors.New("account not found")
	}
	ret := toPbAccount(acc)

	// pack balance and ram information
	balance := dbVisitor.TokenBalanceFixed("iost", req.GetName()).ToFloat()
	ret.Balance = balance
	ret.RamInfo = &rpcpb.Account_RAMInfo{
		Available: dbVisitor.TokenBalance("ram", req.GetName()),
	}

	// pack gas information
	var blkTime int64
	if req.GetByLongestChain() {
		blkTime = as.bc.Head().Head.Time
	} else {
		blkTime = as.bc.LinkedRoot().Head.Time
	}
	pGas := dbVisitor.PGasAtTime(req.GetName(), blkTime)
	tGas := dbVisitor.TGas(req.GetName())
	totalGas := pGas.Add(tGas)
	gasLimit := dbVisitor.GasLimit(req.GetName())
	gasRate := dbVisitor.GasRate(req.GetName())
	pledgedInfo := dbVisitor.PledgerInfo(req.GetName())
	ret.GasInfo = &rpcpb.Account_GasInfo{
		CurrentTotal:    totalGas.ToFloat(),
		PledgeGas:       pGas.ToFloat(),
		TransferableGas: tGas.ToFloat(),
		Limit:           gasLimit.ToFloat(),
		IncreaseSpeed:   gasRate.ToFloat(),
	}
	for _, p := range pledgedInfo {
		ret.GasInfo.PledgedInfo = append(ret.GasInfo.PledgedInfo, &rpcpb.Account_PledgeInfo{
			Amount:  p.Amount.ToFloat(),
			Pledger: p.Pledger,
		})
	}

	// pack frozen balance information
	frozen := dbVisitor.AllFreezedTokenBalanceFixed("iost", req.GetName())
	for _, f := range frozen {
		ret.FrozenBalances = append(ret.FrozenBalances, &rpcpb.FrozenBalance{
			Amount: f.Amount.ToFloat(),
			Time:   f.Ftime,
		})
	}

	return ret, nil
}

// GetTokenBalance returns contract information corresponding to the given contract ID.
func (as *APIService) GetTokenBalance(ctx context.Context, req *rpcpb.GetTokenBalanceRequest) (*rpcpb.GetTokenBalanceResponse, error) {
	dbVisitor := as.getStateDBVisitor(req.ByLongestChain)
	// pack basic account information
	acc, _ := host.ReadAuth(dbVisitor, req.GetAccount())
	if acc == nil {
		return nil, errors.New("account not found")
	}
	balance := dbVisitor.TokenBalanceFixed(req.GetToken(), req.GetAccount()).ToFloat()
	// pack frozen balance information
	frozen := dbVisitor.AllFreezedTokenBalanceFixed(req.GetToken(), req.GetAccount())
	frozenBalances := make([]*rpcpb.FrozenBalance, 0)
	for _, f := range frozen {
		frozenBalances = append(frozenBalances, &rpcpb.FrozenBalance{
			Amount: f.Amount.ToFloat(),
			Time:   f.Ftime,
		})
	}
	return &rpcpb.GetTokenBalanceResponse{
		Balance:        balance,
		FrozenBalances: frozenBalances,
	}, nil
}

// GetContract returns contract information corresponding to the given contract ID.
func (as *APIService) GetContract(ctx context.Context, req *rpcpb.GetContractRequest) (*rpcpb.Contract, error) {
	dbVisitor := as.getStateDBVisitor(req.ByLongestChain)
	contract := dbVisitor.Contract(req.GetId())
	if contract == nil {
		return nil, errors.New("contract not found")
	}
	return toPbContract(contract), nil
}

// GetGasRatio returns gas ratio information in head block
func (as *APIService) GetGasRatio(ctx context.Context, req *rpcpb.EmptyRequest) (*rpcpb.GasRatioResponse, error) {
	ratios := make([]float64, 0)
	for _, tx := range as.bc.Head().Block.Txs {
		if tx.Publisher != "_Block_Base" {
			ratios = append(ratios, float64(tx.GasRatio)/100.0)
		}
	}
	if len(ratios) == 0 {
		return &rpcpb.GasRatioResponse{
			LowestGasRatio: 1.0,
			MedianGasRatio: 1.0,
		}, nil
	}
	sort.Float64s(ratios)
	lowest := ratios[0]
	mid := ratios[len(ratios)/2]
	return &rpcpb.GasRatioResponse{
		LowestGasRatio: lowest,
		MedianGasRatio: mid,
	}, nil
}

// GetContractStorage returns contract storage corresponding to the given key and field.
func (as *APIService) GetContractStorage(ctx context.Context, req *rpcpb.GetContractStorageRequest) (*rpcpb.GetContractStorageResponse, error) {
	dbVisitor := as.getStateDBVisitor(req.ByLongestChain)
	h := host.NewHost(host.NewContext(nil), dbVisitor, nil, nil)
	var value interface{}
	if req.GetField() == "" {
		value, _ = h.GlobalGet(req.GetId(), req.GetKey())
	} else {
		value, _ = h.GlobalMapGet(req.GetId(), req.GetKey(), req.GetField())
	}
	var data string
	if value != nil && reflect.TypeOf(value).Kind() == reflect.String {
		data = value.(string)
	} else {
		bytes, err := json.Marshal(value)
		if err != nil {
			return nil, fmt.Errorf("cannot unmarshal %v", value)
		}
		data = string(bytes)
	}
	return &rpcpb.GetContractStorageResponse{
		Data: data,
	}, nil
}

func (as *APIService) tryTransaction(t *tx.Tx) (*tx.TxReceipt, error) {
	topBlock := as.bc.Head()
	blkHead := &block.BlockHead{
		Version:    0,
		ParentHash: topBlock.HeadHash(),
		Number:     topBlock.Head.Number + 1,
		Time:       time.Now().UnixNano(),
	}
	v := verifier.Verifier{}
	stateDB := as.bv.StateDB().Fork()
	stateDB.Checkout(string(topBlock.HeadHash()))
	return v.Try(blkHead, stateDB, t, cverifier.TxExecTimeLimit)
}

// SendTransaction sends a transaction to iserver.
func (as *APIService) SendTransaction(ctx context.Context, req *rpcpb.TransactionRequest) (*rpcpb.SendTransactionResponse, error) {
	t := toCoreTx(req)
	if as.bv.Config().RPC.TryTx {
		_, err := as.tryTransaction(t)
		if err != nil {
			return nil, fmt.Errorf("try transaction failed: %v", err)
		}
	}
	dbVisitor := as.getStateDBVisitor(true)
	gasLimit := &common.Fixed{Value: t.GasLimit, Decimal: 2}
	gas := dbVisitor.TotalGasAtTime(t.Publisher, as.bc.Head().Head.Time)
	if gas.LessThan(gasLimit) {
		return nil, fmt.Errorf("invalid gas of user %v has %v < %v", t.Publisher, gas.ToString(), gasLimit.ToString())
	}
	err := as.txpool.AddTx(t)
	if err != nil {
		return nil, err
	}
	return &rpcpb.SendTransactionResponse{
		Hash: common.Base58Encode(t.Hash()),
	}, nil
}

// ExecTransaction executes a transaction by the node and returns the receipt.
func (as *APIService) ExecTransaction(ctx context.Context, req *rpcpb.TransactionRequest) (*rpcpb.TxReceipt, error) {
	t := toCoreTx(req)
	receipt, err := as.tryTransaction(t)
	if err != nil {
		return nil, err
	}
	return toPbTxReceipt(receipt), nil
}

// Subscribe used for event.
func (as *APIService) Subscribe(req *rpcpb.SubscribeRequest, res rpcpb.ApiService_SubscribeServer) error {

	topics := make([]event.Topic, 0)
	for _, t := range req.Topics {
		topics = append(topics, event.Topic(t))
	}
	var filter *event.Meta
	if req.GetFilter() != nil {
		filter = &event.Meta{
			ContractID: req.GetFilter().GetContractId(),
		}
	}

	ec := event.GetCollector()
	id := time.Now().UnixNano()
	ch := ec.Subscribe(id, topics, filter)
	defer ec.Unsubscribe(id, topics)

	timeup := time.NewTimer(time.Hour)
	for {
		select {
		case <-timeup.C:
			return nil
		case <-res.Context().Done():
			return res.Context().Err()
		case ev := <-ch:
			e := &rpcpb.Event{
				Topic: rpcpb.Event_Topic(ev.Topic),
				Data:  ev.Data,
				Time:  ev.Time,
			}
			err := res.Send(&rpcpb.SubscribeResponse{Event: e})
			if err != nil {
				ilog.Errorf("stream send failed. err=%v", err)
				return err
			}
		}
	}
}

func (as *APIService) getStateDBVisitor(longestChain bool) *database.Visitor {
	stateDB := as.bv.StateDB().Fork()
	if longestChain {
		stateDB.Checkout(string(as.bc.Head().HeadHash()))
	} else {
		stateDB.Checkout(string(as.bc.LinkedRoot().HeadHash()))
	}
	return database.NewVisitor(0, stateDB)
}
