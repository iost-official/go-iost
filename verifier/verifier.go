package verifier

import (
	"errors"
	"time"

	"encoding/json"

	"fmt"

	"github.com/iost-official/go-iost/core/block"
	"github.com/iost-official/go-iost/core/tx"
	"github.com/iost-official/go-iost/ilog"
	"github.com/iost-official/go-iost/vm"
	"github.com/iost-official/go-iost/vm/database"
)

// values
var (
	ErrInvalidTimeTx = errors.New("invalid time tx")
)

// Verifier ..
type Verifier struct {
}

// Config config of verifier
type Config struct {
	Mode        int
	Timeout     time.Duration
	TxTimeLimit time.Duration
	Thread      int
}

// Info info in block
type Info struct {
	Mode   int   `json:"mode"`
	Thread int   `json:"thread"`
	Batch  []int `json:"batch"`
}

//var ParallelMask int64 = 1 // 0000 0001

// Exec exec single tx and flush changes to db
func (v *Verifier) Exec(bh *block.BlockHead, db database.IMultiValue, t *tx.Tx, limit time.Duration) (*tx.TxReceipt, error) {
	var isolator vm.Isolator
	vi := database.NewVisitor(100, db)
	var l ilog.Logger
	l.Stop()
	err := isolator.Prepare(bh, vi, &l)
	if err != nil {
		return &tx.TxReceipt{}, err
	}
	err = isolator.PrepareTx(t, limit)

	if err != nil {
		return &tx.TxReceipt{}, err
	}
	r, err := isolator.Run()
	if err != nil {
		return &tx.TxReceipt{}, err
	}
	isolator.Commit()
	return r, err
}

// Try exec tx and only return receipt
func (v *Verifier) Try(bh *block.BlockHead, db database.IMultiValue, t *tx.Tx, limit time.Duration) (*tx.TxReceipt, error) {
	var isolator vm.Isolator
	vi := database.NewVisitor(100, db)
	var l ilog.Logger
	l.Stop()
	err := isolator.Prepare(bh, vi, &l)
	if err != nil {
		return &tx.TxReceipt{}, err
	}
	err = isolator.PrepareTx(t, limit)
	if err != nil {
		return &tx.TxReceipt{}, err
	}
	r, err := isolator.Run()
	if err != nil {
		return &tx.TxReceipt{}, err
	}
	return r, err
}

// Gen gen block
func (v *Verifier) Gen(blk *block.Block, db database.IMultiValue, iter TxIter, c *Config) (droplist []*tx.Tx, errs []error, err error) {
	isolator := &vm.Isolator{}
	r, err := blockBaseExec(blk, db, isolator, BlockBaseTx, c)
	if err != nil {
		return nil, nil, err
	}
	blk.Txs = append(blk.Txs, BlockBaseTx)
	blk.Receipts = append(blk.Receipts, r)
	var pi = NewProvider(iter)
	switch c.Mode {
	case 0:
		err = baseGen(blk, db, pi, isolator, c)
		droplist, errs = pi.List()
		return
	case 1:
		batcher := NewBatcher()
		err = batchGen(blk, db, pi, batcher, c)
		droplist, errs = pi.List()
		return
	}
	return []*tx.Tx{}, []error{}, fmt.Errorf("mode unexpected: %v", c.Mode)
}

func blockBaseExec(blk *block.Block, db database.IMultiValue, isolator *vm.Isolator, t *tx.Tx, c *Config) (tr *tx.TxReceipt, err error) {
	vi := database.NewVisitor(100, db)
	var l ilog.Logger
	l.Stop()
	isolator.Prepare(blk.Head, vi, &l)
	isolator.TriggerBlockBaseMode()
	err = isolator.PrepareTx(t, c.Timeout)
	if err != nil {
		return nil, err
	}
	r, err := isolator.Run()
	if err != nil {
		return nil, err
	}
	if r.Status.Code != tx.Success {
		return nil, fmt.Errorf(r.Status.Message)
	}
	isolator.Commit()
	isolator.ClearTx()

	return r, nil
}

func baseGen(blk *block.Block, db database.IMultiValue, provider Provider, isolator *vm.Isolator, c *Config) (err error) {
	info := Info{
		Mode: 0,
	}
	var tn time.Time
	to := time.Now().Add(c.Timeout)

L:
	for tn.Before(to) {
		isolator.ClearTx()
		tn = time.Now()

		limit := to.Sub(tn)
		if limit > c.TxTimeLimit {
			limit = c.TxTimeLimit
		}
		t := provider.Tx()
		if t == nil {
			break L
		}
		if !t.IsTimeValid(blk.Head.Time) && !t.IsDefer() {
			provider.Drop(t, ErrInvalidTimeTx)
			continue L
		}
		err := isolator.PrepareTx(t, limit)
		if err != nil {
			ilog.Errorf("PrepareTx failed. tx %v limit %v err %v", t.String(), limit, err)
			provider.Drop(t, err)
			continue L
		}
		var r *tx.TxReceipt
		r, err = isolator.Run()
		if err != nil {
			ilog.Errorf("isolator run error %v", err)
			provider.Drop(t, err)
			continue L
		}
		if r.Status.Code == tx.ErrorTimeout && limit < c.TxTimeLimit {
			ilog.Errorf("isolator run time out")
			provider.Return(t)
			break L
		}
		r, err = isolator.PayCost()
		if err != nil {
			ilog.Errorf("pay cost err %v", err)
			provider.Drop(t, err)
			continue L
		}
		isolator.Commit()
		blk.Txs = append(blk.Txs, t)
		blk.Receipts = append(blk.Receipts, r)
	}
	buf, err := json.Marshal(info)
	if err != nil {
		panic(err)
	}
	blk.Head.Info = buf
	for _, t := range blk.Txs {
		provider.Drop(t, nil)
	}
	return err
}

