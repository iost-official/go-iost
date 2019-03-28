package chainbase

import (
	"fmt"
	"os"

	"github.com/iost-official/go-iost/common"
	"github.com/iost-official/go-iost/consensus/snapshot"
	"github.com/iost-official/go-iost/core/block"
	"github.com/iost-official/go-iost/core/blockcache"
	"github.com/iost-official/go-iost/core/txpool"
	"github.com/iost-official/go-iost/db"
	"github.com/iost-official/go-iost/ilog"
)

// ChainBase will maintain blockchain data for memory and hard disk.
type ChainBase struct {
	config  *common.Config
	bChain  block.Chain
	bCache  blockcache.BlockCache
	stateDB db.MVCCDB
	txPool  txpool.TxPool
}

// New will return a ChainBase.
func New(conf *common.Config) (*ChainBase, error) {
	if conf.Snapshot.Enable {
		conf.Snapshot.Enable = false
		s, err := os.Stat(conf.DB.LdbPath + "BlockChainDB")
		if err == nil && s.IsDir() {
			ilog.Warnln("start iserver with the snapshot failed, blockchain db already has.")
		} else {
			s, err = os.Stat(conf.DB.LdbPath + "StateDB")
			if err == nil && s.IsDir() {
				ilog.Warnln("start iserver with the snapshot failed, state db already has.")
			} else {
				err = snapshot.FromSnapshot(conf)
				if err != nil {
					ilog.Fatalf("start iserver with the snapshot failed, err:%v", err)
				}
				conf.Snapshot.Enable = true
			}
		}
	}

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
	}

	if err := c.checkGenesis(conf); err != nil {
		return nil, fmt.Errorf("Check genesis failed: %v", err)
	}
	if err := c.recoverDB(conf); err != nil {
		return nil, fmt.Errorf("Recover DB failed: %v", err)
	}

	bCache, err := blockcache.NewBlockCache(conf, bChain, stateDB)
	if err != nil {
		return nil, fmt.Errorf("blockcache initialization failed, stop the program! err:%v", err)
	}
	c.bCache = bCache

	return c, nil
}

// Close will close the chainbase.
func (c *ChainBase) Close() {
	c.bChain.Close()
	c.stateDB.Close()
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

// SetTxPool will inject the tx pool for chainbase.
func (c *ChainBase) SetTxPool(txPool txpool.TxPool) {
	c.txPool = txPool
}
