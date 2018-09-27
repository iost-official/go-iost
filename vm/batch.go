package vm

import (
	"time"

	"sync"

	"github.com/iost-official/Go-IOS-Protocol/core/block"
	"github.com/iost-official/Go-IOS-Protocol/core/tx"
	"github.com/iost-official/Go-IOS-Protocol/vm/database"
)

type Batch struct {
	Txs      []*tx.Tx
	Receipts []*tx.TxReceipt
}

func NewBatch() *Batch {
	return &Batch{
		Txs:      make([]*tx.Tx, 0),
		Receipts: make([]*tx.TxReceipt, 0),
	}
}

type TxSender interface {
	Tx() *tx.Tx
	Return(*tx.Tx)
}

type Maker struct {
	sender TxSender
	wait   sync.WaitGroup
}

func NewMaker(sender TxSender) *Maker {
	return &Maker{
		sender: sender,
	}
}

func (m *Maker) Batch(bh *block.BlockHead, db database.IMultiValue, limit time.Duration, thread int) *Batch {
	var mappers, visitors, txs, receipts sync.Map

	bvr := database.NewBatchVisitorRoot(10000, db)
	for i := 0; i < thread; i++ {
		i2 := i
		go func() {
			vi, mapper := database.NewBatchVisitor(bvr)

			e := newEngine(bh, vi)

			m.wait.Add(1)
			defer m.wait.Done()

			// todo setup engine=
			t := m.sender.Tx()
			tr, err := e.Exec(t, limit)

			if err == nil {
				mappers.Store(i2, mapper.Map())
				txs.Store(i2, t)
				receipts.Store(i2, tr)
				visitors.Store(i2, vi)
			} else {
				mappers.Store(i2, nil)
			}
		}()
	}
	m.wait.Wait()
	ti, td := Resolve(mappers)

	for _, i := range td {
		t, _ := txs.Load(i)
		m.sender.Return(t.(*tx.Tx))
	}

	b := NewBatch()

	for _, i := range ti {
		t, _ := txs.Load(i)
		r, _ := receipts.Load(i)
		b.Txs = append(b.Txs, t.(*tx.Tx))
		b.Receipts = append(b.Receipts, r.(*tx.TxReceipt))
		vi, _ := visitors.Load(i)
		vi.(*database.Visitor).Commit()
	}

	return b
}

func Resolve(mappers sync.Map) (accept, drop []int) {
	workMap := make(map[string]database.Access)
	accept = make([]int, 0)
	drop = make([]int, 0)
	mappers.Range(func(key, value interface{}) bool {
		if value == nil {
			return true
		}
		for k, v := range value.(map[string]database.Access) {
			x, ok := workMap[k]
			switch {
			case !ok:
				workMap[k] = v
			case x == database.Read && v == database.Read:
			case x == database.Write || v == database.Write:
				drop = append(drop, key.(int))
				return true
			}
		}
		accept = append(accept, key.(int))
		return true
	})
	return
}

type Verifier struct {
	wait sync.WaitGroup
}

func (v *Verifier) Do(b *Batch, checkFunc func(t *tx.Tx, r *tx.TxReceipt) error) []error {
	var errs = make([]error, len(b.Txs))
	for i := range b.Txs {
		i2 := i
		go func() {
			v.wait.Add(1)
			defer v.wait.Done()
			errs[i2] = checkFunc(b.Txs[i2], b.Receipts[i2])
		}()
	}
	return errs
}
