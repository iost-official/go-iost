package txpool

import (
	"errors"
	"fmt"
	"runtime"
	"sync"
	"time"

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
	blockList        *sync.Map // map[string]*blockTx
	pendingTx        *SortedTxMap
	mu               sync.RWMutex
	chP2PTx          chan p2p.IncomingMessage
	deferServer      *DeferServer
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
	deferServer, err := NewDeferServer(p)
	if err != nil {
		return nil, err
	}
	p.deferServer = deferServer
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

// AddDefertx adds defer transaction.
func (pool *TxPImpl) AddDefertx(txHash []byte) error {
	if pool.pendingTx.Size() > maxCacheTxs {
		return ErrCacheFull
	}
	referredTx, err := pool.global.BlockChain().GetTx(txHash)
	if err != nil {
		return err
	}
	t := &tx.Tx{
		Actions:    referredTx.Actions,
		Time:       referredTx.Time + referredTx.Delay,
		Expiration: referredTx.Expiration + referredTx.Delay,
		GasLimit:   referredTx.GasLimit,
		GasPrice:   referredTx.GasPrice,
		Publisher:  referredTx.Publisher,
		ReferredTx: txHash,
	}
	return pool.addTx(t)
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
			pool.mu.Lock()
			pool.clearBlock()
			pool.clearTimeoutTx()
			pool.mu.Unlock()
			metricsTxPoolSize.Set(float64(pool.pendingTx.Size()), nil)
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
			ilog.Errorf("decode tx error. err=%v", err)
			continue
		}
		err = pool.verifyTx(&t)
		if err != nil {
			ilog.Errorf("verify tx error. err=%v", err)
			continue
		}
		err = pool.addTx(&t)
		if err != nil {
			continue
		}
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
	dtm := make(map[string]struct{})
	for _, v := range txs {
		trh := string(v.Hash())
		if _, ok := rm[trh]; ok {
			return v, errors.New("duplicate tx in chain")
		}
		if _, ok := dtm[trh]; ok {
			return v, errors.New("duplicate tx in txs")
		}
		dtm[trh] = struct{}{}
		if ok := pool.existTxInPending([]byte(trh)); !ok {
			if pool.verifyTx(v) != nil {
				return v, errors.New("failed to verify")
			}
		}
	}
	return nil, nil
}

func (pool *TxPImpl) createTxMapToChain(chainBlock *block.Block) (map[string]struct{}, error) {
	if chainBlock == nil {
		return nil, errors.New("chainBlock is nil")
	}
	rm := make(map[string]struct{})
	h := chainBlock.HeadHash()
	t := chainBlock.Head.Time
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

func (pool *TxPImpl) createTxMapToBlock(tm map[string]struct{}, blockHash []byte) bool {
	b, ok := pool.blockList.Load(string(blockHash))
	if !ok {
		return false
	}
	b.(*blockTx).txMap.Range(func(key, value interface{}) bool {
		tm[key.(string)] = struct{}{}
		return true
	})
	return true
}

func (pool *TxPImpl) processDelaytx(blk *block.Block) {
	for _, t := range blk.Txs {
		if t.Delay > 0 {
			pool.deferServer.StoreDeferTx(t)
		}
		if t.IsDefer() {
			pool.deferServer.DelDeferTx(t)
		}
	}
}

// AddLinkedNode add the findBlock
func (pool *TxPImpl) AddLinkedNode(linkedNode *blockcache.BlockCacheNode, newHead *blockcache.BlockCacheNode) error {
	pool.processDelaytx(linkedNode.Block)
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
func (pool *TxPImpl) AddTx(t *tx.Tx) error {
	err := pool.verifyTx(t)
	if err != nil {
		return err
	}
	err = pool.addTx(t)
	if err != nil {
		return err
	}
	pool.p2pService.Broadcast(t.Encode(), p2p.PublishTx, p2p.NormalMessage, true)
	metricsReceivedTxCount.Add(1, map[string]string{"from": "rpc"})
	return nil
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
	return pool.pendingTx.Iter(), pool.forkChain.NewHead
}

// ExistTxs determine if the transaction exists
func (pool *TxPImpl) ExistTxs(hash []byte, chainBlock *block.Block) FRet {
	var r FRet
	switch {
	case pool.existTxInPending(hash):
		r = FoundPending
	case pool.existTxInChain(hash, chainBlock):
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
		if blk.Head.Time < filterLimit {
			break
		}
		pool.addBlock(blk)
	}
}

func (pool *TxPImpl) verifyTx(t *tx.Tx) error {
	if pool.pendingTx.Size() > maxCacheTxs {
		return ErrCacheFull
	}
	if t.GasPrice < 100 {
		return fmt.Errorf("GasPriceError. gas price %d", t.GasPrice)
	}
	if !t.IsTimeValid(time.Now().UnixNano()) {
		return fmt.Errorf("TimeError")
	}
	if err := t.VerifySelf(); err != nil {
		return fmt.Errorf("VerifyError %v", err)
	}

	if t.IsDefer() {
		referredTx, err := pool.global.BlockChain().GetTx(t.ReferredTx)
		if err != nil {
			return fmt.Errorf("get referred tx error, %v", err)
		}
		err = t.VerifyDefer(referredTx)
		if err != nil {
			return err
		}
	}

	return nil
}

func (pool *TxPImpl) addBlock(blk *block.Block) error {
	if blk == nil {
		return errors.New("failed to linkedBlock")
	}
	pool.blockList.LoadOrStore(string(blk.HeadHash()), newBlockTx(blk))
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

func (pool *TxPImpl) existTxInChain(txHash []byte, block *block.Block) bool {
	if block == nil {
		return false
	}
	h := block.HeadHash()
	filterLimit := block.Head.Time - filterTime
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

func (pool *TxPImpl) existTxInBlock(txHash []byte, blockHash []byte) bool {
	b, ok := pool.blockList.Load(string(blockHash))
	if !ok {
		return false
	}
	return b.(*blockTx).existTx(txHash)
}

func (pool *TxPImpl) clearBlock() {
	filterLimit := pool.blockCache.LinkedRoot().Block.Head.Time - filterTime
	pool.blockList.Range(func(key, value interface{}) bool {
		if value.(*blockTx).time < filterLimit {
			pool.blockList.Delete(key)
		}
		return true
	})
}

func (pool *TxPImpl) addTx(tx *tx.Tx) error {
	if pool.existTxInPending(tx.Hash()) {
		return ErrDupPendingTx
	}
	if pool.existTxInChain(tx.Hash(), pool.forkChain.NewHead.Block) {
		return ErrDupChainTx
	}
	pool.pendingTx.Add(tx)
	return nil
}

func (pool *TxPImpl) existTxInPending(hash []byte) bool {
	return pool.pendingTx.Get(hash) != nil
}

func (pool *TxPImpl) clearTimeoutTx() {
	iter := pool.pendingTx.Iter()
	t, ok := iter.Next()
	for ok {
		if !t.IsTimeValid(time.Now().UnixNano()) && !t.IsDefer() {
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
		if oldHead == nil || oldHead == forkBCN || oldHead.Block.Head.Time < filterLimit {
			break
		}
		for _, t := range oldHead.Block.Txs {
			pool.pendingTx.Add(t)
		}
		oldHead = oldHead.Parent
	}

	//del txs
	for {
		if newHead == nil || newHead == forkBCN || newHead.Block.Head.Time < filterLimit {
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
