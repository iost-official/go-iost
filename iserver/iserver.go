package iserver

import (
	"github.com/iost-official/go-iost/account"
	"github.com/iost-official/go-iost/common"
	"github.com/iost-official/go-iost/consensus"
	"github.com/iost-official/go-iost/consensus/synchronizer"
	"github.com/iost-official/go-iost/core/blockcache"
	"github.com/iost-official/go-iost/core/global"
	"github.com/iost-official/go-iost/core/tx"
	"github.com/iost-official/go-iost/core/txpool"
	"github.com/iost-official/go-iost/crypto"
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
	sync      *synchronizer.SyncImpl
	txp       *txpool.TxPImpl
	rpcServer *rpc.Server
	consensus consensus.Consensus
	debug     *DebugServer
}

// New returns a iserver application
func New(conf *common.Config) *IServer {
	tx.ChainID = conf.P2P.ChainID

	bv, err := global.New(conf)
	if err != nil {
		ilog.Fatalf("create global failed. err=%v", err)
	}
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

	accSecKey := conf.ACC.SecKey
	acc, err := account.NewKeyPair(common.Base58Decode(accSecKey), crypto.NewAlgorithm(conf.ACC.Algorithm))
	if err != nil {
		ilog.Fatalf("NewKeyPair failed, stop the program! err:%v", err)
	}

	blkCache, err := blockcache.NewBlockCache(bv)
	if err != nil {
		ilog.Fatalf("blockcache initialization failed, stop the program! err:%v", err)
	}

	txp, err := txpool.NewTxPoolImpl(bv, blkCache, p2pService)
	if err != nil {
		ilog.Fatalf("txpool initialization failed, stop the program! err:%v", err)
	}

	consensus := consensus.New(consensus.Pob, acc, bv, blkCache, txp, p2pService)

	rpcServer := rpc.New(txp, blkCache, bv, p2pService)

	sync, err := synchronizer.NewSynchronizer(bv, blkCache, p2pService)
	if err != nil {
		ilog.Fatalf("synchronizer initialization failed, stop the program! err:%v", err)
	}

	debug := NewDebugServer(conf.Debug, p2pService, blkCache, bv.BlockChain())

	return &IServer{
		bv:        bv,
		p2p:       p2pService,
		sync:      sync,
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
		s.sync,
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
		s.sync,
		s.p2p,
	}
	for _, s := range Services {
		s.Stop()
	}
	s.bv.BlockChain().Close()
	s.bv.StateDB().Close()
}
