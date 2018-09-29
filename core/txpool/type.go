package txpool

import (
	"sync"
	"time"

	"github.com/emirpasic/gods/trees/redblacktree"
	"github.com/iost-official/go-iost/common"
	"github.com/iost-official/go-iost/core/block"
	"github.com/iost-official/go-iost/core/blockcache"
	"github.com/iost-official/go-iost/core/tx"
	"github.com/iost-official/go-iost/ilog"
	"github.com/iost-official/go-iost/metrics"
)

var (
	clearInterval          = 10 * time.Second
	Expiration             = int64(90 * time.Second) // Expiration is the transaction expiration
	filterTime             = int64(90 * time.Second)
	maxCacheTxs            = 30000
	metricsReceivedTxCount = metrics.NewCounter("iost_tx_received_count", []string{"from"})
	metricsTxPoolSize      = metrics.NewGauge("iost_txpool_size", nil)
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
	chainMap   *sync.Map
	txMap      *sync.Map
	ParentHash []byte
	time       int64
}

func (pool *TxPImpl) newBlockTx(blk *block.Block, parentBlockTx interface{}) *blockTx {
	ilog.Errorf("new block number: %v", blk.Head.Number)
	b := &blockTx{
		chainMap:   new(sync.Map),
		txMap:      new(sync.Map),
		ParentHash: blk.Head.ParentHash,
		time:       common.SlotLength * blk.Head.Time * int64(time.Second),
	}
	prevcnt := 0
	nowcnt := 0
	delcnt := 0
	addcnt := 0
	if parentBlockTx != nil {
		parentBlockTx.(*blockTx).chainMap.Range(func(key, value interface{}) bool {
			prevcnt += 1
			if !pool.TxTimeOut(value.(*tx.Tx)) {
				nowcnt += 1
				b.chainMap.Store(key, value)
			} else {
				delcnt += 1
			}
			return true
		})
	}
	for _, v := range blk.Txs {
		nowcnt += 1
		addcnt += 1
		b.chainMap.Store(string(v.Hash()), v)
		b.txMap.Store(string(v.Hash()), v)
	}
	ilog.Errorf("new block number: %v, prevcnt: %v, nowcnt: %v, addcnt: %v, delcnt: %v", blk.Head.Number, prevcnt, nowcnt, addcnt, delcnt)
	return b
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
