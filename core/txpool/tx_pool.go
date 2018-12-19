package txpool

import (
	"errors"
	"fmt"
	"runtime"
	"sync"
	"time"

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
	p.forkChain.SetNewHead(blockCache.Head())
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
	go pool.deferServer.Start()
	go pool.loop()
	return nil
}

// Stop stops all the jobs.
func (pool *TxPImpl) Stop() {
	pool.deferServer.Stop()
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
		Actions:      referredTx.Actions,
		Time:         referredTx.Time + referredTx.Delay,
		Expiration:   referredTx.Expiration + referredTx.Delay,
		GasLimit:     referredTx.GasLimit,
		GasRatio:     referredTx.GasRatio,
		Publisher:    referredTx.Publisher,
		ReferredTx:   txHash,
		AmountLimit:  referredTx.AmountLimit,
		PublishSigns: referredTx.PublishSigns,
		Signs:        referredTx.Signs,
		Signers:      referredTx.Signers,
	}
	err = pool.verifyDuplicate(t)
	if err != nil {
		return err
	}
	pool.pendingTx.Add(t)
	return nil
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

// PendingTx is return pendingTx
func (pool *TxPImpl) PendingTx() (*SortedTxMap, *blockcache.BlockCacheNode) {
	return pool.pendingTx, pool.forkChain.NewHead
}

// Release release the txpool
func (pool *TxPImpl) Release() {
	close(pool.quitGenerateMode)
	pool.mu.Unlock()
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
		pool.mu.Lock()
		ret := pool.verifyDuplicate(&t)
		if ret != nil {
			pool.mu.Unlock()
			continue
		}
		ret = pool.verifyTx(&t)
		if ret != nil {
			pool.mu.Unlock()
			continue
		}
		pool.pendingTx.Add(&t)
		pool.mu.Unlock()
		metricsReceivedTxCount.Add(1, map[string]string{"from": "p2p"})
		pool.p2pService.Broadcast(v.Data(), p2p.PublishTx, p2p.NormalMessage)
	}
}

func (pool *TxPImpl) processDelaytx(blk *block.Block) {
	for i, t := range blk.Txs {
		if t.Delay > 0 {
			pool.deferServer.StoreDeferTx(t)
		}
		if t.IsDefer() {
			pool.deferServer.DelDeferTx(t)
		}
		if cancelHash, exist := t.CanceledDelaytxHash(); exist {
			if blk.Receipts[i].Status.Code == tx.Success {
				pool.deferServer.DelDeferTxByHash(cancelHash)
			}
		}
	}
}

