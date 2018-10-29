package txpool

import (
	"sync"
	"time"

	"errors"

	"runtime"

	"fmt"

	"github.com/iost-official/go-iost/common"
	"github.com/iost-official/go-iost/core/block"
	"github.com/iost-official/go-iost/core/blockcache"
	"github.com/iost-official/go-iost/core/global"
	"github.com/iost-official/go-iost/core/tx"
	"github.com/iost-official/go-iost/ilog"
	"github.com/iost-official/go-iost/p2p"
)

// TxPImpl defines all the API of txpool package.
type TxPImpl struct {
	global           global.BaseVariable
	blockCache       blockcache.BlockCache
	p2pService       p2p.Service
	forkChain        *forkChain
	blockList        *sync.Map
	pendingTx        *SortedTxMap
	mu               sync.RWMutex
	chP2PTx          chan p2p.IncomingMessage
	quitGenerateMode chan struct{}
	quitCh           chan struct{}
}

// NewTxPoolImpl returns a default TxPImpl instance.
func NewTxPoolImpl(global global.BaseVariable, blockCache blockcache.BlockCache, p2pService p2p.Service) (*TxPImpl, error) {
	p := &TxPImpl{
		global:           global,
		blockCache:       blockCache,
		p2pService:       p2pService,
		forkChain:        new(forkChain),
		blockList:        new(sync.Map),
		pendingTx:        NewSortedTxMap(),
		chP2PTx:          p2pService.Register("txpool message", p2p.PublishTx),
		quitGenerateMode: make(chan struct{}),
		quitCh:           make(chan struct{}),
	}
	p.forkChain.NewHead = blockCache.Head()
	close(p.quitGenerateMode)
	return p, nil
}

// Start starts the jobs.
func (pool *TxPImpl) Start() error {
	go pool.loop()
	return nil
}

// Stop stops all the jobs.
func (pool *TxPImpl) Stop() {
	close(pool.quitCh)
}

func (pool *TxPImpl) loop() {
	for {
		if pool.global.Mode() != global.ModeInit {
			break
		}
		time.Sleep(time.Second)
	}
	pool.initBlockTx()
	workerCnt := (runtime.NumCPU() + 1) / 2
	if workerCnt == 0 {
		workerCnt = 1
	}
	for i := 0; i < workerCnt; i++ {
		go pool.verifyWorkers()
	}
	clearTx := time.NewTicker(clearInterval)
	defer clearTx.Stop()
	for {
		select {
		case <-clearTx.C:
			metricsTxPoolSize.Set(float64(pool.pendingTx.Size()), nil)
			pool.mu.Lock()
			pool.clearBlock()
			pool.clearTimeOutTx()
			pool.mu.Unlock()
		case <-pool.quitCh:
			return
		}
	}
}

// Lock lock the txpool
func (pool *TxPImpl) Lock() {
	pool.mu.Lock()
	pool.quitGenerateMode = make(chan struct{})
}

// Release release the txpool
func (pool *TxPImpl) Release() {
	pool.mu.Unlock()
	close(pool.quitGenerateMode)
}

func (pool *TxPImpl) verifyWorkers() {
	for v := range pool.chP2PTx {
		select {
		case <-pool.quitGenerateMode:
		}
		var t tx.Tx
		err := t.Decode(v.Data())
		if err != nil {
			continue
		}
		pool.mu.Lock()
		ret := pool.verifyDuplicate(&t)
		if ret != Success {
			pool.mu.Unlock()
			continue
		}
		ret = pool.verifyTx(&t)
		if ret != Success {
			pool.mu.Unlock()
			continue
		}
		pool.pendingTx.Add(&t)
		pool.mu.Unlock()
		metricsReceivedTxCount.Add(1, map[string]string{"from": "p2p"})
		pool.p2pService.Broadcast(v.Data(), p2p.PublishTx, p2p.NormalMessage, true)
	}
}

