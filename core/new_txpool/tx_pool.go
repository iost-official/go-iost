package txpool

import (
	"sort"
	"sync"
	"time"

	"bytes"
	"errors"
	"os"

	"runtime"

	"github.com/iost-official/Go-IOS-Protocol/common"
	"github.com/iost-official/Go-IOS-Protocol/core/global"
	"github.com/iost-official/Go-IOS-Protocol/core/new_block"
	"github.com/iost-official/Go-IOS-Protocol/core/new_blockcache"
	"github.com/iost-official/Go-IOS-Protocol/core/tx"
	"github.com/iost-official/Go-IOS-Protocol/ilog"
	"github.com/iost-official/Go-IOS-Protocol/p2p"
	"github.com/prometheus/client_golang/prometheus"
)

func init() {
	prometheus.MustRegister(receivedTransactionCount)
}

type TxPoolImpl struct {
	chP2PTx chan p2p.IncomingMessage
	chTx    chan *tx.Tx

	chLinkedNode chan *RecNode

	global     global.BaseVariable
	blockCache blockcache.BlockCache
	p2pService p2p.Service

	forkChain *ForkChain
	blockList *sync.Map
	pendingTx *sync.Map

	mu sync.RWMutex
}

func NewTxPoolImpl(global global.BaseVariable, blockCache blockcache.BlockCache, p2ps p2p.Service) (*TxPoolImpl, error) {
	p := &TxPoolImpl{
		blockCache:   blockCache,
		chLinkedNode: make(chan *RecNode, 100),
		chTx:         make(chan *tx.Tx, 10000),
		forkChain:    new(ForkChain),
		blockList:    new(sync.Map),
		pendingTx:    new(sync.Map),
		global:       global,
	}
	p.p2pService = p2ps
	p.chP2PTx = p.p2pService.Register("TxPool message", p2p.PublishTxRequest)

	return p, nil
}

func (pool *TxPoolImpl) Start() {
	ilog.Info("TxPoolImpl Start")
	go pool.loop()
}

func (pool *TxPoolImpl) Stop() {
	ilog.Info("TxPoolImpl Stop")
	close(pool.chP2PTx)
	close(pool.chLinkedNode)
}

