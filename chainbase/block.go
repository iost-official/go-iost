package chainbase

import (
	"errors"
	"time"

	"github.com/iost-official/go-iost/common"
	"github.com/iost-official/go-iost/core/block"
	"github.com/iost-official/go-iost/core/blockcache"
	"github.com/iost-official/go-iost/ilog"
)

var (
	errSingle     = errors.New("single block")
	errDuplicate  = errors.New("duplicate block")
	errOutOfLimit = errors.New("block out of limit in one slot")
)

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
	_, err := c.bCache.Find(blk.HeadHash())
	if err == nil {
		return errDuplicate
	}

	err = blk.VerifySelf()
	if err != nil {
		ilog.Warnf("Verify block basics failed: %v", err)
		return err
	}

	parent, err := c.bCache.Find(blk.Head.ParentHash)
	c.bCache.Add(blk)
	if err == nil && parent.Type == blockcache.Linked {
		err := c.addExistingBlock(blk, parent, replay, gen)
		if err != nil {
			ilog.Warnf("Verify block execute failed: %v", err)
		}
		return err
	}
	return errSingle
}

func (c *ChainBase) addExistingBlock(blk *block.Block, parentNode *blockcache.BlockCacheNode, replay bool, gen bool) error {
	node, _ := c.bCache.Find(blk.HeadHash())

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
		err := verifyBlock(blk, parentNode.Block, &node.GetParent().WitnessList, c.txPool, c.stateDB, c.bChain, replay)
		if err != nil {
			ilog.Errorf("verify block failed, blockNum:%v, blockHash:%v. err=%v", blk.Head.Number, common.Base58Encode(blk.HeadHash()), err)
			c.bCache.Del(node)
			return err
		}
		c.stateDB.Commit(string(blk.HeadHash()))
	}
	c.bCache.Link(node, replay)
	c.bCache.UpdateLib(node)
	// After UpdateLib, the block head active witness list will be right
	// So AddLinkedNode need execute after UpdateLib
	c.txPool.AddLinkedNode(node)
	if replay {
		ilog.Infof("node %d %s active list: %v", node.Head.Number, common.Base58Encode(node.HeadHash()), node.Active())
	}

	c.printStatistics(node.SerialNum, node.Block, replay, gen)

	for child := range node.Children {
		c.addExistingBlock(child.Block, node, replay, gen)
	}
	return nil
}
