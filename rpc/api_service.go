package rpc

import (
	"context"

	"github.com/iost-official/go-iost/common"
	"github.com/iost-official/go-iost/consensus/pob"
	"github.com/iost-official/go-iost/core/block"
	"github.com/iost-official/go-iost/core/blockcache"
	"github.com/iost-official/go-iost/core/global"
	"github.com/iost-official/go-iost/core/txpool"
	"github.com/iost-official/go-iost/p2p"
	"github.com/iost-official/go-iost/rpc/pb"
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
	txHashBytes := common.Base58Decode(req.Hash)
	tx, err := as.txpool.GetFromPending(txHashBytes)
	if err == nil {
		return toPbTx(tx, nil, rpcpb.TransactionResponse_PENDIND), nil
	}

	return nil, nil
}

// GetTxReceiptByTxHash returns transaction receipts corresponding to the given tx hash.
func (as *APIService) GetTxReceiptByTxHash(ctx context.Context, req *rpcpb.TxHashRequest) (*rpcpb.TxReceipt, error) {
	return nil, nil
}

// GetBlockByHash returns block corresponding to the given hash.
func (as *APIService) GetBlockByHash(ctx context.Context, req *rpcpb.GetBlockByHashRequest) (*rpcpb.Block, error) {
	return nil, nil
}

// GetBlockByNumber returns block corresponding to the given number.
func (as *APIService) GetBlockByNumber(ctx context.Context, req *rpcpb.GetBlockByNumberRequest) (*rpcpb.Block, error) {
	return nil, nil
}

// GetAccount returns account information corresponding to the given account name.
func (as *APIService) GetAccount(ctx context.Context, req *rpcpb.GetAccountRequest) (*rpcpb.Account, error) {
	return nil, nil
}

// GetContract returns contract information corresponding to the given contract ID.
func (as *APIService) GetContract(ctx context.Context, req *rpcpb.GetContractRequest) (*rpcpb.Contract, error) {
	return nil, nil
}

// GetContractStorage returns contract storage corresponding to the given key and field.
func (as *APIService) GetContractStorage(ctx context.Context, req *rpcpb.GetContractStorageRequest) (*rpcpb.GetContractStorageResponse, error) {
	return nil, nil
}

// SendTransaction sends a transaction to iserver.
func (as *APIService) SendTransaction(ctx context.Context, req *rpcpb.TransactionRequest) (*rpcpb.SendTransactionResponse, error) {
	return nil, nil
}

// ExecTransaction executes a transaction by the node and returns the receipt.
func (as *APIService) ExecTransaction(ctx context.Context, req *rpcpb.TransactionRequest) (*rpcpb.TxReceipt, error) {
	return nil, nil
}
