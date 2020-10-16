package chainbase

import (
	"errors"
	"time"

	"github.com/iost-official/go-iost/common"
	"github.com/iost-official/go-iost/consensus/cverifier"
	"github.com/iost-official/go-iost/core/block"
	"github.com/iost-official/go-iost/core/blockcache"
	"github.com/iost-official/go-iost/ilog"
	"github.com/iost-official/go-iost/verifier"
)

var (
	errSingle     = errors.New("single block")
	errDuplicate  = errors.New("duplicate block")
	errOutOfLimit = errors.New("block out of limit in one slot")
	errWitness    = errors.New("wrong witness")
	errTxDup      = errors.New("duplicate tx")
	errDoubleTx   = errors.New("double tx in block")
	errDelayTx    = errors.New("delaytx is not allowed")
)

// Block will describe the block of chainbase.
type Block struct {
	*block.Block
	*blockcache.WitnessList
	Irreversible bool
}

// HeadBlock will return the head block of chainbase.
func (c *ChainBase) HeadBlock() *Block {
	head := c.bCache.Head()
	block := &Block{
		Block:        head.Block,
		WitnessList:  &head.WitnessList,
		Irreversible: false,
	}
	return block
}

// LIBlock will return the last irreversible block of chainbase.
func (c *ChainBase) LIBlock() *Block {
	lib := c.bCache.LinkedRoot()
	block := &Block{
		Block:        lib.Block,
		WitnessList:  &lib.WitnessList,
		Irreversible: true,
	}
	return block
}

// GetBlockByHash will return the block by hash.
// If block is not exist, it will return nil and false.
func (c *ChainBase) GetBlockByHash(hash []byte) (*Block, bool) {
	block, err := c.bCache.GetBlockByHash(hash)
	if err != nil {
		block, err := c.bChain.GetBlockByHash(hash)
		if err != nil {
			ilog.Warnf("Get block by hash %v failed: %v", common.Base58Encode(hash), err)
			return nil, false
		}
		return &Block{
			Block:        block,
			Irreversible: true,
		}, true
	}
	return &Block{
		Block:        block,
		Irreversible: false,
	}, true
}

// GetBlockHashByNum will return the block hash by number.
// If block hash is not exist, it will return nil and false.
func (c *ChainBase) GetBlockHashByNum(num int64) ([]byte, bool) {
	var hash []byte
	if blk, err := c.bCache.GetBlockByNumber(num); err != nil {
		hash, err = c.bChain.GetHashByNumber(num)
		if err != nil {
			ilog.Debugf("Get hash by num %v failed: %v", num, err)
			return nil, false
		}
	} else {
		hash = blk.HeadHash()
	}
	return hash, true
}

func (c *ChainBase) printStatistics(num int64, blk *block.Block, replay bool, gen bool) {
	action := "Recover"
	if !replay {
		if gen {
			action = "Generate"
		} else {
			action = "Receive"
		}
	}
	ptx, _ := c.txPool.PendingTx()
	ilog.Infof("%v block - @%v id:%v..., t:%v, num:%v, confirmed:%v, txs:%v, pendingtxs:%v, et:%dms",
		action,
		num,
		blk.Head.Witness[:10],
		blk.Head.Time,
		blk.Head.Number,
		c.bCache.LinkedRoot().Head.Number,
		len(blk.Txs),
		ptx.Size(),
		(time.Now().UnixNano()-blk.Head.Time)/1e6,
	)
}

// Add will add a block to block cache and verify it.
func (c *ChainBase) Add(blk *block.Block, replay bool, gen bool) error {
	ilog.Debug("add block ", blk.Head.Number, " to chain base")
	_, err := c.bCache.GetBlockByHash(blk.HeadHash())
	if err == nil {
		return errDuplicate
	}

	err = blk.VerifySelf()
	if err != nil {
		ilog.Warnf("Verify block basics failed: %v", err)
		return err
	}

	node := c.bCache.Add(blk)
	parent := node.GetParent()
	if parent.Type != blockcache.Linked {
		return errSingle
	}
	if err := c.addExistingBlock(node, replay, gen); err != nil {
		ilog.Warnf("verify block execute failed, blockNum: %v, blockHash: %v, err: %v", node.Head.Number, common.Base58Encode(node.HeadHash()), err)
		return err
	}

	return nil
}

