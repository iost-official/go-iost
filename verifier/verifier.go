package verifier

import (
	"time"

	"encoding/json"

	"fmt"

	"github.com/iost-official/go-iost/core/block"
	"github.com/iost-official/go-iost/core/tx"
	"github.com/iost-official/go-iost/vm"
	"github.com/iost-official/go-iost/vm/database"
)

type Verifier struct {
}

type Config struct {
	Mode        int
	Timeout     time.Duration
	TxTimeLimit time.Duration
	Thread      int
}

type Info struct {
	Mode   int   `json:"mode"`
	Thread int   `json:"thread"`
	Batch  []int `json:"batch"`
}

//var ParallelMask int64 = 1 // 0000 0001

func (v *Verifier) Gen(blk *block.Block, db database.IMultiValue, provider vm.TxIter, c *Config) (droplist []*tx.Tx, err []error, err2 error) {
	switch c.Mode {
	case 0:
		e := vm.NewEngine(blk.Head, db)
		return baseGen(blk, db, provider, e, c),
	case 1:
		batcher := vm.NewBatcher()
		return batchGen(blk, db, provider, batcher, c)
	}
	return []*tx.Tx{} ,[]error{}, fmt.Errorf("mode unexpected: %v", c.Mode)
}

func baseGen(blk *block.Block, db database.IMultiValue, provider vm.Provider, engine vm.Engine, c *Config) error {
	var err error

	blk.Txs = make([]*tx.Tx, 0)
	blk.Receipts = make([]*tx.TxReceipt, 0)
	info := Info{
		Mode: 1,
	}
	to := time.After(c.Timeout)
L:
	for {
		select {
		case <-to:
			break L
		default:
			t := provider.Tx()
			if t == nil {
				break L
			}
			var r *tx.TxReceipt
			r, err = engine.Exec(t, c.TxTimeLimit)
			if err != nil {
				fmt.Println(err)
				continue L
			}
			blk.Txs = append(blk.Txs, t)
			blk.Receipts = append(blk.Receipts, r)
		}
	}
	buf, err := json.Marshal(info)
	blk.Head.Info = buf
	return err
}

func batchGen(blk *block.Block, db database.IMultiValue, provider vm.Provider, batcher vm.Batcher, c *Config) error {
	var err error

	blk.Txs = make([]*tx.Tx, 0)
	blk.Receipts = make([]*tx.TxReceipt, 0)
	info := Info{
		Mode:   1,
		Thread: c.Thread,
		Batch:  make([]int, 0),
	}
	to := time.After(c.Timeout)
L:
	for {
		select {
		case <-to:
			break L
		default:
			batch := batcher.Batch(blk.Head, db, provider, time.Duration(c.TxTimeLimit), c.Thread)

			info.Batch = append(info.Batch, len(batch.Txs))
			for i, t := range batch.Txs {
				blk.Txs = append(blk.Txs, t)
				blk.Receipts = append(blk.Receipts, batch.Receipts[i])
			}
		}

	}

	blk.Head.Info, err = json.Marshal(info)

	return err
}

func (v *Verifier) Verify(blk *block.Block, db database.IMultiValue, c *Config) error {
	ri := blk.Head.Info
	var info Info
	err := json.Unmarshal(ri, &info)
	if err != nil {
		return err
	}
	switch info.Mode {
	case 0:
		e := vm.NewEngine(blk.Head, db)
		return baseVerify(e, c, blk.Txs, blk.Receipts)
	case 1:
		bs := batches(blk, info)
		var batcher vm.Batcher
		return batchVerify(batcher, blk.Head, c, db, bs)
	}
	return nil
}

func batches(blk *block.Block, info Info) []*vm.Batch {
	var rtn = make([]*vm.Batch, 0)
	var k = 0
	for _, j := range info.Batch {
		txs := blk.Txs[k:j]
		rs := blk.Receipts[k:j]
		k = j
		rtn = append(rtn, &vm.Batch{
			Txs:      txs,
			Receipts: rs,
		})
	}
	return rtn
}

func verify(e vm.Engine, t *tx.Tx, r *tx.TxReceipt, timeout time.Duration) error {
	var to time.Duration
	if r.Status.Code == tx.ErrorTimeout {
		to = timeout / 2
	} else {
		to = timeout * 2
	}
	receipt, err := e.Exec(t, to)
	if err != nil {
		return err
	}
	if r.Status != receipt.Status ||
		r.GasUsage != receipt.GasUsage ||
		r.SuccActionNum != receipt.SuccActionNum {
		return fmt.Errorf("receipt not match: %v, %v", r, receipt)
	}
	return nil
}

func baseVerify(engine vm.Engine, c *Config, txs []*tx.Tx, receipts []*tx.TxReceipt) error {
	for k, t := range txs {
		err := verify(engine, t, receipts[k], c.TxTimeLimit)
		if err != nil {
			return err
		}
	}
	return nil
}

func batchVerify(verifier vm.Batcher, bh *block.BlockHead, c *Config, db database.IMultiValue, batches []*vm.Batch) error {
	for _, batch := range batches {
		err := verifier.Verify(bh, db, func(e vm.Engine, t *tx.Tx, r *tx.TxReceipt) error {
			err := verify(e, t, r, c.TxTimeLimit)
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