// CheckTxs check txs
func (pool *TxPImpl) CheckTxs(txs []*tx.Tx, chainBlock *block.Block) (*tx.Tx, error) {
	rm, err := pool.createTxMapToChain(chainBlock)
	if err != nil {
		return nil, err
	}
	dtm := new(sync.Map)
	for _, v := range txs {
		trh := string(v.Hash())
		if _, ok := rm.Load(trh); ok {
			return v, errors.New("duplicate tx in chain")
		}
		if _, ok := dtm.Load(trh); ok {
			return v, errors.New("duplicate tx in txs")
		}
		dtm.Store(trh, nil)
		if ok := pool.existTxInPending([]byte(trh)); !ok {
			if pool.verifyTx(v) != nil {
				return v, errors.New("failed to verify")
			}
		}
	}
	return nil, nil
}

func (pool *TxPImpl) createTxMapToChain(chainBlock *block.Block) (*sync.Map, error) {
	if chainBlock == nil {
		return nil, errors.New("chainBlock is nil")
	}
	rm := new(sync.Map)
	h := chainBlock.HeadHash()
	t := slotToNSec(chainBlock.Head.Time)
	var ok bool
	for {
		ret := pool.createTxMapToBlock(rm, h)
		if !ret {
			return nil, errors.New("failed to create tx map")
		}
		h, ok = pool.parentHash(h)
		if !ok {
			return nil, errors.New("failed to get parent chainBlock")
		}
		if b, ok := pool.findBlock(h); ok {
			if (t - b.time) > filterTime {
				return rm, nil
			}
		}
	}
}

func (pool *TxPImpl) createTxMapToBlock(tm *sync.Map, blockHash []byte) bool {
	b, ok := pool.blockList.Load(string(blockHash))
	if !ok {
		return false
	}
	b.(*blockTx).txMap.Range(func(key, value interface{}) bool {
		tm.Store(key.(string), nil)
		return true
	})
	return true
}

// AddLinkedNode add the findBlock
func (pool *TxPImpl) AddLinkedNode(linkedNode *blockcache.BlockCacheNode) error {
	pool.mu.Lock()
	defer pool.mu.Unlock()
	err := pool.addBlock(linkedNode.Block)
	if err != nil {
		return fmt.Errorf("failed to add findBlock: %v", err)
	}
	var newHead *blockcache.BlockCacheNode
	h := pool.blockCache.Head()
	if linkedNode.Number > h.Number {
		newHead = linkedNode
	} else {
		newHead = h
	}
	typeOfFork := pool.updateForkChain(newHead)
	switch typeOfFork {
	case forkBCN:
		pool.doChainChangeByForkBCN()
	case noForkBCN:
		pool.doChainChangeByTimeout()
	case sameHead:
	default:
		return errors.New("failed to tFort")
	}
	return nil
}

// AddTx add the transaction
func (pool *TxPImpl) AddTx(t *tx.Tx) TAddTx {
	//ret := pool.verifyDuplicate(t)
	//if ret != Success {
	//	return ret
	//}
	//ret = pool.verifyTx(t)
	//if ret != Success {
	//	return ret
	//}
	pool.pendingTx.Add(t)
	pool.p2pService.Broadcast(t.Encode(), p2p.PublishTx, p2p.NormalMessage)
	metricsReceivedTxCount.Add(1, map[string]string{"from": "rpc"})
	//return ret
	return Success
}

// DelTx del the transaction
func (pool *TxPImpl) DelTx(hash []byte) error {
	pool.pendingTx.Del(hash)
	return nil
}

// DelTxList deletes the tx list in txpool.
func (pool *TxPImpl) DelTxList(delList []*tx.Tx) {
	for _, t := range delList {
		pool.pendingTx.Del(t.Hash())
	}
}

// TxIterator ...
func (pool *TxPImpl) TxIterator() (*Iterator, *blockcache.BlockCacheNode) {
	metricsTxPoolSize.Set(float64(pool.pendingTx.Size()), nil)
	return pool.pendingTx.Iter(), pool.forkChain.NewHead
}

// ExistTxs determine if the transaction exists
func (pool *TxPImpl) ExistTxs(hash []byte, chainBlock *block.Block) FRet {
	var r FRet
	switch {
	case pool.existTxInPending(hash):
		r = FoundPending
	case pool.ExistTxInChain(hash, chainBlock):
		r = FoundChain
	default:
		r = NotFound
	}
	return r
}

