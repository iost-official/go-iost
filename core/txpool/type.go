package txpool

import (
	"sync"
	"time"

	"github.com/emirpasic/gods/trees/redblacktree"
	"github.com/iost-official/Go-IOS-Protocol/core/block"
	"github.com/iost-official/Go-IOS-Protocol/core/blockcache"
	"github.com/iost-official/Go-IOS-Protocol/core/tx"
	"github.com/iost-official/Go-IOS-Protocol/metrics"
)

var (
	clearInterval = 10 * time.Second
	// Expiration is the transaction expiration
	Expiration  = int64(90 * time.Second)
	filterTime  = int64(90 * time.Second)
	maxCacheTxs = 30000

	metricsReceivedTxCount      = metrics.NewCounter("iost_tx_received_count", []string{"from"})
	metricsGetPendingTxTime     = metrics.NewGauge("iost_get_pending_tx_time", nil)
	metricsGetPendingTxLockTime = metrics.NewGauge("iost_get_pending_tx_lock_time", nil)
	//metricsGetPendingTxSortTime   = metrics.NewGauge("iost_get_pending_tx_sort_time", nil)
	metricsGetPendingTxAppendTime = metrics.NewGauge("iost_get_pending_tx_append_time", nil)
	metricsExistTxTime            = metrics.NewSummary("iost_exist_tx_time", nil)
	metricsExistTxCount           = metrics.NewCounter("iost_exist_tx_count", nil)
	metricsVerifyTxTime           = metrics.NewSummary("iost_verify_tx_time", nil)
	metricsVerifyTxCount          = metrics.NewCounter("iost_verify_tx_count", nil)
	metricsAddTxTime              = metrics.NewSummary("iost_add_tx_time", nil)
	metricsAddTxCount             = metrics.NewCounter("iost_add_tx_count", nil)
	metricsTxPoolSize             = metrics.NewGauge("iost_txpool_size", nil)
	metricsTxErrType              = metrics.NewCounter("iost_txerr_type", []string{"type"})
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

// TAddTx add the return value of the tx
type TAddTx uint

const (
	// Success ...
	Success TAddTx = iota
	// TimeError ...
	TimeError
	// VerifyError ...
	VerifyError
	// DupError ...
	DupError
	// GasPriceError ...
	GasPriceError
	// CacheFullError ...
	CacheFullError
)

type forkChain struct {
	NewHead *blockcache.BlockCacheNode
	OldHead *blockcache.BlockCacheNode
	ForkBCN *blockcache.BlockCacheNode
}

// TxsList tx sort
type TxsList []*tx.Tx

// Len ...
func (s TxsList) Len() int { return len(s) }

// Less ...
func (s TxsList) Less(i, j int) bool {
	if s[i].GasPrice > s[j].GasPrice {
		return true
	}

	if s[i].GasPrice == s[j].GasPrice {
		if s[i].Time < s[j].Time {
			return true
		}
	}
	return false
}

// Swap ...
func (s TxsList) Swap(i, j int) { s[i], s[j] = s[j], s[i] }

// Push ...
func (s *TxsList) Push(x *tx.Tx) {
	*s = append(*s, x)
}

type blockTx struct {
	txMap      *sync.Map
	ParentHash []byte
	cTime      int64
}

func newBlockTx() *blockTx {
	b := &blockTx{
		txMap:      new(sync.Map),
		ParentHash: make([]byte, 32),
	}

	return b
}

func (b *blockTx) time() int64 {
	return b.cTime
}

func (b *blockTx) setTime(t int64) {
	b.cTime = t
}

func (b *blockTx) addBlock(ib *block.Block) {

	for _, v := range ib.Txs {
		b.txMap.Store(string(v.Hash()), v)
	}
	b.ParentHash = ib.Head.ParentHash
}

func (b *blockTx) existTx(hash []byte) bool {

	_, r := b.txMap.Load(string(hash))

	return r
}

// SortedTxMap is a red black tree of tx.
type SortedTxMap struct {
	tree  *redblacktree.Tree
	txMap map[string]*tx.Tx
	rw    *sync.RWMutex
}

func compareTx(a, b interface{}) int {
	txa := a.(*tx.Tx)
	txb := b.(*tx.Tx)
	if txa.GasPrice == txb.GasPrice {
		return int(txb.Time - txa.Time)
	}
	return int(txa.GasPrice - txb.GasPrice)
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
	st.rw.Lock()
	defer st.rw.Unlock()

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
	ok := iter.iter.Prev()
	iter.rw.RUnlock()
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
