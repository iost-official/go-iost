package txpool

import (
	"bytes"
	"errors"
	"sync"
	"time"

	"github.com/emirpasic/gods/trees/redblacktree"
	"github.com/iost-official/go-iost/v3/core/block"
	"github.com/iost-official/go-iost/v3/core/blockcache"
	"github.com/iost-official/go-iost/v3/core/tx"
	"github.com/iost-official/go-iost/v3/metrics"
)

// Values.
var (
	clearInterval = 10 * time.Second
	filterTime    = int64(90 * time.Second)
	maxCacheTxs   = 10000
	maxTxTimeGap  = 30 * time.Second.Nanoseconds()

	metricsReceivedTxCount = metrics.NewCounter("iost_tx_received_count", []string{"from"})
	metricsTxPoolSize      = metrics.NewGauge("iost_txpool_size", nil)

	ErrDupPendingTx = errors.New("tx exists in pending")
	ErrDupChainTx   = errors.New("tx exists in chain")
	ErrCacheFull    = errors.New("txpool is full")
	ErrTxNotFound   = errors.New("tx not found")
)

// FRet find the return value of the tx
type FRet uint

const (
	// NotFound ...
	NotFound FRet = iota
	// FoundPending ...
	FoundPending
	// FoundChain ...
	FoundChain
)

// tFork ...
type tFork uint

const (
	sameHead tFork = iota
	forkBCN
	noForkBCN
)

type forkChain struct {
	nrw     sync.RWMutex
	NewHead *blockcache.BlockCacheNode
	orw     sync.RWMutex
	OldHead *blockcache.BlockCacheNode
	frw     sync.RWMutex
	ForkBCN *blockcache.BlockCacheNode
}

// GetNewHead returns forkChain's new head.
func (f *forkChain) GetNewHead() *blockcache.BlockCacheNode {
	f.nrw.RLock()
	defer f.nrw.RUnlock()
	return f.NewHead
}

// SetNewHead sets forkChain's new head.
func (f *forkChain) SetNewHead(n *blockcache.BlockCacheNode) {
	f.nrw.Lock()
	f.NewHead = n
	f.nrw.Unlock()
}

// GetOldHead returns forkChain's old head.
func (f *forkChain) GetOldHead() *blockcache.BlockCacheNode {
	f.orw.RLock()
	defer f.orw.RUnlock()
	return f.OldHead
}

// SetOldHead sets forkChain's old head.
func (f *forkChain) SetOldHead(n *blockcache.BlockCacheNode) {
	f.orw.Lock()
	f.OldHead = n
	f.orw.Unlock()
}

// GetForkBCN returns forkChain's fork bcn.
func (f *forkChain) GetForkBCN() *blockcache.BlockCacheNode {
	f.frw.RLock()
	defer f.frw.RUnlock()
	return f.ForkBCN
}

func (f *forkChain) SetForkBCN(n *blockcache.BlockCacheNode) {
	f.frw.Lock()
	f.ForkBCN = n
	f.frw.Unlock()
}

type blockTx struct {
	txMap        *sync.Map // map[string]*tx.Tx
	txReceiptMap *sync.Map // map[string]*tx.TxReceipt
	ParentHash   []byte
	time         int64
}

func newBlockTx(blk *block.Block) *blockTx {
	b := &blockTx{
		txMap:        new(sync.Map),
		txReceiptMap: new(sync.Map),
		ParentHash:   blk.Head.ParentHash,
		time:         blk.Head.Time,
	}
	for _, v := range blk.Txs {
		b.txMap.Store(string(v.Hash()), v)
	}
	for _, v := range blk.Receipts {
		b.txReceiptMap.Store(string(v.TxHash), v)
	}
	return b
}

func (b *blockTx) getTxAndReceipt(hash []byte) (*tx.Tx, *tx.TxReceipt) {
	t, exist := b.txMap.Load(string(hash))
	if !exist {
		return nil, nil
	}
	retTx := t.(*tx.Tx)
	tr, exist := b.txReceiptMap.Load(string(hash))
	if exist {
		return retTx, tr.(*tx.TxReceipt)
	}
	return retTx, nil
}

// SortedTxMap is a red black tree of tx.
type SortedTxMap struct {
	tree  *redblacktree.Tree
	txMap map[string]*tx.Tx
	rw    *sync.RWMutex
}

func compareTx(a, b any) int {
	txa := a.(*tx.Tx)
	txb := b.(*tx.Tx)
	if txa.GasRatio == txb.GasRatio && txb.Time == txa.Time {
		return bytes.Compare(txa.Hash(), txb.Hash())
	}
	if txa.GasRatio == txb.GasRatio {
		return int(txb.Time - txa.Time)
	}
	return int(txa.GasRatio - txb.GasRatio)
}

// NewSortedTxMap returns a new SortedTxMap instance.
func NewSortedTxMap() *SortedTxMap {
	return &SortedTxMap{
		tree:  redblacktree.NewWith(compareTx),
		txMap: make(map[string]*tx.Tx),
		rw:    new(sync.RWMutex),
	}
}

// Get returns a tx of hash.
func (st *SortedTxMap) Get(hash []byte) *tx.Tx {
	st.rw.RLock()
	defer st.rw.RUnlock()
	return st.txMap[string(hash)]
}

// Add adds a tx in SortedTxMap.
func (st *SortedTxMap) Add(tx *tx.Tx) {
	st.rw.Lock()
	st.tree.Put(tx, true)
	st.txMap[string(tx.Hash())] = tx
	st.rw.Unlock()
}

// Del deletes a tx in SortedTxMap.
func (st *SortedTxMap) Del(hash []byte) {
	st.rw.Lock()
	defer st.rw.Unlock()

	tx := st.txMap[string(hash)]
	if tx == nil {
		return
	}
	st.tree.Remove(tx)
	delete(st.txMap, string(hash))
}

// Size returns the size of SortedTxMap.
func (st *SortedTxMap) Size() int {
	st.rw.RLock()
	defer st.rw.RUnlock()

	return len(st.txMap)
}

// Iter returns the iterator of SortedTxMap.
func (st *SortedTxMap) Iter() *Iterator {
	iter := st.tree.Iterator()
	iter.End()
	ret := &Iterator{
		iter: &iter,
		rw:   st.rw,
		res:  make(chan *iterRes, 1),
	}
	go ret.getNext()
	return ret
}

// Iterator This is the iterator
type Iterator struct {
	iter *redblacktree.Iterator
	rw   *sync.RWMutex
	res  chan *iterRes
}

type iterRes struct {
	tx *tx.Tx
	ok bool
}

func (iter *Iterator) getNext() {
	iter.rw.RLock()
	defer iter.rw.RUnlock()

	ok := iter.iter.Prev()
	if !ok {
		iter.res <- &iterRes{nil, false}
		return
	}
	iter.res <- &iterRes{iter.iter.Key().(*tx.Tx), true}
}

// Next next the tx
func (iter *Iterator) Next() (*tx.Tx, bool) {
	ret := <-iter.res
	go iter.getNext()
	return ret.tx, ret.ok
}
