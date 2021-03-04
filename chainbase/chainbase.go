package chainbase

import (
	"fmt"
	"sync"

	"github.com/iost-official/go-iost/v3/common"
	"github.com/iost-official/go-iost/v3/core/block"
	"github.com/iost-official/go-iost/v3/core/blockcache"
	"github.com/iost-official/go-iost/v3/core/txpool"
	"github.com/iost-official/go-iost/v3/db"
	"github.com/iost-official/go-iost/v3/ilog"
)

// ChainBase will maintain blockchain data for memory and hard disk.
type ChainBase struct {
	config  *common.Config
	bChain  block.Chain
	bCache  blockcache.BlockCache
	stateDB db.MVCCDB
	txPool  txpool.TxPool

	quitCh chan struct{}
	done   *sync.WaitGroup
}

// New will return a ChainBase.
func New(conf *common.Config) (*ChainBase, error) {
	bChain, err := block.NewBlockChain(conf.DB.LdbPath + "BlockChainDB")
	if err != nil {
		return nil, fmt.Errorf("new blockchain failed, stop the program. err: %v", err)
	}

	stateDB, err := db.NewMVCCDB(conf.DB.LdbPath + "StateDB")
	if err != nil {
		return nil, fmt.Errorf("new statedb failed, stop the program. err: %v", err)
	}

	c := &ChainBase{
		config:  conf,
		bChain:  bChain,
		stateDB: stateDB,

		quitCh: make(chan struct{}),
		done:   new(sync.WaitGroup),
	}
	if conf.SPV != nil && conf.SPV.IsSPV {
		if bChain.Length() == 0 {
			syncFrom := conf.SPV.SyncFromBlock
			blk, err := SPVFetchInitialBlockFromSeed(conf.SPV.SeedServer, syncFrom)
			if err != nil {
				return nil, err
			}
			// we trust this seed block in spv mode
			err = bChain.Push(blk)
			if err != nil {
				return nil, err
			}
		} else { // nolint
			// recover db?
		}
	} else {
		if err := c.checkGenesis(conf); err != nil {
			return nil, fmt.Errorf("check genesis failed: %v", err)
		}
		if err := c.recoverDB(conf); err != nil {
			return nil, fmt.Errorf("recover database failed: %v", err)
		}
	}

	ilog.Info("recover db done")

	bCache, err := blockcache.NewBlockCache(conf, bChain, stateDB)
	if err != nil {
		return nil, fmt.Errorf("initialize blockcache failed: %v", err)
	}
	c.bCache = bCache

	txPool, err := txpool.NewTxPoolImpl(bChain, bCache)
	if err != nil {
		return nil, fmt.Errorf("initialize txpool failed: %v", err)
	}
	c.txPool = txPool

	if err := c.recoverBlockCache(); err != nil {
		return nil, fmt.Errorf("recover chainbase failed: %v", err)
	}
	ilog.Info("recover block cache done")

	c.done.Add(1)
	go c.metricsController()

	return c, nil
}

// Close will close the chainbase.
func (c *ChainBase) Close() {
	close(c.quitCh)
	c.done.Wait()

	c.txPool.Close()
	c.stateDB.Close()
	c.bChain.Close()

	ilog.Infof("Closed chainbase.")
}

// =============== Temporarily compatible ===============

// StateDB return the state database.
func (c *ChainBase) StateDB() db.MVCCDB {
	return c.stateDB
}

// BlockChain return the block chain database.
func (c *ChainBase) BlockChain() block.Chain {
	return c.bChain
}

// BlockCache return the block cache.
func (c *ChainBase) BlockCache() blockcache.BlockCache {
	return c.bCache
}

// TxPool will return the tx pool.
func (c *ChainBase) TxPool() txpool.TxPool {
	return c.txPool
}

// NewMock will return the chainbase composed of blockchain and blockcache.
func NewMock(bChain block.Chain, bCache blockcache.BlockCache) *ChainBase {
	return &ChainBase{
		bChain: bChain,
		bCache: bCache,
	}
}
