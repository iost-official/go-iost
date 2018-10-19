package iserver

import (
	"fmt"
	"time"

	"github.com/iost-official/go-iost/common"
	"github.com/iost-official/go-iost/consensus/genesis"
	"github.com/iost-official/go-iost/consensus/verifier"
	"github.com/iost-official/go-iost/core/global"
	"github.com/iost-official/go-iost/ilog"
)

func recoverDB(bv global.BaseVariable) error {
	blockChain := bv.BlockChain()
	stateDB := bv.StateDB()
	txDB := bv.TxDB()
	conf := bv.Config()

	blk, err := blockChain.GetBlockByNumber(0)
	if err != nil { //blockchaindb is empty
		hash := stateDB.CurrentTag()
		if hash != "" {
			return fmt.Errorf("blockchaindb is empty, but statedb is not")
		}

		blk, err = genesis.GenGenesisByFile(stateDB, conf.Genesis)
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
		err = txDB.Push(blk.Txs, blk.Receipts)
		if err != nil {
			return fmt.Errorf("push txDB failed, stop the pogram. err: %v", err)
		}
		genesisBlock, _ := blockChain.GetBlockByNumber(0)
		ilog.Infof("createGenesisHash: %v", common.Base58Encode(genesisBlock.HeadHash()))
		return nil
	}
	var startNumebr int64
	hash := stateDB.CurrentTag()
	if hash != "" {
		blk, err = blockChain.GetBlockByHash([]byte(hash))
		if err != nil {
			return fmt.Errorf("statedb doesn't coincides with blockchaindb. err: %v", err)
		}
		startNumebr = blk.Head.Number + 1
	}
	for i := startNumebr; i < blockChain.Length(); i++ {
		blk, err = blockChain.GetBlockByNumber(i)
		if err != nil {
			return fmt.Errorf("get block by number failed, stop the pogram. err: %v", err)
		}
		v := verifier.Verifier{}
		err = v.Verify(blk, stateDB, &verifier.Config{
			Mode:        0,
			Timeout:     common.SlotLength / 3 * time.Second,
			TxTimeLimit: time.Millisecond * 100,
		})
		if err != nil {
			return fmt.Errorf("verify block with VM failed, stop the pogram. err: %v", err)
		}
		stateDB.Tag(string(blk.HeadHash()))
		err = stateDB.Flush(string(blk.HeadHash()))
		if err != nil {
			return fmt.Errorf("flush stateDB failed, stop the pogram. err: %v", err)
		}
	}
	return nil
}
