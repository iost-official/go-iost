package chainbase

import (
	"fmt"

	"github.com/iost-official/go-iost/common"
	"github.com/iost-official/go-iost/consensus/genesis"
	"github.com/iost-official/go-iost/ilog"
)

func (c *ChainBase) checkGenesis(conf *common.Config) error {
	blockChain := c.BlockChain()
	stateDB := c.StateDB()
	if !conf.Snapshot.Enable && blockChain.Length() == int64(0) { //blockchaindb is empty
		// TODO: remove the module of starting iserver from yaml.

		ilog.Infof("Genesis is not exist.")
		hash := stateDB.CurrentTag()
		if hash != "" {
			return fmt.Errorf("blockchaindb is empty, but statedb is not")
		}

		blk, err := genesis.GenGenesisByFile(stateDB, conf.Genesis)
		if err != nil {
			return fmt.Errorf("new GenGenesis failed, stop the program. err: %v", err)
		}
		err = blockChain.Push(blk)
		if err != nil {
			return fmt.Errorf("push block in blockChain failed, stop the program. err: %v", err)
		}

		err = stateDB.Flush(string(blk.HeadHash()))
		if err != nil {
			return fmt.Errorf("flush block into stateDB failed, stop the program. err: %v", err)
		}
		ilog.Infof("Created Genesis.")

		// TODO check genesis hash between config and db
		ilog.Infof("GenesisHash: %v", common.Base58Encode(blk.HeadHash()))
	}
	return nil
}

func (c *ChainBase) recoverDB(conf *common.Config) error {
	blockChain := c.BlockChain()
	stateDB := c.StateDB()

	if conf.Snapshot.Enable {
		/*
			blk, err := snapshot.Load(stateDB)
			if err != nil {
				return fmt.Errorf("load block from snapshot failed. err: %v", err)
			}
			blockChain.SetLength(blk.Head.Number + 1)
		*/
	} else {
		length := int64(0)
		hash := stateDB.CurrentTag()
		ilog.Infoln("current Tag:", common.Base58Encode([]byte(hash)))
		//var parent *block.Block
		if hash != "" {
			blk, err := blockChain.GetBlockByHash([]byte(hash))
			if err != nil {
				return fmt.Errorf("statedb doesn't coincides with blockchaindb. err: %v", err)
			}
			length = blk.Head.Number + 1
		}
		blockChain.SetLength(length)
	}
	return nil
}

func (c *ChainBase) Recover() error {
	err := c.bCache.Recover(c)
	if err != nil {
		ilog.Error("Failed to recover blockCache, err: ", err)
		ilog.Info("Don't Recover, Move old file to BlockCacheWALCorrupted")
		err = c.bCache.NewWAL(c.config)
		if err != nil {
			ilog.Error(" Failed to NewWAL, err: ", err)
		}
	}
	return err
}
