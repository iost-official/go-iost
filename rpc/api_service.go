package rpc

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"

	"github.com/iost-official/go-iost/common"
	"github.com/iost-official/go-iost/consensus/pob"
	"github.com/iost-official/go-iost/core/block"
	"github.com/iost-official/go-iost/core/blockcache"
	"github.com/iost-official/go-iost/core/global"
	"github.com/iost-official/go-iost/core/tx"
	"github.com/iost-official/go-iost/core/txpool"
	"github.com/iost-official/go-iost/p2p"
	"github.com/iost-official/go-iost/rpc/pb"
	"github.com/iost-official/go-iost/vm/database"
	"github.com/iost-official/go-iost/vm/host"
)

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

// GetChainInfo returns the chain info.
func (as *APIService) GetChainInfo(context.Context, *rpcpb.EmptyRequest) (*rpcpb.ChainInfoResponse, error) {
	headBlock := as.bc.Head().Block
	libBlock := as.bc.LinkedRoot().Block
	return &rpcpb.ChainInfoResponse{
		NetName:         as.bv.Config().Version.NetName,
		ProtocolVersion: as.bv.Config().Version.ProtocolVersion,
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
	status := rpcpb.TransactionResponse_PENDIND
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
	status := rpcpb.BlockResponse_PENDIND
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
	status := rpcpb.BlockResponse_PENDIND
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
	acc, _ := host.ReadAuth(dbVisitor, req.GetName())
	if acc == nil {
		return nil, errors.New("account not found")
	}
	ret := toPbAccount(acc)
	balance := dbVisitor.TokenBalanceFixed("iost", req.GetName()).ToString()
	ret.Balance = balance
	ret.RamInfo = &rpcpb.Account_RAMInfo{
		Available: dbVisitor.TokenBalance("ram", req.GetName()),
	}

	var blkTime int64
	if req.GetByLongestChain() {
		blkTime = as.bc.Head().Head.Time
	} else {
		blkTime = as.bc.LinkedRoot().Head.Time
	}
	gasManager := host.NewGasManager(host.NewHost(host.NewContext(nil), dbVisitor, nil, nil))
	totalGas, _ := gasManager.CurrentTotalGas(req.GetName(), blkTime)
	gasLimit, _ := gasManager.GasLimit(req.GetName())
	gasRate, _ := gasManager.GasRate(req.GetName())
	ret.GasInfo = &rpcpb.Account_GasInfo{
		CurrentTotal:  totalGas.ToFloat(),
		Limit:         gasLimit.ToFloat(),
		IncreaseSpeed: gasRate.ToFloat(),
	}

	// TODO: pack pledged info and frozen balance
	return ret, nil
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

// GetContractStorage returns contract storage corresponding to the given key and field.
func (as *APIService) GetContractStorage(ctx context.Context, req *rpcpb.GetContractStorageRequest) (*rpcpb.GetContractStorageResponse, error) {
	dbVisitor := as.getStateDBVisitor(req.ByLongestChain)
	h := host.NewHost(host.NewContext(nil), dbVisitor, nil, nil)
	var value interface{}
	if req.GetField() == "" {
		value, _ = h.GlobalGet(req.GetId(), req.GetKey(), req.GetOwner())
	} else {
		value, _ = h.GlobalMapGet(req.GetId(), req.GetKey(), req.GetField(), req.GetOwner())
	}
	var data string
	if reflect.TypeOf(value).Kind() != reflect.String {
		bytes, err := json.Marshal(value)
		if err != nil {
			return nil, fmt.Errorf("cannot unmarshal %v", value)
		}
		data = string(bytes)
	} else {
		data = value.(string)
	}
	return &rpcpb.GetContractStorageResponse{
		Data: data,
	}, nil
}

// SendTransaction sends a transaction to iserver.
func (as *APIService) SendTransaction(ctx context.Context, req *rpcpb.TransactionRequest) (*rpcpb.SendTransactionResponse, error) {
	return nil, nil
}

// ExecTransaction executes a transaction by the node and returns the receipt.
func (as *APIService) ExecTransaction(ctx context.Context, req *rpcpb.TransactionRequest) (*rpcpb.TxReceipt, error) {
	return nil, nil
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
