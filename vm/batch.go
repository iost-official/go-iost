package vm

import (
	"time"

	"sync"

	"github.com/iost-official/go-iost/core/block"
	"github.com/iost-official/go-iost/core/tx"
	"github.com/iost-official/go-iost/vm/database"
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
	var (
		mappers = make([]map[string]database.Access, thread)
		txs = make([]*tx.Tx, thread)
		receipts = make([]*tx.TxReceipt, thread)
		visitors = make([]*database.Visitor, thread)
	)

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
				mappers[i2] = mapper.Map()
				txs[i2] = t
				receipts[i2] = tr
				visitors[i2] = vi
			} else {
				mappers[i2] = nil
			}
		}()
	}
	m.wait.Wait()
	ti, td := Resolve(mappers)

	for _, i := range td {
		m.sender.Return(txs[i])
	}

	b := NewBatch()

	for _, i := range ti {
		b.Txs = append(b.Txs, txs[i])
		b.Receipts = append(b.Receipts,receipts[i])
		visitors[i].Commit()
	}

	return b
}

func Resolve(mappers []map[string]database.Access) (accept, drop []int) {
	workMap := make(map[string]database.Access)
	accept = make([]int, 0)
	drop = make([]int, 0)

	for i, m := range mappers {
		if m == nil {
			continue
		}
		for k, v := range m {
			x, ok := workMap[k]
			switch {
			case !ok:
				workMap[k] = v
			case x == database.Read && v == database.Read:
			case x == database.Write || v == database.Write:
				drop = append(drop, i)
				continue
			}
		}
		accept = append(accept, i)
	}
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
	v.wait.Wait()
	return errs
}
