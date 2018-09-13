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

// TxPoolImpl defines all the API of txpool package.
type TxPoolImpl struct {
	chP2PTx chan p2p.IncomingMessage
	chTx    chan *tx.Tx

	global     global.BaseVariable
	blockCache blockcache.BlockCache
	p2pService p2p.Service

	forkChain *ForkChain
	blockList *sync.Map
	// pendingTx *sync.Map
	pendingTx *sortedTxMap

	mu               sync.RWMutex
	quitGenerateMode chan struct{}
	quitCh           chan struct{}
}

// NewTxPoolImpl returns a default TxPoolImpl instance.
func NewTxPoolImpl(global global.BaseVariable, blockCache blockcache.BlockCache, p2ps p2p.Service) (*TxPoolImpl, error) {
	p := &TxPoolImpl{
		blockCache:       blockCache,
		chTx:             make(chan *tx.Tx, 102400),
		forkChain:        new(ForkChain),
		blockList:        new(sync.Map),
		pendingTx:        newSortedTxMap(),
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
func (pool *TxPoolImpl) Start() error {
	go pool.loop()
	return nil
}

// Stop stops all the jobs.
func (pool *TxPoolImpl) Stop() {
	ilog.Infof("TxPoolImpl Stop")
	close(pool.quitCh)
}

func (pool *TxPoolImpl) loop() {
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

func (pool *TxPoolImpl) Lock() {
	pool.mu.Lock()
	pool.quitGenerateMode = make(chan struct{})
}

func (pool *TxPoolImpl) Release() {
	pool.mu.Unlock()
	close(pool.quitGenerateMode)
}

func (pool *TxPoolImpl) verifyWorkers(p2pCh chan p2p.IncomingMessage, tCn chan *tx.Tx) {
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
func (pool *TxPoolImpl) AddLinkedNode(linkedNode *blockcache.BlockCacheNode, headNode *blockcache.BlockCacheNode) error {
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
	case ForkBCN:
		pool.mu.Lock()
		defer pool.mu.Unlock()
		pool.doChainChangeByForkBCN()
	case NoForkBCN:
		pool.mu.Lock()
		defer pool.mu.Unlock()
		pool.doChainChangeByTimeout()
	case SameHead:
	default:
		return errors.New("failed to tFort")
	}

	return nil
}

// AddTx add the transaction
func (pool *TxPoolImpl) AddTx(t *tx.Tx) TAddTx {

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
func (pool *TxPoolImpl) DelTx(hash []byte) error {

	pool.pendingTx.Del(hash)

	return nil
}

func (pool *TxPoolImpl) TxIterator() (*Iterator, *blockcache.BlockCacheNode) {
	metricsTxPoolSize.Set(float64(pool.pendingTx.Size()), nil)
	return pool.pendingTx.Iter(), pool.forkChain.NewHead
}

// PendingTxs get the pending transactions
func (pool *TxPoolImpl) PendingTxs(maxCnt int) (TxsList, *blockcache.BlockCacheNode, error) {
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
func (pool *TxPoolImpl) ExistTxs(hash []byte, chainBlock *block.Block) (FRet, error) {
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

// ExistTxs check txs
func (pool *TxPoolImpl) CheckTxs(txs []*tx.Tx, chainBlock *block.Block) (*tx.Tx, error) {

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

func (pool *TxPoolImpl) createTxMapToChain(chainBlock *block.Block) (*sync.Map, error) {

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

func (pool *TxPoolImpl) createTxMapToBlock(tm *sync.Map, blockHash []byte) bool {

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

func (pool *TxPoolImpl) initBlockTx() {
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

func (pool *TxPoolImpl) verifyTx(t *tx.Tx) TAddTx {

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

func (pool *TxPoolImpl) slotToNSec(t int64) int64 {
	slot := common.Timestamp{Slot: t}
	return slot.ToUnixSec() * int64(time.Second)
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

	b.setTime(pool.slotToNSec(linkedBlock.Head.Time))
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

func (pool *TxPoolImpl) existTxInBlock(txHash []byte, blockHash []byte) bool {

	b, ok := pool.blockList.Load(string(blockHash))
	if !ok {
		return false
	}

	return b.(*blockTx).existTx(txHash)
}

func (pool *TxPoolImpl) clearBlock() {
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

func (pool *TxPoolImpl) addTx(tx *tx.Tx) TAddTx {
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

func (pool *TxPoolImpl) existTxInPending(hash []byte) bool {

	tx := pool.pendingTx.Get(hash)

	return tx != nil
}

func (pool *TxPoolImpl) TxTimeOut(tx *tx.Tx) bool {
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

	if nTime-txTime > expiration {
		ilog.Error("nTime:", nTime, "txTime:", txTime, "nTime-txTime:", nTime-txTime, "expiration:", expiration)
		metricsTxErrType.Add(1, map[string]string{"type": "nTime-txTime > expiration"})
		return true
	}
	return false
}

func (pool *TxPoolImpl) clearTimeOutTx() {

	iter := pool.pendingTx.Iter()
	tx, ok := iter.Next()
	for ok {
		if pool.TxTimeOut(tx) {
			pool.DelTx(tx.Hash())
		}
		tx, ok = iter.Next()
	}

}

func (pool *TxPoolImpl) delBlockTxInPending(hash []byte) error {

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

func (pool *TxPoolImpl) clearTxPending() {
	pool.pendingTx = newSortedTxMap()
}

func (pool *TxPoolImpl) updatePending(blockHash []byte) error {

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

func (pool *TxPoolImpl) updateForkChain(headNode *blockcache.BlockCacheNode) TFork {
	if pool.forkChain.NewHead == headNode {
		return SameHead
	}
	pool.forkChain.OldHead, pool.forkChain.NewHead = pool.forkChain.NewHead, headNode
	bcn, ok := pool.findForkBCN(pool.forkChain.NewHead, pool.forkChain.OldHead)
	if ok {
		pool.forkChain.ForkBCN = bcn
		return ForkBCN
	}
	pool.forkChain.ForkBCN = nil
	return NoForkBCN

}

func (pool *TxPoolImpl) findForkBCN(newHead *blockcache.BlockCacheNode, oldHead *blockcache.BlockCacheNode) (*blockcache.BlockCacheNode, bool) {
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
				ilog.Errorf("failed to add block, err = ", err)
			}
		}
		newHead = newHead.Parent
		if newHead == nil {
			return nil, false
		}
	}
}

func (pool *TxPoolImpl) doChainChangeByForkBCN() {
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

func (pool *TxPoolImpl) doChainChangeByTimeout() {
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

func (pool *TxPoolImpl) testPendingTxsNum() int64 {
	return int64(pool.pendingTx.Size())
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