// AddLinkedNode add the findBlock
func (pool *TxPImpl) AddLinkedNode(linkedNode *blockcache.BlockCacheNode) error {
	pool.mu.Lock()
	defer pool.mu.Unlock()
	pool.processDelaytx(linkedNode.Block)
	err := pool.addBlock(linkedNode.Block)
	if err != nil {
		return fmt.Errorf("failed to add findBlock: %v", err)
	}
	var newHead *blockcache.BlockCacheNode
	h := pool.blockCache.Head()
	if linkedNode.Head.Number > h.Head.Number {
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
func (pool *TxPImpl) AddTx(t *tx.Tx) error {
	err := pool.verifyDuplicate(t)
	if err != nil {
		return err
	}
	err = pool.verifyTx(t)
	if err != nil {
		return err
	}
	pool.pendingTx.Add(t)
	ilog.Debugf(
		"Added %v to pendingTx, now size is %v.",
		common.Base58Encode(t.Hash()),
		pool.pendingTx.Size(),
	)

	pool.p2pService.Broadcast(t.Encode(), p2p.PublishTx, p2p.NormalMessage)
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
	if err := t.CheckGas(); err != nil {
		return err
	}
	// Add one second delay for tx created time check
	if !t.IsCreatedBefore(time.Now().UnixNano()+(time.Second).Nanoseconds()) || t.IsExpired(time.Now().UnixNano()) {
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

func (pool *TxPImpl) getTxAndReceiptInChain(txHash []byte, block *block.Block) (*tx.Tx, *tx.TxReceipt) {
	if block == nil {
		return nil, nil
	}
	blkHash := block.HeadHash()
	filterLimit := block.Head.Time - filterTime
	var ok bool
	for {
		t, tr := pool.getTxAndReceiptInBlock(txHash, blkHash)
		if t != nil {
			return t, tr
		}
		blkHash, ok = pool.parentHash(blkHash)
		if !ok {
			return nil, nil
		}
		if b, ok := pool.findBlock(blkHash); ok {
			if b.time < filterLimit {
				return nil, nil
			}
		}
	}
}

func (pool *TxPImpl) existTxInChain(txHash []byte, block *block.Block) bool {
	t, _ := pool.getTxAndReceiptInChain(txHash, block)
	return t != nil
}

func (pool *TxPImpl) getTxAndReceiptInBlock(txHash []byte, blockHash []byte) (*tx.Tx, *tx.TxReceipt) {
	b, ok := pool.blockList.Load(string(blockHash))
	if !ok {
		return nil, nil
	}
	return b.(*blockTx).getTxAndReceipt(txHash)
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

func (pool *TxPImpl) verifyDuplicate(t *tx.Tx) error {
	if pool.existTxInPending(t.Hash()) {
		return ErrDupPendingTx
	}
	if pool.existTxInChain(t.Hash(), pool.forkChain.GetNewHead().Block) {
		return ErrDupChainTx
	}
	return nil
}

func (pool *TxPImpl) existTxInPending(hash []byte) bool {
	return pool.pendingTx.Get(hash) != nil
}

func (pool *TxPImpl) clearTimeoutTx() {
	iter := pool.pendingTx.Iter()
	t, ok := iter.Next()
	for ok {
		if t.IsExpired(time.Now().UnixNano()) && !t.IsDefer() {
			pool.pendingTx.Del(t.Hash())
		}
		t, ok = iter.Next()
	}
}

func (pool *TxPImpl) updateForkChain(newHead *blockcache.BlockCacheNode) tFork {
	if pool.forkChain.GetNewHead() == newHead {
		return sameHead
	}
	pool.forkChain.SetOldHead(pool.forkChain.GetNewHead())
	pool.forkChain.SetNewHead(newHead)
	bcn, ok := pool.findForkBCN(pool.forkChain.GetNewHead(), pool.forkChain.GetOldHead())
	if ok {
		pool.forkChain.SetForkBCN(bcn)
		return forkBCN
	}
	pool.forkChain.SetForkBCN(nil)
	return noForkBCN
}

func (pool *TxPImpl) findForkBCN(newHead *blockcache.BlockCacheNode, oldHead *blockcache.BlockCacheNode) (*blockcache.BlockCacheNode, bool) {
	for {
		for oldHead != nil && oldHead.Head.Number > newHead.Head.Number {
			oldHead = oldHead.GetParent()
		}
		if oldHead == nil {
			return nil, false
		}
		if oldHead == newHead {
			return oldHead, true
		}
		newHead = newHead.GetParent()
		if newHead == nil {
			return nil, false
		}
	}
}

func (pool *TxPImpl) doChainChangeByForkBCN() {
	newHead := pool.forkChain.GetNewHead()
	oldHead := pool.forkChain.GetOldHead()
	forkBCN := pool.forkChain.GetForkBCN()
	//add txs
	filterLimit := time.Now().UnixNano() - filterTime
	for {
		if oldHead == nil || oldHead == forkBCN || oldHead.Block.Head.Time < filterLimit {
			break
		}
		for _, t := range oldHead.Block.Txs {
			pool.pendingTx.Add(t)
		}
		oldHead = oldHead.GetParent()
	}

	//del txs
	for {
		if newHead == nil || newHead == forkBCN || newHead.Block.Head.Time < filterLimit {
			break
		}
		for _, t := range newHead.Block.Txs {
			pool.DelTx(t.Hash())
		}
		newHead = newHead.GetParent()
	}
}

func (pool *TxPImpl) doChainChangeByTimeout() {
	newHead := pool.forkChain.GetNewHead()
	oldHead := pool.forkChain.GetOldHead()
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

// GetFromPending gets transaction from pending list.
func (pool *TxPImpl) GetFromPending(hash []byte) (*tx.Tx, error) {
	tx := pool.pendingTx.Get(hash)
	if tx == nil {
		return nil, ErrTxNotFound
	}
	return tx, nil
}

// GetFromChain gets transaction from longest chain.
func (pool *TxPImpl) GetFromChain(hash []byte) (*tx.Tx, *tx.TxReceipt, error) {
	t, tr := pool.getTxAndReceiptInChain(hash, pool.forkChain.GetNewHead().Block)
	if t == nil {
		return nil, nil, ErrTxNotFound
	}
	return t, tr, nil
}
