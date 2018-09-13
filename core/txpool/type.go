package txpool

import (
	"sync"
	"time"

	"github.com/iost-official/Go-IOS-Protocol/core/block"
	"github.com/iost-official/Go-IOS-Protocol/core/blockcache"
	"github.com/iost-official/Go-IOS-Protocol/core/tx"
	"github.com/iost-official/Go-IOS-Protocol/metrics"
	"github.com/yasushi-saito/rbtree"
)

var (
	clearInterval = 10 * time.Second
	expiration    = int64(60 * time.Second)
	filterTime    = int64(expiration + expiration/2)
	//expiration    = 60*60*24*7

	metricsReceivedTxCount        = metrics.NewCounter("iost_tx_received_count", []string{"from"})
	metricsGetPendingTxTime       = metrics.NewGauge("iost_get_pending_tx_time", nil)
	metricsGetPendingTxLockTime   = metrics.NewGauge("iost_get_pending_tx_lock_time", nil)
	metricsGetPendingTxSortTime   = metrics.NewGauge("iost_get_pending_tx_sort_time", nil)
	metricsGetPendingTxAppendTime = metrics.NewGauge("iost_get_pending_tx_append_time", nil)
	metricsExistTxTime            = metrics.NewSummary("iost_exist_tx_time", nil)
	metricsExistTxCount           = metrics.NewCounter("iost_exist_tx_count", nil)
	metricsVerifyTxTime           = metrics.NewSummary("iost_verify_tx_time", nil)
	metricsVerifyTxCount          = metrics.NewCounter("iost_verify_tx_count", nil)
	metricsAddTxTime              = metrics.NewSummary("iost_add_tx_time", nil)
	metricsAddTxCount             = metrics.NewCounter("iost_add_tx_count", nil)
	metricsTxPoolSize             = metrics.NewGauge("iost_txpool_size", nil)
)

type FRet uint

const (
	NotFound FRet = iota
	FoundPending
	FoundChain
)

type TFork uint

const (
	NotFork TFork = iota
	Fork
	ForkError
)

type TAddTx uint

const (
	Success TAddTx = iota
	TimeError
	VerifyError
	DupError
	GasPriceError
)

type ForkChain struct {
	NewHead *blockcache.BlockCacheNode
	OldHead *blockcache.BlockCacheNode
	ForkBCN *blockcache.BlockCacheNode
}

type TxsList []*tx.Tx

func (s TxsList) Len() int { return len(s) }
func (s TxsList) Less(i, j int) bool {
	if s[i].GasPrice > s[j].GasPrice {
		return true
	}

	if s[i].GasPrice == s[j].GasPrice {
		if s[i].Time > s[j].Time {
			return false
		} else {
			return true
		}
	}
	return false
}
func (s TxsList) Swap(i, j int) { s[i], s[j] = s[j], s[i] }

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

type sortedTxMap struct {
	tree  *rbtree.Tree
	txMap map[string]*tx.Tx
	rw    *sync.RWMutex
}

func compareTx(a, b rbtree.Item) int {
	txa := a.(*tx.Tx)
	txb := b.(*tx.Tx)
	if txa.GasPrice == txb.GasPrice {
		return int(txb.Time - txa.Time)
	}
	return int(txa.GasPrice - txb.GasPrice)
}

func newSortedTxMap() *sortedTxMap {
	return &sortedTxMap{
		tree:  rbtree.NewTree(compareTx),
		txMap: make(map[string]*tx.Tx),
		rw:    new(sync.RWMutex),
	}
}

func (st *sortedTxMap) Get(hash []byte) *tx.Tx {
	st.rw.RLock()
	defer st.rw.RUnlock()
	return st.txMap[string(hash)]
}

func (st *sortedTxMap) Add(tx *tx.Tx) {
	st.rw.Lock()
	st.tree.Insert(tx)
	st.txMap[string(tx.Hash())] = tx
	st.rw.Unlock()
}

func (st *sortedTxMap) Del(hash []byte) {
	st.rw.Lock()
	defer st.rw.Unlock()

	tx := st.txMap[string(hash)]
	if tx == nil {
		return
	}
	st.tree.DeleteWithKey(tx)
	delete(st.txMap, string(hash))
}

func (st *sortedTxMap) Size() int {
	st.rw.Lock()
	defer st.rw.Unlock()

	return len(st.txMap)
}

func (st *sortedTxMap) Iter() *Iterator {
	iter := st.tree.Limit()
	ret := &Iterator{
		iter: &iter,
		rw:   st.rw,
		res:  make(chan *iterRes, 1),
	}
	go ret.getNext()
	return ret
}

type Iterator struct {
	iter *rbtree.Iterator
	rw   *sync.RWMutex
	res  chan *iterRes
}

type iterRes struct {
	tx *tx.Tx
	ok bool
}

func (iter *Iterator) getNext() {
	i := iter.iter.Prev()
	if i.NegativeLimit() {
		iter.res <- &iterRes{nil, false}
		return
	}
	iter.iter = &i
	iter.res <- &iterRes{i.Item().(*tx.Tx), true}
}

func (iter *Iterator) Next() (*tx.Tx, bool) {
	ret := <-iter.res
	go iter.getNext()
	return ret.tx, ret.ok
}
