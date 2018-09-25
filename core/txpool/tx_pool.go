package txpool

import (
	"sync"
	"time"

	"errors"

	"runtime"

	"github.com/iost-official/Go-IOS-Protocol/common"
	"github.com/iost-official/Go-IOS-Protocol/core/block"
	"github.com/iost-official/Go-IOS-Protocol/core/blockcache"
	"github.com/iost-official/Go-IOS-Protocol/core/global"
	"github.com/iost-official/Go-IOS-Protocol/core/tx"
	"github.com/iost-official/Go-IOS-Protocol/ilog"
	"github.com/iost-official/Go-IOS-Protocol/p2p"
)

// TxPImpl defines all the API of txpool package.
type TxPImpl struct {
	chP2PTx chan p2p.IncomingMessage
	chTx    chan *tx.Tx

	global     global.BaseVariable
	blockCache blockcache.BlockCache
	p2pService p2p.Service

	forkChain *forkChain
	blockList *sync.Map
	// pendingTx *sync.Map
	pendingTx *SortedTxMap

	mu               sync.RWMutex
	quitGenerateMode chan struct{}
	quitCh           chan struct{}
}

// NewTxPoolImpl returns a default TxPImpl instance.
func NewTxPoolImpl(global global.BaseVariable, blockCache blockcache.BlockCache, p2ps p2p.Service) (*TxPImpl, error) {
	p := &TxPImpl{
		blockCache:       blockCache,
		chTx:             make(chan *tx.Tx, 102400),
		forkChain:        new(forkChain),
		blockList:        new(sync.Map),
		pendingTx:        NewSortedTxMap(),
		global:           global,
		p2pService:       p2ps,
		chP2PTx:          p2ps.Register("TxPool message", p2p.PublishTxRequest),
		quitGenerateMode: make(chan struct{}),
		quitCh:           make(chan struct{}),
	}
	p.forkChain.NewHead = blockCache.Head()
	if p.forkChain.NewHead == nil {
		return nil, errors.New("failed to head")
	}
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
	ilog.Infof("TxPImpl Stop")
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
		go pool.verifyWorkers(pool.chP2PTx, pool.chTx)
	}

	clearTx := time.NewTicker(clearInterval)
	defer clearTx.Stop()

	for {
		select {
		case tr := <-pool.chTx:
			metricsReceivedTxCount.Add(1, map[string]string{"from": "p2p"})

			if ret := pool.addTx(tr); ret == Success {
				pool.p2pService.Broadcast(tr.Encode(), p2p.PublishTxRequest, p2p.NormalMessage)
			}

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

func (pool *TxPImpl) verifyWorkers(p2pCh chan p2p.IncomingMessage, tCn chan *tx.Tx) {
	for v := range p2pCh {
		select {
		case <-pool.quitGenerateMode:
		}
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

// AddLinkedNode add the block
func (pool *TxPImpl) AddLinkedNode(linkedNode *blockcache.BlockCacheNode, headNode *blockcache.BlockCacheNode) error {
	//ilog.Infof("block: %+v", linkedNode.Block)
	//ilog.Infof("headNode block:%+v", headNode.Block)
	if linkedNode == nil || headNode == nil {
		return errors.New("parameter is nil")
	}

	if pool.addBlock(linkedNode.Block) != nil {
		return errors.New("failed to add block")
	}

	tFort := pool.updateForkChain(headNode)
	switch tFort {
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
	var r TAddTx

	if r = pool.verifyTx(t); r != Success {
		return r
	}
	if r = pool.addTx(t); r == Success {
		pool.p2pService.Broadcast(t.Encode(), p2p.PublishTxRequest, p2p.NormalMessage)
		metricsReceivedTxCount.Add(1, map[string]string{"from": "rpc"})
	}
	return r
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

// PendingTxs get the pending transactions
func (pool *TxPImpl) PendingTxs(maxCnt int) (TxsList, *blockcache.BlockCacheNode, error) {
	start := time.Now()
	defer func(t time.Time) {
		cost := time.Since(t).Nanoseconds() / int64(time.Microsecond)
		metricsGetPendingTxTime.Set(float64(cost), nil)
	}(start)

	pool.mu.Lock()
	cost := time.Since(start).Nanoseconds() / int64(time.Microsecond)
	metricsGetPendingTxLockTime.Set(float64(cost), nil)
	defer pool.mu.Unlock()

	start = time.Now()
	var pendingList TxsList
	iter := pool.pendingTx.Iter()
	var i int
	tx, ok := iter.Next()
	for ok && i < maxCnt {
		if !pool.TxTimeOut(tx) {
			pendingList = append(pendingList, tx)
		}
		tx, ok = iter.Next()
	}
	cost = time.Since(start).Nanoseconds() / int64(time.Microsecond)
	metricsGetPendingTxAppendTime.Set(float64(cost), nil)

	metricsTxPoolSize.Set(float64(pool.pendingTx.Size()), nil)

	return pendingList, pool.forkChain.NewHead, nil
}

// ExistTxs determine if the transaction exists
func (pool *TxPImpl) ExistTxs(hash []byte, chainBlock *block.Block) (FRet, error) {
	start := time.Now()
	defer func(t time.Time) {
		cost := time.Since(start).Nanoseconds() / int64(time.Microsecond)
		metricsExistTxTime.Observe(float64(cost), nil)
		metricsExistTxCount.Add(1, nil)
	}(start)

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
			if pool.verifyTx(v) != Success {
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

	t := pool.slotToNSec(chainBlock.Head.Time)
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

		if b, ok := pool.block(h); ok {
			if (t - b.time()) > filterTime {
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

func (pool *TxPImpl) initBlockTx() {
	chain := pool.global.BlockChain()
	timeNow := time.Now().UnixNano()

	for i := chain.Length() - 1; i > 0; i-- {
		blk, err := chain.GetBlockByNumber(i)
		if err != nil {
			return
		}

		t := pool.slotToNSec(blk.Head.Time)
		if timeNow-t <= filterTime {
			pool.addBlock(blk)
		}
	}

}

func (pool *TxPImpl) verifyTx(t *tx.Tx) TAddTx {
	if pool.pendingTx.Size() > maxCacheTxs {
		return CacheFullError
	}
	start := time.Now()
	defer func(t time.Time) {
		cost := time.Since(start).Nanoseconds() / int64(time.Microsecond)
		metricsVerifyTxTime.Observe(float64(cost), nil)
		metricsVerifyTxCount.Add(1, nil)
	}(start)

	if t.GasPrice <= 0 {
		return GasPriceError
	}

	if pool.TxTimeOut(t) {
		return TimeError
	}

	if err := t.VerifySelf(); err != nil {
		return VerifyError
	}

	return Success
}

func (pool *TxPImpl) slotToNSec(t int64) int64 {
	slot := common.Timestamp{Slot: t}
	return slot.ToUnixSec() * int64(time.Second)
}

func (pool *TxPImpl) addBlock(linkedBlock *block.Block) error {

	if linkedBlock == nil {
		return errors.New("failed to linkedBlock")
	}

	h := linkedBlock.HeadHash()

	if _, ok := pool.blockList.Load(string(h)); ok {
		return nil
	}

	b := newBlockTx()

	b.setTime(pool.slotToNSec(linkedBlock.Head.Time))
	b.addBlock(linkedBlock)

	pool.blockList.Store(string(h), b)

	return nil
}

func (pool *TxPImpl) parentHash(hash []byte) ([]byte, bool) {

	v, ok := pool.block(hash)
	if !ok {
		return nil, false
	}

	return v.ParentHash, true
}

func (pool *TxPImpl) block(hash []byte) (*blockTx, bool) {

	if v, ok := pool.blockList.Load(string(hash)); ok {
		return v.(*blockTx), true
	}

	return nil, false
}

func (pool *TxPImpl) existTxInChain(txHash []byte, block *block.Block) bool {

	if block == nil {
		return false
	}

	h := block.HeadHash()

	t := pool.slotToNSec(block.Head.Time)
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

}

func (pool *TxPImpl) existTxInBlock(txHash []byte, blockHash []byte) bool {

	b, ok := pool.blockList.Load(string(blockHash))
	if !ok {
		return false
	}

	return b.(*blockTx).existTx(txHash)
}

func (pool *TxPImpl) clearBlock() {
	if pool.global.Mode() == global.ModeInit {
		return
	}
	ft := pool.slotToNSec(pool.blockCache.LinkedRoot().Block.Head.Time) - filterTime

	pool.blockList.Range(func(key, value interface{}) bool {
		if value.(*blockTx).time() < ft {
			pool.blockList.Delete(key.(string))
		}
		return true
	})

}

func (pool *TxPImpl) addTx(tx *tx.Tx) TAddTx {
	start := time.Now()
	defer func(t time.Time) {
		cost := time.Since(start).Nanoseconds() / int64(time.Microsecond)
		metricsAddTxTime.Observe(float64(cost), nil)
		metricsAddTxCount.Add(1, nil)
	}(start)

	h := tx.Hash()
	if pool.existTxInChain(h, pool.forkChain.NewHead.Block) {
		return DupError
	}
	if pool.existTxInPending(h) {
		return DupError
	}
	pool.pendingTx.Add(tx)
	return Success
}

func (pool *TxPImpl) existTxInPending(hash []byte) bool {

	tx := pool.pendingTx.Get(hash)

	return tx != nil
}

// TxTimeOut time to verify the tx
func (pool *TxPImpl) TxTimeOut(tx *tx.Tx) bool {
	nTime := time.Now().UnixNano()
	txTime := tx.Time
	exTime := tx.Expiration

	if txTime > nTime {
		metricsTxErrType.Add(1, map[string]string{"type": "txTime > nTime"})
		return true
	}

	if exTime <= nTime {
		metricsTxErrType.Add(1, map[string]string{"type": "exTime <= nTime"})
		return true
	}

	if nTime-txTime > Expiration {
		metricsTxErrType.Add(1, map[string]string{"type": "nTime-txTime > expiration"})
		return true
	}
	return false
}

func (pool *TxPImpl) clearTimeOutTx() {

	iter := pool.pendingTx.Iter()
	tx, ok := iter.Next()
	for ok {
		if pool.TxTimeOut(tx) {
			pool.DelTx(tx.Hash())
		}
		tx, ok = iter.Next()
	}

}

func (pool *TxPImpl) delBlockTxInPending(hash []byte) error {

	b, ok := pool.block(hash)
	if !ok {
		return nil
	}

	b.txMap.Range(func(key, value interface{}) bool {
		pool.pendingTx.Del([]byte(key.(string)))
		return true
	})

	return nil
}

func (pool *TxPImpl) clearTxPending() {
	pool.pendingTx = NewSortedTxMap()
}

func (pool *TxPImpl) updatePending(blockHash []byte) error {

	b, ok := pool.block(blockHash)
	if !ok {
		return errors.New("updatePending is error")
	}

	b.txMap.Range(func(key, value interface{}) bool {
		pool.DelTx(key.([]byte))
		return true
	})

	return nil
}

func (pool *TxPImpl) updateForkChain(headNode *blockcache.BlockCacheNode) tFork {
	if pool.forkChain.NewHead == headNode {
		return sameHead
	}
	pool.forkChain.OldHead, pool.forkChain.NewHead = pool.forkChain.NewHead, headNode
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

		_, ok := pool.block(newHead.Block.HeadHash())
		if !ok {
			if err := pool.addBlock(newHead.Block); err != nil {
				ilog.Errorf("failed to add block, err = %v", err)
			}
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
	//Reply to txs
	ft := time.Now().UnixNano() - filterTime
	for {
		if oldHead == nil || oldHead == forkBCN || pool.slotToNSec(oldHead.Block.Head.Time) < ft {
			break
		}
		for _, t := range oldHead.Block.Txs {
			pool.pendingTx.Add(t)
		}
		oldHead = oldHead.Parent
	}

	//Check duplicate txs
	for {
		if newHead == nil || newHead == forkBCN || pool.slotToNSec(newHead.Block.Head.Time) < ft {
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
	ft := time.Now().UnixNano() - filterTime
	ob, ok := pool.block(oldHead.Block.HeadHash())
	if ok {
		for {
			if ob.time() < ft {
				break
			}
			ob.txMap.Range(func(k, v interface{}) bool {
				t := v.(*tx.Tx)
				pool.pendingTx.Add(t)
				return true
			})
			ob, ok = pool.block(ob.ParentHash)
			if !ok {
				break
			}
		}
	}
	nb, ok := pool.block(newHead.Block.HeadHash())
	if ok {
		for {
			if nb.time() < ft {
				break
			}
			nb.txMap.Range(func(k, v interface{}) bool {
				t := v.(*tx.Tx)
				pool.DelTx(t.Hash())
				return true
			})
			nb, ok = pool.block(nb.ParentHash)
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
		//fmt.Println("blockList hash:", []byte(key.(string)))
		return true
	})
	return r
}
