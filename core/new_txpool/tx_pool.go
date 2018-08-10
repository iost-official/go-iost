package new_txpool

import (
	"fmt"
	"sort"
	"sync"
	"time"

	"bytes"
	"errors"
	"os"

	"github.com/iost-official/Go-IOS-Protocol/consensus/common"
	"github.com/iost-official/Go-IOS-Protocol/core/global"
	"github.com/iost-official/Go-IOS-Protocol/core/message"
	"github.com/iost-official/Go-IOS-Protocol/core/new_block"
	"github.com/iost-official/Go-IOS-Protocol/core/new_blockcache"
	"github.com/iost-official/Go-IOS-Protocol/core/new_tx"
	"github.com/iost-official/Go-IOS-Protocol/log"
	"github.com/iost-official/Go-IOS-Protocol/p2p"
	"github.com/prometheus/client_golang/prometheus"
)

func init() {
	prometheus.MustRegister(receivedTransactionCount)
}

type TxPoolImpl struct {
	chTx         chan message.Message
	chLinkedNode chan *RecNode

	global global.Global
	chain  blockcache.BlockCache
	p2pService p2p.Service

	forkChain *ForkChain
	blockList *sync.Map
	pendingTx *sync.Map

	mu sync.RWMutex
}

func NewTxPoolImpl(global global.Global chain blockcache.BlockCache, p2pserv p2p.Service) (TxPool, error) {
	p := &TxPoolImpl{
		chain:        chain,
		chLinkedNode: make(chan *RecNode, 100),
		forkChain:    new(ForkChain),
		blockList:    new(sync.Map),
		pendingTx:    new(sync.Map),
		global:       global,
	}
	p.p2pService = p2pserv
	p.chTx = p.p2pService.Register("TxPool message", p2p.PublishTxRequest)
	return p, nil
}

func (pool *TxPoolImpl) Start() {
	log.Log.I("TxPoolImpl Start")
	go pool.loop()
}

func (pool *TxPoolImpl) Stop() {
	log.Log.I("TxPoolImpl Stop")
	close(pool.chTx)
	close(pool.chLinkedNode)
}

func (pool *TxPoolImpl) loop() {

	pool.initBlockTx()

	clearTx := time.NewTicker(clearInterval)
	defer clearTx.Stop()

	for {
		select {
		case tr, ok := <-pool.chTx:
			if !ok {
				log.Log.E("tx_pool - chTx is error")
				os.Exit(1)
			}

			var tx tx.Tx
			err := tx.Decode(tr.Body)
			if err != nil {
				continue
			}

			if pool.txTimeOut(&tx) {
				continue
			}

			if tx.VerifySelf() != nil {
				pool.mu.Lock()

				pool.addTx(&tx)

				pool.mu.Unlock()
				receivedTransactionCount.Inc()
			}

		case bl, ok := <-pool.chLinkedNode:
			if !ok {
				log.Log.E("tx_pool - chLinkedNode is error")
				os.Exit(1)
			}

			if pool.addBlock(bl.LinkedNode.Block) != nil {
				continue
			}

			pool.mu.Lock()

			tFort := pool.updateForkChain(bl.HeadNode)
			switch tFort {
			case ForkError:
				log.Log.E("tx_pool - updateForkChain is error")
				pool.clearTxPending()

			case Fork:
				if err := pool.doChainChange(); err != nil {
					log.Log.E("tx_pool - doChainChange is error")
					pool.clearTxPending()
				}

			case NotFork:
				if err := pool.delBlockTxInPending(bl.LinkedNode.Block.HeadHash()); err != nil {
					log.Log.E("tx_pool - delBlockTxInPending is error")
				}

			default:
				log.Log.E("tx_pool - updateForkChain is error")
			}
			pool.mu.Unlock()

		case <-clearTx.C:
			pool.mu.Lock()

			pool.clearBlock()
			pool.clearTimeOutTx()

			pool.mu.Unlock()
		}
	}
}

func (pool *TxPoolImpl) AddLinkedNode(linkedNode *blockcache.BlockCacheNode, headNode *blockcache.BlockCacheNode) error {

	if linkedNode == nil || headNode == nil {
		return errors.New("parameter is nil")
	}

	r := &RecNode{
		LinkedNode: linkedNode,
		HeadNode:   headNode,
	}

	pool.chLinkedNode <- r

	return nil
}

