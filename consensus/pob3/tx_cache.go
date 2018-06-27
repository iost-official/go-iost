package pob3

import (
	"container/heap"

	"github.com/iost-official/prototype/common"
	"github.com/iost-official/prototype/core/tx"
)

type pQueue []*tx.Tx

func (pq pQueue) Len() int {
	return len(pq)
}
func (pq pQueue) Less(i, j int) bool {
	return pq[i].Contract.Info().Price > pq[j].Contract.Info().Price
}
func (pq pQueue) Swap(i, j int) {
	pq[i], pq[j] = pq[j], pq[i]
}
func (pq *pQueue) Push(x interface{}) {
	*pq = append(*pq, x.(*tx.Tx))
}
func (pq *pQueue) Pop() interface{} {
	old := *pq
	n := len(old)
	item := old[n-1]
	*pq = old[0 : n-1]
	return item
}

type TxHeap struct {
	pQueue
}

func NewTxHeap() TxHeap {
	return TxHeap{
		pQueue: make([]*tx.Tx, 0),
	}
}

func (t *TxHeap) Push(tx tx.Tx) {
	heap.Push(&t.pQueue, &tx)
}

func (t *TxHeap) Top() tx.Tx {
	return *t.pQueue[t.pQueue.Len()-1]
}

func (t *TxHeap) Pop() tx.Tx {
	rtn := heap.Pop(&t.pQueue)
	return *rtn.(*tx.Tx)
}

type CBT struct {
	buf []interface{}
}

func (c CBT) Put(x interface{}) int {
	c.buf = append(c.buf, x)
	return len(c.buf) - 1
}

func (c CBT) Get(i int) *interface{} {
	return &c.buf[i]
}

func (c CBT) Parent(i int) int {
	if i == 0 {
		return -1
	}
	return i / 2
}
func (c CBT) Sibling(i int) int {
	if i == 0 || i+1 >= len(c.buf) {
		return -1
	}
	if i%2 == 1 {
		return i + 1
	} else {
		return i - 1
	}
}

type TxTree struct {
	txs  []tx.Tx
	hash CBT
}

func (t TxTree) Push(tx tx.Tx) {
	t.txs = append(t.txs, tx)
	hash := tx.Hash()
	i := t.hash.Put(hash)
	for i >= 0 {
		p := t.hash.Parent(i)
		if p < 0 {
			return
		}
		s := t.hash.Sibling(i)
		if s < 0 {
			tmp := t.hash.Get(p)
			s = t.hash.Put(*tmp)
		}
		pp := t.hash.Get(p)
		ps := t.hash.Get(s)

		*pp = common.Sha256(append(hash, (*ps).([]byte)...))
		i = p
	}
}

func (t TxTree) Hash() []byte {
	return (*t.hash.Get(0)).([]byte)
}
