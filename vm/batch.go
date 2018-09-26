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
}

type Maker struct {
	sender TxSender
	wait   sync.WaitGroup
}

func (m *Maker) Batch(bh *block.BlockHead, db database.IMultiValue, limit time.Duration, thread int) *Batch {
	var mappers, txs, receipts sync.Map

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
				mappers.Store(i2, mapper)
				txs.Store(i2, t)
				receipts.Store(i2, tr)
			} else {
				mappers.Store(i2, nil)
				txs.Store(i2, nil)
				receipts.Store(i2, nil)
			}
		}()
	}
	m.wait.Wait()
	ti := m.Resolve(mappers)

	b := NewBatch()

	for _, i := range ti {
		t, _ := txs.Load(i)
		r, _ := receipts.Load(i)
		b.Txs = append(b.Txs, t.(*tx.Tx))
		b.Receipts = append(b.Receipts, r.(*tx.TxReceipt))
	}

	return b
}

func (m *Maker) Resolve(mappers sync.Map) []int {
	return nil
}