func (pool *TxPImpl) initBlockTx() {
	filterLimit := time.Now().UnixNano() - filterTime
	for i := pool.global.BlockChain().Length() - 1; i > 0; i-- {
		blk, err := pool.global.BlockChain().GetBlockByNumber(i)
		if err != nil {
			break
		}
		if slotToNSec(blk.Head.Time) < filterLimit {
			break
		}
		pool.addBlock(blk)
	}
}

func (pool *TxPImpl) verifyTx(t *tx.Tx) error {
	if pool.pendingTx.Size() > maxCacheTxs {
		return fmt.Errorf("CacheFullError. Pending tx size is %d. Max cache is %d", pool.pendingTx.Size(), maxCacheTxs)
	}
	if t.GasPrice <= 0 {
		return fmt.Errorf("GasPriceError. gas price %d", t.GasPrice)
	}
	if pool.TxTimeOut(t) {
		return fmt.Errorf("TimeError")
	}
	if err := t.VerifySelf(); err != nil {
		return fmt.Errorf("VerifyError %v", err)
	}
	return nil
}

func slotToNSec(t int64) int64 {
	return common.SlotLength * t * int64(time.Second)
}

func (pool *TxPImpl) addBlock(blk *block.Block) error {
	if blk == nil {
		return errors.New("failed to linkedBlock")
	}
	if _, ok := pool.blockList.Load(string(blk.HeadHash())); ok {
		return nil
	}
	pool.blockList.Store(string(blk.HeadHash()), pool.newBlockTx(blk))
	return nil
}

func (pool *TxPImpl) parentHash(hash []byte) ([]byte, bool) {
	v, ok := pool.findBlock(hash)
	if !ok {
		return nil, false
	}
	return v.ParentHash, true
}

func (pool *TxPImpl) findBlock(hash []byte) (*blockTx, bool) {
	if v, ok := pool.blockList.Load(string(hash)); ok {
		return v.(*blockTx), true
	}
	return nil, false
}

func (pool *TxPImpl) ExistTxInChain(txHash []byte, block *block.Block) bool {
	if block == nil {
		return false
	}
	h := block.HeadHash()
	filterLimit := slotToNSec(block.Head.Time) - filterTime
	var ok bool
	for {
		ret := pool.existTxInBlock(txHash, h)
		if ret {
			ilog.Infof("find in chain")
			bcn, err := pool.blockCache.Find(h)
			if err == nil {
				ilog.Infof("find in blockcache")
				ilog.Info(bcn.Number, " ", bcn.Witness)
			} else {
				blk, err := pool.global.BlockChain().GetBlockByHash(h)
				if err == nil {
					ilog.Infof("find in blockchain")
					ilog.Info(blk.Head.Number, " ", blk.Head.Witness)
				}
			}
			return true
		}
		h, ok = pool.parentHash(h)
		if !ok {
			return false
		}
		if b, ok := pool.findBlock(h); ok {
			if b.time < filterLimit {
				return false
			}
		}
	}
}

func (pool *TxPImpl) existTxInBlock(txHash []byte, blockHash []byte) bool {
	b, ok := pool.blockList.Load(string(blockHash))
	if !ok {
		return false
	}
	return b.(*blockTx).existTx(txHash)
}

func (pool *TxPImpl) clearBlock() {
	filterLimit := slotToNSec(pool.blockCache.LinkedRoot().Block.Head.Time) - filterTime
	pool.blockList.Range(func(key, value interface{}) bool {
		if value.(*blockTx).time < filterLimit {
			pool.blockList.Delete(key)
		}
		return true
	})
}

func (pool *TxPImpl) verifyDuplicate(t *tx.Tx) error {
	if pool.existTxInPending(t.Hash()) {
		return fmt.Errorf("DupError. tx exists in pending")
	}
	if pool.ExistTxInChain(t.Hash(), pool.forkChain.NewHead.Block) {
		ilog.Infof("tx found in chain")
		return fmt.Errorf("DupError. tx exists in chain")
	}
	return nil
}

