package txpool

import (
	"sync"
	"time"

	"errors"

	"runtime"

	"fmt"

	"unsafe"

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
		ret := pool.verifyTx(&t)
		if ret != Success {
			continue
		}
		ret = pool.addTx(&t)
		if ret != Success {
			continue
		}
		metricsReceivedTxCount.Add(1, map[string]string{"from": "p2p"})
		pool.p2pService.Broadcast(v.Data(), p2p.PublishTx, p2p.NormalMessage)
	}
}

// AddLinkedNode add the findBlock
func (pool *TxPImpl) AddLinkedNode(linkedNode *blockcache.BlockCacheNode, newHead *blockcache.BlockCacheNode) error {
	err := pool.addBlock(linkedNode.Block)
	if err != nil {
		return fmt.Errorf("failed to add findBlock: %v", err)
	}
	typeOfFork := pool.updateForkChain(newHead)
	switch typeOfFork {
	case forkBCN:
		pool.mu.Lock()
		defer pool.mu.Unlock()
		pool.doChainChangeByForkBCN()
	case noForkBCN:
		pool.mu.Lock()
		defer pool.mu.Unlock()
		pool.doChainChangeByTimeout()
	case sameHead:
	default:
		return errors.New("failed to tFort")
	}
	return nil
}

// AddTx add the transaction
func (pool *TxPImpl) AddTx(t *tx.Tx) TAddTx {
	ret := pool.verifyTx(t)
	if ret != Success {
		return ret
	}
	ret = pool.addTx(t)
	if ret != Success {
		return ret
	}
	pool.p2pService.Broadcast(t.Encode(), p2p.PublishTx, p2p.NormalMessage)
	metricsReceivedTxCount.Add(1, map[string]string{"from": "rpc"})
	return ret
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
	ilog.Errorf("pendingTx.Size(): %v", pool.pendingTx.Size())
	metricsTxPoolSize.Set(float64(pool.pendingTx.Size()), nil)
	return pool.pendingTx.Iter(), pool.forkChain.NewHead
}

// ExistTxs determine if the transaction exists
func (pool *TxPImpl) ExistTxs(hash []byte, chainBlock *block.Block) FRet {
	var r FRet
	switch {
	//case pool.existTxInPending(hash):
	//	r = FoundPending
	case pool.existTxInChain2(hash, chainBlock):
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
		if pool.slotToNSec(blk.Head.Time) < filterLimit {
			break
		}
		pool.addBlock(blk)
	}
}

func (pool *TxPImpl) verifyTx(t *tx.Tx) TAddTx {
	if pool.pendingTx.Size() > maxCacheTxs {
		return CacheFullError
	}
	if t.GasPrice <= 0 {
		return GasPriceError
	}
	//if pool.TxTimeOut(t) {
	//	return TimeError
	//}
	if err := t.VerifySelf(); err != nil {
		return VerifyError
	}
	return Success
}

func (pool *TxPImpl) slotToNSec(t int64) int64 {
	return common.SlotLength * t * int64(time.Second)
}

func (pool *TxPImpl) addBlock(blk *block.Block) error {
	if blk == nil {
		return errors.New("failed to linkedBlock")
	}
	if _, ok := pool.blockList.Load(string(blk.HeadHash())); ok {
		return nil
	}
	parentBlockTx, _ := pool.blockList.Load(string(blk.Head.ParentHash))
	pool.blockList.Store(string(blk.HeadHash()), pool.newBlockTx(blk, parentBlockTx))
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

func (pool *TxPImpl) existTxInChain1(txHash []byte, block *block.Block) bool {
	if block == nil {
		return false
	}
	h := block.HeadHash()
	filterLimit := pool.slotToNSec(block.Head.Time) - filterTime
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
		if b, ok := pool.findBlock(h); ok {
			if b.time < filterLimit {
				return false
			}
		}
	}
}

func (pool *TxPImpl) existTxInChain2(txHash []byte, blk *block.Block) bool {
	if blk == nil {
		return false
	}
	b, ok := pool.findBlock(blk.HeadHash())
	if !ok {
		return false
	}
	_, ok = b.chainMap.Load(string(txHash))
	return ok
}

func (pool *TxPImpl) existTxInBlock(txHash []byte, blockHash []byte) bool {
	b, ok := pool.blockList.Load(string(blockHash))
	if !ok {
		return false
	}
	return b.(*blockTx).existTx(txHash)
}

func (pool *TxPImpl) clearBlock() {
	filterLimit := pool.slotToNSec(pool.blockCache.LinkedRoot().Block.Head.Time) - filterTime
	pool.blockList.Range(func(key, value interface{}) bool {
		if value.(*blockTx).time < filterLimit {
			pool.blockList.Delete(key)
		}
		return true
	})
	listnum := 0
	totalBytes := 0
	pool.blockList.Range(func(key, value interface{}) bool {
		listnum++
		chainmaplength := 0
		var keysize, valuesize uintptr
		value.(*blockTx).chainMap.Range(func(key, value interface{}) bool {
			chainmaplength++
			keysize = unsafe.Sizeof(key)
			valuesize = unsafe.Sizeof(value)
			return true
		})
		ilog.Errorf("keysize: %v, valuesize: %v, chainmaplength: %v", keysize, valuesize, chainmaplength)
		totalBytes += 8 * chainmaplength * int(keysize+valuesize)
		return true
	})
	ilog.Errorf("totalBytes: %v", totalBytes)
}

func (pool *TxPImpl) addTx(tx *tx.Tx) TAddTx {
	//h := tx.Hash()
	//if pool.existTxInChain2(h, pool.forkChain.NewHead.Block) {
	//	return DupError
	//}
	//if pool.existTxInPending(h) {
	//	return DupError
	//}
	pool.pendingTx.Add(tx)
	return Success
}

func (pool *TxPImpl) existTxInPending(hash []byte) bool {
	return pool.pendingTx.Get(hash) != nil
}

// TxTimeOut time to verify the tx
func (pool *TxPImpl) TxTimeOut(tx *tx.Tx) bool {
	return false
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
		pool.forkChain.ForkBCN = bcn
		return forkBCN
	}
	pool.forkChain.ForkBCN = nil
	return noForkBCN
}

func (pool *TxPImpl) findForkBCN(newHead *blockcache.BlockCacheNode, oldHead *blockcache.BlockCacheNode) (*blockcache.BlockCacheNode, bool) {
	for {
		for oldHead != nil && oldHead.Number > newHead.Number {
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
		if oldHead == nil || oldHead == forkBCN || pool.slotToNSec(oldHead.Block.Head.Time) < filterLimit {
			break
		}
		for _, t := range oldHead.Block.Txs {
			pool.pendingTx.Add(t)
		}
		oldHead = oldHead.Parent
	}

	//del txs
	for {
		if newHead == nil || newHead == forkBCN || pool.slotToNSec(newHead.Block.Head.Time) < filterLimit {
			break
		}
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

func (pool *TxPImpl) testPendingTxsNum() int64 {
	return int64(pool.pendingTx.Size())
}

func (pool *TxPImpl) testBlockListNum() int64 {
	var r int64
	pool.blockList.Range(func(key, value interface{}) bool {
		r++
		return true
	})
	return r
}