func (pool *TxPoolImpl) loop() {

	pool.initBlockTx()

	workerCnt := (runtime.NumCPU() + 1) / 2
	if workerCnt == 0 {
		workerCnt = 1
	}

	for i := 0; i < workerCnt; i++ {
		go pool.verifyWorkers(pool.chP2PTx, pool.chTx)
	}

	clearTx := time.NewTicker(clearInterval)
	defer clearTx.Stop()

	for {
		select {
		case tr, ok := <-pool.chTx:
			if !ok {
				ilog.Error("failed to chTx")
				os.Exit(1)
			}

			if ret := pool.addTx(tr); ret == Success {
				pool.p2pService.Broadcast(tr.Encode(), p2p.PublishTxRequest, p2p.UrgentMessage)
				receivedTransactionCount.Inc()
			}

		case bl, ok := <-pool.chLinkedNode:
			if !ok {
				ilog.Error("failed to ch linked node")
				os.Exit(1)
			}

			if pool.addBlock(bl.LinkedNode.Block) != nil {
				continue
			}

			pool.mu.Lock()

			tFort := pool.updateForkChain(bl.HeadNode)
			switch tFort {
			case ForkError:
				ilog.Error("failed to update fork chain")
				pool.clearTxPending()

			case Fork:
				if err := pool.doChainChange(); err != nil {
					ilog.Error("failed to chain change")
					pool.clearTxPending()
				}

			case NotFork:

				if err := pool.delBlockTxInPending(bl.LinkedNode.Block.HeadHash()); err != nil {
					ilog.Error("failed to del block tx")
				}

			default:
				ilog.Error("failed to tFort")
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

func (pool *TxPoolImpl) verifyWorkers(p2pCh chan p2p.IncomingMessage, tCn chan *tx.Tx) {

	for v := range p2pCh {

		var t tx.Tx
		err := t.Decode(v.Data())
		if err != nil {
			continue
		}

		if r := pool.verifyTx(&t); r == Success {
			tCn <- &t
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

func (pool *TxPoolImpl) AddTx(t *tx.Tx) TAddTx {

	var r TAddTx

	if r = pool.verifyTx(t); r != Success {
		return r
	}

	if r = pool.addTx(t); r == Success {
		pool.p2pService.Broadcast(t.Encode(), p2p.PublishTxRequest, p2p.UrgentMessage)
		receivedTransactionCount.Inc()
	}

	return r
}

func (pool *TxPoolImpl) PendingTxs(maxCnt int) (TxsList, error) {

	var pendingList TxsList

	pool.pendingTx.Range(func(key, value interface{}) bool {
		pendingList = append(pendingList, value.(*tx.Tx))

		return true
	})

	sort.Sort(pendingList)

	l := len(pendingList)
	if l >= maxCnt {
		l = maxCnt
	}

	return pendingList[:l], nil
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

func (pool *TxPoolImpl) verifyTx(t *tx.Tx) TAddTx {

	if pool.txTimeOut(t) {
		return TimeError
	}

	if err := t.VerifySelf(); err != nil {
		return VerifyError
	}

	return Success
}

func (pool *TxPoolImpl) slotToSec(t int64) int64 {
	slot := common.Timestamp{Slot: t}
	return slot.ToUnixSec()
}

func (pool *TxPoolImpl) addBlock(linkedBlock *block.Block) error {

	if linkedBlock == nil {
		return errors.New("failed to linkedBlock")
	}

	h := linkedBlock.HeadHash()

	if _, ok := pool.blockList.Load(string(h)); ok {
		return nil
	}

	b := newBlockTx()

	b.setTime(pool.slotToSec(linkedBlock.Head.Time))
	b.addBlock(linkedBlock)

	pool.blockList.Store(string(h), b)

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

	if v, ok := pool.blockList.Load(string(hash)); ok {
		return v.(*blockTx), true
	}

	return nil, false
}

func (pool *TxPoolImpl) existTxInChain(txHash []byte, block *block.Block) bool {

	if block == nil {
		return false
	}

	h := block.HeadHash()

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

	b, ok := pool.blockList.Load(string(blockHash))
	if !ok {
		return false
	}

	return b.(*blockTx).existTx(txHash)
}

func (pool *TxPoolImpl) clearBlock() {
	ft := pool.slotToSec(pool.blockCache.LinkedRoot().Block.Head.Time) - filterTime

	pool.blockList.Range(func(key, value interface{}) bool {
		if value.(*blockTx).time() < ft {
			pool.blockList.Delete(key.(string))
		}

		return true
	})

}

func (pool *TxPoolImpl) addTx(tx *tx.Tx) TAddTx {

	pool.mu.Lock()
	defer pool.mu.Unlock()

	h := tx.Hash()

	if pool.forkChain.NewHead != nil {
		if pool.existTxInChain(h, pool.forkChain.NewHead.Block) {
			return DupError
		}
	}

	if pool.existTxInPending(h) {
		return DupError
	} else {
		pool.pendingTx.Store(string(h), tx)
	}

	return Success
}

func (pool *TxPoolImpl) existTxInPending(hash []byte) bool {

	_, ok := pool.pendingTx.Load(string(hash))

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
	pool.pendingTx.Delete(string(hash))
}

func (pool *TxPoolImpl) delBlockTxInPending(hash []byte) error {

	b, ok := pool.block(hash)
	if !ok {
		return nil
	}

	b.txMap.Range(func(key, value interface{}) bool {
		pool.pendingTx.Delete(key.(string))
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

	nh := pool.forkChain.NewHead.Block.HeadHash()

	if bytes.Equal(nh, headNode.Block.Head.ParentHash) {
		pool.forkChain.NewHead = headNode
		return NotFork
	}

	pool.forkChain.OldHead, pool.forkChain.NewHead = pool.forkChain.NewHead, headNode

	nh = pool.forkChain.NewHead.Block.HeadHash()
	on := pool.forkChain.OldHead.Block.HeadHash()

	hash, ok := pool.fundForkBlockHash(nh, on)
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
			bb, err := pool.blockCache.Find(n)
			if err != nil {
				ilog.Error("failed to find block ,err = ", err)
				return nil, false
			}

			if err = pool.addBlock(bb.Block); err != nil {
				ilog.Error("failed to add block, err = ", err)
				return nil, false
			}

			b, ok = pool.block(n)
			if !ok {
				ilog.Error("failed to get block ,err = ", err)
				return nil, false
			}
		}

		n = b.ParentHash

		if bytes.Equal(pool.blockCache.LinkedRoot().Block.Head.ParentHash, b.ParentHash) {
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
		//fmt.Println("blockCache.Find:", o)
		b, err := pool.blockCache.Find(o)
		if err != nil {
			return err
		}

		for _, v := range b.Block.Txs {
			if b := pool.existTxInPending(v.Hash()); !b {
				pool.pendingTx.Store(string(v.Hash()), v)
			}
		}

		if bytes.Equal(b.Block.Head.ParentHash, f) {
			break
		}

		o = b.Block.Head.ParentHash
	}

	//Check duplicate txs
	for {
		b, ok := pool.block(n)
		if !ok {
			return errors.New("doForkChain is error")
		}

		b.txMap.Range(func(key, value interface{}) bool {
			pool.delTxInPending([]byte(key.(string)))
			return true
		})

		if bytes.Equal(b.ParentHash, f) {
			break
		}

		n = b.ParentHash
	}

	return nil
}

func (pool *TxPoolImpl) testPendingTxsNum() int64 {
	var r int64 = 0

	pool.pendingTx.Range(func(key, value interface{}) bool {
		r++
		return true
	})

	return r
}

func (pool *TxPoolImpl) testBlockListNum() int64 {
	var r int64 = 0

	pool.blockList.Range(func(key, value interface{}) bool {
		r++
		//fmt.Println("blockList hash:", []byte(key.(string)))
		return true
	})

	return r
}
