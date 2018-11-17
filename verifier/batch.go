package verifier

import (
	"time"

	"sync"

	"fmt"

	"github.com/iost-official/go-iost/core/block"
	"github.com/iost-official/go-iost/core/tx"
	"github.com/iost-official/go-iost/ilog"
	"github.com/iost-official/go-iost/vm"
	"github.com/iost-official/go-iost/vm/database"
)

//go:generate mockgen -destination mock/batcher_mock.go -package mock github.com/iost-official/go-iost/vm Batcher

// Batcher batch generator and verifier
type Batcher interface {
	Batch(bh *block.BlockHead, db database.IMultiValue, provider Provider, limit time.Duration, thread int) *Batch
	Verify(bh *block.BlockHead, db database.IMultiValue, checkFunc func(e vm.Isolator, t *tx.Tx, r *tx.TxReceipt) error, b *Batch) error
}

// Batch tx batch in parallel
type Batch struct {
	Txs      []*tx.Tx
	Receipts []*tx.TxReceipt
}

// NewBatch make a new batch pointer
func NewBatch() *Batch {
	return &Batch{
		Txs:      make([]*tx.Tx, 0),
		Receipts: make([]*tx.TxReceipt, 0),
	}
}

//go:generate mockgen -destination mock/provider_mock.go -package mock github.com/iost-official/go-iost/vm Provider

// Provider of tx
type Provider interface {
	Tx() *tx.Tx
	Return(*tx.Tx)
	Drop(t *tx.Tx, err error)
	Close()
}

type batcherImpl struct {
	wait sync.WaitGroup
}

// NewBatcher init of Batcher
func NewBatcher() Batcher {
	return &batcherImpl{}
}

// Batch gen batch with verifier
func (m *batcherImpl) Batch(bh *block.BlockHead, db database.IMultiValue, provider Provider, limit time.Duration, thread int) *Batch {
	var (
		mappers  = make([]map[string]database.Access, thread)
		txs      = make([]*tx.Tx, thread)
		receipts = make([]*tx.TxReceipt, thread)
		visitors = make([]*database.Visitor, thread)
		errs     = make([]error, thread)
	)

	bvr := database.NewBatchVisitorRoot(10000, db)
	for i := 0; i < thread; i++ {
		i2 := i
		t := provider.Tx()
		if t == nil {
			break
		}
		if !t.IsTimeValid(bh.Time) && !t.IsDefer() {
			i--
			provider.Drop(t, ErrInvalidTimeTx)
			continue
		}
		go func() {
			vi, mapper := database.NewBatchVisitor(bvr)

			e := vm.Isolator{}
			err := e.Prepare(bh, vi, &ilog.Logger{})
			if err != nil {
				return
			}

			m.wait.Add(1)
			defer m.wait.Done()

			e.PrepareTx(t, limit)

			tr, err := e.Run()

			if err == nil {
				mappers[i2] = mapper.Map()
			} else {
				mappers[i2] = nil
			}
			txs[i2] = t
			receipts[i2] = tr
			visitors[i2] = vi
			errs[i2] = err
		}()
	}
	m.wait.Wait()
	ti, td := Resolve(mappers)

	for _, i := range td {
		provider.Return(txs[i])
	}

	for i := range txs {
		if errs[i] != nil {
			provider.Drop(txs[i], errs[i])
		}
	}

	b := NewBatch()

	for _, i := range ti {
		b.Txs = append(b.Txs, txs[i])
		b.Receipts = append(b.Receipts, receipts[i])
		visitors[i].Commit()
	}

	return b
}

// Resolve Resolve conflict of parallel exec
func Resolve(mappers []map[string]database.Access) (accept, drop []int) {
	workMap := make(map[string]database.Access)
	accept = make([]int, 0)
	drop = make([]int, 0)
L:
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
				continue L
			}
		}
		accept = append(accept, i)
	}
	return
}

// Verify use check function to verify batch
func (m *batcherImpl) Verify(bh *block.BlockHead, db database.IMultiValue, checkFunc func(e vm.Isolator, t *tx.Tx, r *tx.TxReceipt) error, b *Batch) error {
	var (
		thread  = len(b.Txs)
		mappers = make([]map[string]database.Access, thread)
		errs    = make([]error, thread)
	)

	bvr := database.NewBatchVisitorRoot(10000, db)
	for i := 0; i < thread; i++ {
		i2 := i
		go func() {
			vi, mapper := database.NewBatchVisitor(bvr)

			e := vm.Isolator{}
			err := e.Prepare(bh, vi, &ilog.Logger{})
			if err != nil {
				return
			}

			m.wait.Add(1)
			defer m.wait.Done()

			t := b.Txs[i2]
			errs[i2] = checkFunc(e, t, b.Receipts[i2])
			if errs[i2] == nil {
				mappers[i2] = mapper.Map()
			}

		}()
	}
	m.wait.Wait()
	_, td := Resolve(mappers)
	if len(td) != 0 {
		return fmt.Errorf("transaction conflicted")
	}
	for _, e := range errs {
		if e != nil {
			return e
		}
	}
	return nil
}
