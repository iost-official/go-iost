package iserver

import (
	"github.com/iost-official/go-iost/chainbase"
	"github.com/iost-official/go-iost/common"
	"github.com/iost-official/go-iost/consensus"
	"github.com/iost-official/go-iost/core/blockcache"
	"github.com/iost-official/go-iost/core/global"
	"github.com/iost-official/go-iost/core/tx"
	"github.com/iost-official/go-iost/core/txpool"
	"github.com/iost-official/go-iost/ilog"
	"github.com/iost-official/go-iost/p2p"
	"github.com/iost-official/go-iost/rpc"
)

// Service defines APIs of resident goroutines.
type Service interface {
	Start() error
	Stop()
}

// IServer is application for IOST.
type IServer struct {
	bv        global.BaseVariable
	p2p       *p2p.NetService
	txp       *txpool.TxPImpl
	rpcServer *rpc.Server
	consensus consensus.Consensus
	debug     *DebugServer
}

// New returns a iserver application
func New(conf *common.Config) *IServer {
	tx.ChainID = conf.P2P.ChainID

	chainBase, err := chainbase.New(conf)
	if err != nil {
		ilog.Fatalf("New chainbase failed: %v.", err)
	}
	bv := global.New(chainBase, conf)

	if err := checkGenesis(bv); err != nil {
		ilog.Fatalf("Check genesis failed: %v", err)
	}
	if err := recoverDB(bv); err != nil {
		ilog.Fatalf("Recover DB failed: %v", err)
	}

	p2pService, err := p2p.NewNetService(conf.P2P)
	if err != nil {
		ilog.Fatalf("network initialization failed, stop the program! err:%v", err)
	}

	blkCache, err := blockcache.NewBlockCache(conf, bv.BlockChain(), bv.StateDB())
	if err != nil {
		ilog.Fatalf("blockcache initialization failed, stop the program! err:%v", err)
	}

	txp, err := txpool.NewTxPoolImpl(bv, blkCache, p2pService)
	if err != nil {
		ilog.Fatalf("txpool initialization failed, stop the program! err:%v", err)
	}

	consensus := consensus.New(consensus.Pob, bv, blkCache, txp, p2pService)

	rpcServer := rpc.New(txp, blkCache, bv, p2pService, consensus)

	debug := NewDebugServer(conf.Debug, p2pService, blkCache, bv.BlockChain())

	return &IServer{
		bv:        bv,
		p2p:       p2pService,
		txp:       txp,
		rpcServer: rpcServer,
		consensus: consensus,
		debug:     debug,
	}
}

// Start starts iserver application.
func (s *IServer) Start() error {
	Services := []Service{
		s.p2p,
		s.txp,
		s.consensus,
		s.rpcServer,
	}
	for _, s := range Services {
		if err := s.Start(); err != nil {
			return err
		}
	}
	conf := s.bv.Config()
	if conf.Debug != nil {
		if err := s.debug.Start(); err != nil {
			return err
		}
	}
	return nil
}

// Stop stops iserver application.
func (s *IServer) Stop() {
	conf := s.bv.Config()
	if conf.Debug != nil {
		s.debug.Stop()
	}
	Services := []Service{
		s.rpcServer,
		s.consensus,
		s.txp,
		s.p2p,
	}
	for _, s := range Services {
		s.Stop()
	}
	s.bv.BlockChain().Close()
	s.bv.StateDB().Close()
}