func (pool *TxPoolImpl) AddTx(tx message.Message) error {
	pool.chTx <- tx
	return nil
}

func (pool *TxPoolImpl) PendingTxs(maxCnt int) (TxsList, error) {

	var pendingList TxsList

	pool.pendingTx.Range(func(key, value interface{}) bool {
		pendingList = append(pendingList, value.(*tx.Tx))

		return true
	})

	sort.Sort(pendingList)

	len := len(pendingList)
	if len >= maxCnt {
		len = maxCnt
	}

	return pendingList[:len], nil
}

func (pool *TxPoolImpl) ExistTxs(hash []byte, chainBlock *block.Block) (FRet, error) {

	var r FRet

	switch {
	case pool.existTxInPending(hash):
		r = FoundPending
	case pool.existTxInChain(hash, chainBlock):
		r = FoundChain
	default:
		r = NotFound

	}

	return r, nil
}

func (pool *TxPoolImpl) initBlockTx() {
	chain := pool.global.BlockChain()
	timeNow := time.Now().Unix()

	for i := chain.Length() - 1; i > 0; i-- {
		blk, err := chain.GetBlockByNumber(i)
		if err != nil {
			return
		}

		t := pool.slotToSec(blk.Head.Time)
		if timeNow-t <= filterTime {
			pool.addBlock(blk)
		}
	}

}

func (pool *TxPoolImpl) slotToSec(t int64) int64 {
	slot := consensus_common.Timestamp{Slot: t}
	return slot.ToUnixSec()
}

func (pool *TxPoolImpl) addBlock(linkedBlock *block.Block) error {

	if _, ok := pool.blockList.Load(linkedBlock.HeadHash()); ok {
		return nil
	}

	b := new(blockTx)

	b.setTime(pool.slotToSec(linkedBlock.Head.Time))
	b.addBlock(linkedBlock)

	pool.blockList.Store(linkedBlock.HeadHash(), b)

	return nil
}

func (pool *TxPoolImpl) parentHash(hash []byte) ([]byte, bool) {

	v, ok := pool.block(hash)
	if !ok {
		return nil, false
	}

	return v.ParentHash, true
}

func (pool *TxPoolImpl) block(hash []byte) (*blockTx, bool) {

	if v, ok := pool.blockList.Load(hash); ok {
		return v.(*blockTx), true
	}

	return nil, false
}

func (pool *TxPoolImpl) existTxInChain(txHash []byte, block *block.Block) bool {

	h := block.Head.Hash()
	t := pool.slotToSec(block.Head.Time)
	var ok bool

	for {
		ret := pool.existTxInBlock(txHash, h)
		if ret {
			return true
		}

		h, ok = pool.parentHash(h)
		if !ok {
			return false
		}

		if b, ok := pool.block(h); ok {
			if (t - b.time()) > filterTime {
				return false
			}
		}

	}

	return false
}

func (pool *TxPoolImpl) existTxInBlock(txHash []byte, blockHash []byte) bool {

	b, ok := pool.blockList.Load(blockHash)
	if !ok {
		return false
	}

	return b.(*blockTx).existTx(txHash)
}

func (pool *TxPoolImpl) clearBlock() {
	ft := pool.slotToSec(pool.chain.LinkedTree.Block.Head.Time) - filterTime

	pool.blockList.Range(func(key, value interface{}) bool {
		if value.(*blockTx).time() < ft {
			pool.blockList.Delete(key)
		}

		return true
	})

}

func (pool *TxPoolImpl) addTx(tx *tx.Tx) {

	h := tx.Hash()

	if !pool.existTxInChain(h, pool.forkChain.NewHead.Block) && !pool.existTxInPending(h) {
		pool.pendingTx.Store(h, tx)
	}

}

func (pool *TxPoolImpl) existTxInPending(hash []byte) bool {

	_, ok := pool.pendingTx.Load(hash)

	return ok
}

func (pool *TxPoolImpl) txTimeOut(tx *tx.Tx) bool {

	nTime := time.Now().Unix()
	txTime := tx.Time / 1e9
	exTime := tx.Expiration / 1e9

	if exTime <= nTime {
		return true
	}

	if nTime-txTime > expiration {
		return true
	}
	return false
}

