package txpool

import (
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/iost-official/go-iost/v3/common"
	"github.com/iost-official/go-iost/v3/core/block"
	"github.com/iost-official/go-iost/v3/core/blockcache"
	"github.com/iost-official/go-iost/v3/core/tx"
	"github.com/iost-official/go-iost/v3/ilog"
)

// TxPImpl defines all the API of txpool package.
type TxPImpl struct {
	bChain     block.Chain
	blockCache blockcache.BlockCache
	forkChain  *forkChain
	blockList  *sync.Map // map[string]*blockTx
	pendingTx  *SortedTxMap
	mu         sync.RWMutex
	quitCh     chan struct{}
}

// NewTxPoolImpl returns a default TxPImpl instance.
func NewTxPoolImpl(bChain block.Chain, blockCache blockcache.BlockCache) (*TxPImpl, error) {
	p := &TxPImpl{
		bChain:     bChain,
		blockCache: blockCache,
		forkChain:  new(forkChain),
		blockList:  new(sync.Map),
		pendingTx:  NewSortedTxMap(),
		quitCh:     make(chan struct{}),
	}
	p.forkChain.SetNewHead(blockCache.Head())
	p.initBlockTx()
	go p.loop()
	return p, nil
}

// Close will close the tx pool.
func (pool *TxPImpl) Close() {
	close(pool.quitCh)
}

func (pool *TxPImpl) loop() {
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

// PendingTx is return pendingTx
func (pool *TxPImpl) PendingTx() (*SortedTxMap, *blockcache.BlockCacheNode) {
	return pool.pendingTx, pool.forkChain.NewHead
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
func (pool *TxPImpl) AddTx(t *tx.Tx, from string) error {
	pool.mu.Lock()
	err := pool.verifyDuplicate(t)
	if err != nil {
		pool.mu.Unlock()
		return err
	}
	err = pool.verifyTx(t)
	if err != nil {
		pool.mu.Unlock()
		return err
	}
	pool.pendingTx.Add(t)
	pool.mu.Unlock()

	ilog.Debugf(
		"Added %v to pendingTx, now size is %v.",
		common.Base58Encode(t.Hash()),
		pool.pendingTx.Size(),
	)

	metricsReceivedTxCount.Add(1, map[string]string{"from": from})
	return nil
}

// DelTx del the transaction
func (pool *TxPImpl) DelTx(hash []byte) error {
	pool.pendingTx.Del(hash)
	return nil
}

// ExistTxs determine if the transaction exists
func (pool *TxPImpl) ExistTxs(hash []byte, chainBlock *block.Block) bool {
	t, _ := pool.getTxAndReceiptInChain(hash, chainBlock)
	return t != nil
}

func (pool *TxPImpl) initBlockTx() {
	filterLimit := time.Now().UnixNano() - filterTime
	for i := pool.bChain.Length() - 1; i > 0; i-- {
		blk, err := pool.bChain.GetBlockByNumber(i)
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
	if t.IsDefer() {
		return errors.New("reject defertx")
	}
	// Add one second delay for tx created time check
	currentTime := time.Now().UnixNano()
	if !t.IsCreatedBefore(currentTime + maxTxTimeGap) {
		return fmt.Errorf("TimeError: tx.time is too large(tx.time: %v, now: %v). Please sync time",
			t.Time, currentTime)
	}
	if t.IsExpired(time.Now().UnixNano()) {
		return fmt.Errorf("TimeError: tx.time is expired(tx.time: %v, now: %v). Please sync time",
			t.Time, currentTime)
	}
	if err := t.VerifySelf(); err != nil {
		return fmt.Errorf("VerifyError %v", err)
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
		ilog.Errorf("When get tx %v in chain, the block is nil!", common.Base58Encode(txHash))
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

func (pool *TxPImpl) getTxAndReceiptInBlock(txHash []byte, blockHash []byte) (*tx.Tx, *tx.TxReceipt) {
	b, ok := pool.blockList.Load(string(blockHash))
	if !ok {
		return nil, nil
	}
	return b.(*blockTx).getTxAndReceipt(txHash)
}

func (pool *TxPImpl) clearBlock() {
	filterLimit := pool.blockCache.LinkedRoot().Block.Head.Time - filterTime
	pool.blockList.Range(func(key, value any) bool {
		if value.(*blockTx).time < filterLimit {
			pool.blockList.Delete(key)
		}
		return true
	})
}

func (pool *TxPImpl) verifyDuplicate(t *tx.Tx) error {
	if pool.pendingTx.Get(t.Hash()) != nil {
		return ErrDupPendingTx
	}
	if t, _ := pool.getTxAndReceiptInChain(t.Hash(), pool.forkChain.GetNewHead().Block); t != nil {
		return ErrDupChainTx
	}
	return nil
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
			ob.txMap.Range(func(k, v any) bool {
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
			nb.txMap.Range(func(k, v any) bool {
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