func (pool *TxPImpl) existTxInPending(hash []byte) bool {
	return pool.pendingTx.Get(hash) != nil
}

// TxTimeOut time to verify the tx
func (pool *TxPImpl) TxTimeOut(tx *tx.Tx) bool {
	currentTime := time.Now().UnixNano()
	if tx.Time > currentTime {
		return true
	}
	if tx.Expiration <= currentTime {
		return true
	}
	if currentTime-tx.Time > Expiration {
		return true
	}
	return false
}

func (pool *TxPImpl) clearTimeOutTx() {
	iter := pool.pendingTx.Iter()
	t, ok := iter.Next()
	for ok {
		if pool.TxTimeOut(t) {
			pool.pendingTx.Del(t.Hash())
		}
		t, ok = iter.Next()
	}
}

func (pool *TxPImpl) updateForkChain(newHead *blockcache.BlockCacheNode) tFork {
	if pool.forkChain.NewHead == newHead {
		return sameHead
	}
	pool.forkChain.OldHead, pool.forkChain.NewHead = pool.forkChain.NewHead, newHead
	bcn, ok := pool.findForkBCN(pool.forkChain.NewHead, pool.forkChain.OldHead)
	if ok {
		ilog.Infof("find forkbcn:%v, %v", bcn.Number, bcn.Witness)
		pool.forkChain.ForkBCN = bcn
		return forkBCN
	}
	pool.forkChain.ForkBCN = nil
	return noForkBCN
}

func (pool *TxPImpl) findForkBCN(newHead *blockcache.BlockCacheNode, oldHead *blockcache.BlockCacheNode) (*blockcache.BlockCacheNode, bool) {
	for {
		for oldHead != nil && oldHead.Head.Number > newHead.Head.Number {
			oldHead = oldHead.Parent
		}
		if oldHead == nil {
			return nil, false
		}
		if oldHead == newHead {
			return oldHead, true
		}
		newHead = newHead.Parent
		if newHead == nil {
			return nil, false
		}
	}
}

func (pool *TxPImpl) doChainChangeByForkBCN() {
	newHead := pool.forkChain.NewHead
	oldHead := pool.forkChain.OldHead
	forkBCN := pool.forkChain.ForkBCN
	//add txs
	filterLimit := time.Now().UnixNano() - filterTime
	for {
		if oldHead == nil || oldHead == forkBCN || slotToNSec(oldHead.Block.Head.Time) < filterLimit {
			break
		}
		for _, t := range oldHead.Block.Txs {
			pool.pendingTx.Add(t)
		}
		oldHead = oldHead.Parent
	}

	//del txs
	for {
		if newHead == nil || newHead == forkBCN || slotToNSec(newHead.Block.Head.Time) < filterLimit {
			break
		}
		ilog.Infof("delete tx from block: %v, %v", newHead.Block.Head.Number, newHead.Block.Head.Witness)
		for _, t := range newHead.Block.Txs {
			pool.DelTx(t.Hash())
		}
		newHead = newHead.Parent
	}
}

func (pool *TxPImpl) doChainChangeByTimeout() {
	newHead := pool.forkChain.NewHead
	oldHead := pool.forkChain.OldHead
	filterLimit := time.Now().UnixNano() - filterTime
	ob, ok := pool.findBlock(oldHead.Block.HeadHash())
	if ok {
		for {
			if ob.time < filterLimit {
				break
			}
			ob.txMap.Range(func(k, v interface{}) bool {
				pool.pendingTx.Add(v.(*tx.Tx))
				return true
			})
			ob, ok = pool.findBlock(ob.ParentHash)
			if !ok {
				break
			}
		}
	}
	nb, ok := pool.findBlock(newHead.Block.HeadHash())
	if ok {
		for {
			if nb.time < filterLimit {
				break
			}
			nb.txMap.Range(func(k, v interface{}) bool {
				pool.DelTx(v.(*tx.Tx).Hash())
				return true
			})
			nb, ok = pool.findBlock(nb.ParentHash)
			if !ok {
				break
			}
		}
	}
}