func (pool *TxPoolImpl) clearTimeOutTx() {

	pool.pendingTx.Range(func(key, value interface{}) bool {

		if pool.txTimeOut(value.(*tx.Tx)) {
			pool.delTxInPending(value.(*tx.Tx).Hash())
		}

		return true
	})

}

func (pool *TxPoolImpl) delTxInPending(hash []byte) {
	pool.pendingTx.Delete(hash)
}

func (pool *TxPoolImpl) delBlockTxInPending(hash []byte) error {

	b, ok := pool.block(hash)
	if !ok {
		return nil
	}

	b.txMap.Range(func(key, value interface{}) bool {
		pool.pendingTx.Delete(key)
		return true
	})

	return nil
}

func (pool *TxPoolImpl) clearTxPending() {
	pool.pendingTx = new(sync.Map)
}

func (pool *TxPoolImpl) updatePending(blockHash []byte) error {

	b, ok := pool.block(blockHash)
	if !ok {
		return errors.New("updatePending is error")
	}

	b.txMap.Range(func(key, value interface{}) bool {

		pool.delTxInPending(key.([]byte))
		return true
	})

	return nil
}

func (pool *TxPoolImpl) updateForkChain(headNode *blockcache.BlockCacheNode) TFork {

	if pool.forkChain.NewHead == nil {
		pool.forkChain.NewHead = headNode
		return NotFork
	}

	if bytes.Equal(pool.forkChain.NewHead.Block.HeadHash(), headNode.Block.Head.ParentHash) {
		pool.forkChain.NewHead = headNode
		return NotFork
	}

	pool.forkChain.OldHead, pool.forkChain.NewHead = pool.forkChain.NewHead, headNode

	hash, ok := pool.fundForkBlockHash(pool.forkChain.NewHead.Block.HeadHash(), pool.forkChain.OldHead.Block.HeadHash())
	if !ok {
		return ForkError
	}

	pool.forkChain.ForkBlockHash = hash

	return Fork
}

func (pool *TxPoolImpl) fundForkBlockHash(newHash []byte, oldHash []byte) ([]byte, bool) {
	n := newHash
	o := oldHash

	if bytes.Equal(n, o) {
		return n, true
	}

	for {

		forkHash, ok := pool.fundBlockInChain(n, o)
		if ok {
			return forkHash, true
		}

		b, ok := pool.block(n)
		if !ok {
			bb, err := pool.chain.Find(n)
			if err != nil {
				log.Log.E("tx_pool - fundForkBlockHash FindBlock error")
				return nil, false
			}

			if err = pool.addBlock(bb.Block); err != nil {
				log.Log.E("tx_pool - fundForkBlockHash addBlock error")
				return nil, false
			}

			b, ok = pool.block(n)
			if !ok {
				log.Log.E("tx_pool - fundForkBlockHash block error")
				return nil, false
			}
		}

		n = b.ParentHash

		if bytes.Equal(pool.chain.LinkedTree.Block.Head.ParentHash, b.ParentHash) {
			return nil, false
		}

	}

	return nil, false
}

func (pool *TxPoolImpl) fundBlockInChain(hash []byte, chainHead []byte) ([]byte, bool) {
	h := hash
	c := chainHead

	if bytes.Equal(h, c) {
		return h, true
	}

	for {
		b, ok := pool.block(c)
		if !ok {
			return nil, false
		}

		if bytes.Equal(b.ParentHash, h) {
			return h, true
		}

		c = b.ParentHash

	}

	return nil, false
}

func (pool *TxPoolImpl) doChainChange() error {

	n := pool.forkChain.NewHead.Block.HeadHash()
	o := pool.forkChain.OldHead.Block.HeadHash()
	f := pool.forkChain.ForkBlockHash

	//Reply to txs
	for {
		b, err := pool.chain.Find(o)
		if err != nil {
			return err
		}

		for _, v := range b.Block.Txs {
			pool.addTx(v)
		}

		if bytes.Equal(b.Block.Head.ParentHash, f) {
			break
		}

		o = b.Block.Head.ParentHash
	}

	//Duplicate txs
	for {
		b, ok := pool.block(n)
		if !ok {
			return errors.New("doForkChain is error")
		}

		b.txMap.Range(func(key, value interface{}) bool {
			pool.delTxInPending(key.([]byte))
			return true
		})

		if bytes.Equal(b.ParentHash, f) {
			break
		}

		n = b.ParentHash
	}

	return nil
}