func batchGen(blk *block.Block, db database.IMultiValue, provider Provider, batcher Batcher, c *Config) (err error) {
	info := Info{
		Mode:   1,
		Thread: c.Thread,
		Batch:  make([]int, 0),
	}
	tn := time.Now()
	to := time.Now().Add(c.Timeout)
	for tn.Before(to) {
		limit := to.Sub(tn)
		if limit > c.TxTimeLimit {
			limit = c.TxTimeLimit
		}
		batch := batcher.Batch(blk.Head, db, provider, limit, c.Thread)

		info.Batch = append(info.Batch, len(batch.Txs))
		for i, t := range batch.Txs {
			if limit < c.TxTimeLimit && batch.Receipts[i].Status.Code == 5 {
				provider.Return(t)
				continue
			}
			blk.Txs = append(blk.Txs, t)
			blk.Receipts = append(blk.Receipts, batch.Receipts[i])
			provider.Drop(t, nil)
		}
	}

	blk.Head.Info, err = json.Marshal(info)

	return err
}

// Verify verify block generated by Verifier
func (v *Verifier) Verify(blk *block.Block, db database.IMultiValue, c *Config) error {
	ri := blk.Head.Info
	var info Info
	err := json.Unmarshal(ri, &info)
	if err != nil {
		return err
	}

	err = verifyBlockBase(blk, db, c)
	if err != nil {
		return err
	}
	switch info.Mode {
	case 0:
		isolator := vm.Isolator{}
		vi, _ := database.NewBatchVisitor(database.NewBatchVisitorRoot(100, db))
		var l ilog.Logger
		l.Stop()
		isolator.Prepare(blk.Head, vi, &l)
		return baseVerify(isolator, c, blk.Txs[1:], blk.Receipts[1:], blk)
	case 1:
		bs := batches(blk, info)
		var batcher Batcher
		return batchVerify(batcher, blk.Head, c, db, bs, blk)
	}
	return nil
}

func batches(blk *block.Block, info Info) []*Batch {
	var rtn = make([]*Batch, 0)
	var k = 0
	for _, j := range info.Batch {
		txs := blk.Txs[k:j]
		rs := blk.Receipts[k:j]
		k = j
		rtn = append(rtn, &Batch{
			Txs:      txs,
			Receipts: rs,
		})
	}
	return rtn
}

func verifyBlockBase(blk *block.Block, db database.IMultiValue, c *Config) error {
	if len(blk.Txs) < 1 || len(blk.Receipts) < 1 {
		return fmt.Errorf("block did not contain block base tx")
	}

	for i, a := range blk.Txs[0].Actions {
		if a.ActionName != BlockBaseTx.Actions[i].ActionName ||
			a.Contract != BlockBaseTx.Actions[i].Contract ||
			a.Data != BlockBaseTx.Actions[i].Data {
			return fmt.Errorf("block base tx not match")
		}
	}
	isolator := &vm.Isolator{}
	r, err := blockBaseExec(blk, db, isolator, blk.Txs[0], c)
	if err != nil {
		return err
	}
	if r.Status.Code != tx.Success {
		return fmt.Errorf("block base tx receipt error: %v", r.Status.Message)
	}

	return nil
}

func verify(isolator vm.Isolator, t *tx.Tx, r *tx.TxReceipt, timeout time.Duration, isBlockBase bool, blk *block.Block) error { // nolint
	if !t.IsTimeValid(blk.Head.Time) && !t.IsDefer() {
		return ErrInvalidTimeTx
	}
	isolator.ClearTx()
	if isBlockBase {
		isolator.TriggerBlockBaseMode()
	}
	var to time.Duration
	if r.Status.Code == tx.ErrorTimeout {
		to = timeout / 2
	} else {
		to = timeout * 2
	}
	err := isolator.PrepareTx(t, to)
	if err != nil {
		return err
	}
	receipt, err := isolator.Run()
	if err != nil {
		return err
	}
	if r.Status != receipt.Status ||
		r.GasUsage != receipt.GasUsage {
		return fmt.Errorf("receipt not match: %v, %v", r, receipt)
	}
	for k, v := range r.RAMUsage {
		if v != receipt.RAMUsage[k] {
			return fmt.Errorf("receipt not match: %v, %v", r, receipt)
		}
	}
	for i, br := range r.Receipts {
		if br.FuncName != receipt.Receipts[i].FuncName ||
			br.Content != receipt.Receipts[i].Content {
			return fmt.Errorf("receipt not match: %v, %v", r, receipt)
		}
	}
	for i, br := range r.Returns {
		if br.FuncName != receipt.Receipts[i].FuncName ||
			br.Value != receipt.Receipts[i].Content {
			return fmt.Errorf("receipt not match: %v, %v", r, receipt)
		}
	}

	isolator.Commit()
	return nil
}

func baseVerify(engine vm.Isolator, c *Config, txs []*tx.Tx, receipts []*tx.TxReceipt, blk *block.Block) error {
	for k, t := range txs {
		err := verify(engine, t, receipts[k], c.TxTimeLimit, false, blk)
		if err != nil {
			return err
		}
	}
	return nil
}

func batchVerify(verifier Batcher, bh *block.BlockHead, c *Config, db database.IMultiValue, batches []*Batch, blk *block.Block) error {

	for _, batch := range batches {
		err := verifier.Verify(bh, db, func(e vm.Isolator, t *tx.Tx, r *tx.TxReceipt) error {
			err := verify(e, t, r, c.TxTimeLimit, false, blk)
			if err != nil {
				return err
			}
			return nil
		}, batch)
		if err != nil {
			return err
		}
	}
	return nil
}
