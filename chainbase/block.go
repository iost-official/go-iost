package chainbase

import (
	"errors"
	"sync"
	"time"

	"github.com/iost-official/go-iost/account"
	"github.com/iost-official/go-iost/common"
	"github.com/iost-official/go-iost/consensus/synchro"
	"github.com/iost-official/go-iost/core/block"
	"github.com/iost-official/go-iost/core/blockcache"
	"github.com/iost-official/go-iost/core/txpool"
	"github.com/iost-official/go-iost/crypto"
	"github.com/iost-official/go-iost/db"
	"github.com/iost-official/go-iost/ilog"
	"github.com/iost-official/go-iost/metrics"
	"github.com/iost-official/go-iost/p2p"
)

// Add will add a block to block cache and verify it.
func (p *PoB) Add(blk *block.Block, replay bool) error {
	_, err := p.blockCache.Find(blk.HeadHash())
	if err == nil {
		return errDuplicate
	}

	err = verifyBasics(blk, blk.Sign)
	if err != nil {
		return err
	}

	parent, err := p.blockCache.Find(blk.Head.ParentHash)
	p.blockCache.Add(blk)
	if err == nil && parent.Type == blockcache.Linked {
		return p.addExistingBlock(blk, parent, replay)
	}
	return errSingle
}

func (p *PoB) addExistingBlock(blk *block.Block, parentNode *blockcache.BlockCacheNode, replay bool) error {
	node, _ := p.blockCache.Find(blk.HeadHash())

	if parentNode.Block.Head.Witness != blk.Head.Witness ||
		common.SlotOfNanoSec(parentNode.Block.Head.Time) != common.SlotOfNanoSec(blk.Head.Time) {
		node.SerialNum = 0
	} else {
		node.SerialNum = parentNode.SerialNum + 1
	}

	if node.SerialNum >= int64(blockNumPerWitness) {
		return errOutOfLimit
	}
	ok := p.verifyDB.Checkout(string(blk.HeadHash()))
	if !ok {
		p.verifyDB.Checkout(string(blk.Head.ParentHash))
		p.txPool.Lock()
		err := verifyBlock(blk, parentNode.Block, &node.GetParent().WitnessList, p.txPool, p.verifyDB, p.blockChain, replay)
		p.txPool.Release()
		if err != nil {
			ilog.Errorf("verify block failed, blockNum:%v, blockHash:%v. err=%v", blk.Head.Number, common.Base58Encode(blk.HeadHash()), err)
			p.blockCache.Del(node)
			return err
		}
		p.verifyDB.Commit(string(blk.HeadHash()))
	}
	p.blockCache.Link(node, replay)
	p.blockCache.UpdateLib(node)
	// After UpdateLib, the block head active witness list will be right
	// So AddLinkedNode need execute after UpdateLib
	p.txPool.AddLinkedNode(node)

	metricsConfirmedLength.Set(float64(p.blockCache.LinkedRoot().Head.Number), nil)

	p.printStatistics(node.SerialNum, node.Block)

	for child := range node.Children {
		p.addExistingBlock(child.Block, node, replay)
	}
	return nil
}