func (c *ChainBase) addExistingBlock(node *blockcache.BlockCacheNode, replay bool, gen bool) error {
	blk := node.Block
	parentNode := node.GetParent()

	if parentNode.Block.Head.Witness != blk.Head.Witness ||
		common.SlotOfUnixNano(parentNode.Block.Head.Time) != common.SlotOfUnixNano(blk.Head.Time) {
		node.SerialNum = 0
	} else {
		node.SerialNum = parentNode.SerialNum + 1
	}

	if node.SerialNum >= int64(common.BlockNumPerWitness) {
		return errOutOfLimit
	}
	ok := c.stateDB.Checkout(string(blk.HeadHash()))
	if !ok {
		c.stateDB.Checkout(string(blk.Head.ParentHash))
		err := c.verifyBlock(blk, parentNode.Block, &node.GetParent().WitnessList)
		if err != nil {
			// TODO: Decouple add and link of blockcache, then remove the Del().
			c.bCache.Del(node)
			return err
		}
		c.stateDB.Commit(string(blk.HeadHash()))
	}
	c.bCache.Link(node)
	if !replay {
		c.bCache.AddNodeToWAL(node)
	}
	c.bCache.UpdateLib(node)
	// After UpdateLib, the block head active witness list will be right
	// So AddLinkedNode need execute after UpdateLib
	c.txPool.AddLinkedNode(node)
	if replay {
		ilog.Infof("node %d %s active list: %v", node.Head.Number, common.Base58Encode(node.HeadHash()), node.Active())
	}

	c.printStatistics(node.SerialNum, node.Block, replay, gen)

	for child := range node.Children {
		if err := c.addExistingBlock(child, replay, gen); err != nil {
			ilog.Warnf("verify block execute failed, blockNum: %v, blockHash: %v, err: %v", node.Head.Number, common.Base58Encode(node.HeadHash()), err)
		}
	}
	return nil
}

func (c *ChainBase) verifyBlock(blk, parent *block.Block, witnessList *blockcache.WitnessList) error {
	err := cverifier.VerifyBlockHead(blk, parent)
	if err != nil {
		return err
	}

	if common.WitnessOfNanoSec(blk.Head.Time, witnessList.Active()) != blk.Head.Witness {
		ilog.Errorf("blk num: %v, time: %v, witness: %v, witness len: %v, witness list: %v",
			blk.Head.Number, blk.Head.Time, blk.Head.Witness, len(witnessList.Active()), witnessList.Active())
		return errWitness
	}
	ilog.Debugf("[pob] start to verify block if foundchain, number: %v, hash = %v, witness = %v", blk.Head.Number, common.Base58Encode(blk.HeadHash()), blk.Head.Witness[4:6])
	blkTxSet := make(map[string]bool, len(blk.Txs))
	rules := blk.Head.Rules()
	for i, t := range blk.Txs {
		if blkTxSet[string(t.Hash())] {
			return errDoubleTx
		}
		blkTxSet[string(t.Hash())] = true

		if i == 0 {
			// base tx
			continue
		}
		if rules.IsFork3_1_0 {
			// reject delay tx
			if t.Delay > 0 {
				return errDelayTx
			}
		}
		if c.txPool.ExistTxs(t.Hash(), parent) {
			ilog.Infof("FoundChain: %v, %v", t, common.Base58Encode(t.Hash()))
			return errTxDup
		}
		err := t.VerifySelf()
		if err != nil {
			return err
		}
	}
	v := verifier.Verifier{}
	if c.config.SPV != nil && c.config.SPV.IsSPV {
		// in SPV mode, only verify the block structure, not exec the txs
		return nil
	}
	return v.Verify(blk, parent, witnessList, c.stateDB, &verifier.Config{
		Mode:        0,
		Timeout:     common.MaxBlockTimeLimit,
		TxTimeLimit: common.MaxTxTimeLimit,
	})
}
