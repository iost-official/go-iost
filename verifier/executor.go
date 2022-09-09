package verifier

import (
	"fmt"
	"time"

	"github.com/iost-official/go-iost/v3/common"
	"github.com/iost-official/go-iost/v3/core/block"
	"github.com/iost-official/go-iost/v3/core/blockcache"
	"github.com/iost-official/go-iost/v3/core/global"
	"github.com/iost-official/go-iost/v3/core/tx"
	"github.com/iost-official/go-iost/v3/core/txpool"
	"github.com/iost-official/go-iost/v3/ilog"
	"github.com/iost-official/go-iost/v3/vm"
	"github.com/iost-official/go-iost/v3/vm/database"
)

type Executor struct {
}

// Exec exec single tx and flush changes to db
func (v *Executor) Exec(bh *block.BlockHead, db database.IMultiValue, t *tx.Tx, limit time.Duration) (*tx.TxReceipt, error) {
	var isolator vm.Isolator
	vi := database.NewVisitor(100, db, bh.Rules())
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
func (v *Executor) Try(bh *block.BlockHead, db database.IMultiValue, t *tx.Tx, limit time.Duration) (*tx.TxReceipt, error) {
	var isolator vm.Isolator
	vi := database.NewVisitor(100, db, bh.Rules())
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
	_, err = isolator.Run()
	if err != nil {
		return &tx.TxReceipt{}, err
	}
	r, err := isolator.PayCost()
	return r, err
}

// Gen gen block
func (v *Executor) Gen(blk, parent *block.Block, witnessList *blockcache.WitnessList, db database.IMultiValue, iter *txpool.SortedTxMap, c *Config) (droplist []*tx.Tx, errs []error, err error) {
	isolator := &vm.Isolator{}
	baseTx, err := NewBaseTx(blk, parent, witnessList)
	if err != nil {
		return nil, nil, err
	}
	r, err := blockBaseExec(blk, db, isolator, baseTx, c)
	if err != nil {
		return nil, nil, err
	}
	blk.Txs = append(blk.Txs, baseTx)
	blk.Receipts = append(blk.Receipts, r)
	var pi = NewProvider(iter)
	err = baseGen(blk, db, pi, isolator, c)
	droplist, errs = pi.List()
	pi.Close()
	return
}

func blockBaseExec(blk *block.Block, db database.IMultiValue, isolator *vm.Isolator, t *tx.Tx, c *Config) (tr *tx.TxReceipt, err error) {
	vi := database.NewVisitor(100, db, blk.Head.Rules())
	isolator.Prepare(blk.Head, vi, getLogger(global.GetGlobalConf() != nil && global.GetGlobalConf().Log != nil && global.GetGlobalConf().Log.EnableContractLog))
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

// nolint:gocyclo
func baseGen(blk *block.Block, db database.IMultiValue, provider Provider, isolator *vm.Isolator, c *Config) (err error) {
	var tn time.Time
	to := time.Now().Add(c.Timeout)
	blockGasLimit := common.MaxBlockGasLimit

L:
	for tn.Before(to) {
		isolator.ClearTx()
		tn = time.Now()
		limit := to.Sub(tn)
		if limit > c.TxTimeLimit {
			limit = c.TxTimeLimit
		}
		if limit < 500*time.Microsecond {
			break L
		}
		t := provider.Tx()
		if t == nil {
			break L
		}
		if !t.IsCreatedBefore(blk.Head.Time) {
			ilog.Debugf(
				"Tx %v has not arrived. tx time is %v, blk time is %v",
				common.Base58Encode(t.Hash()),
				t.Time,
				blk.Head.Time,
			)
			continue L
		}
		if t.Delay != 0 {
			ilog.Debug("Ignore delay tx.")
			continue L
		}
		if tx.CheckBadTx(t) != nil {
			ilog.Errorf("bad tx %v", t)
			continue L
		}
		if t.IsExpired(blk.Head.Time) {
			ilog.Errorf(
				"Tx %v is expired, tx time is %v, tx expiration time is %v, blk time is %v",
				common.Base58Encode(t.Hash()),
				t.Time,
				t.Expiration,
				blk.Head.Time,
			)
			provider.Drop(t, ErrExpiredTx)
			continue L
		}
		if t.GasLimit > blockGasLimit {
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
			ilog.Errorf("isolator run error %v %v", t.String(), err)
			provider.Drop(t, err)
			continue L
		}
		if r.Status.Code == tx.ErrorTimeout && limit < c.TxTimeLimit {
			ilog.Debugf(
				"isolator run time out, but time limit %v less than std time limit %v",
				limit,
				c.TxTimeLimit,
			)
			provider.Return(t)
			break L
		}
		//ilog.Debugf("exec tx %v success", common.Base58Encode(t.Hash()))
		r, err = isolator.PayCost()
		if err != nil {
			ilog.Errorf("pay cost err %v %v", t.String(), err)
			provider.Drop(t, err)
			continue L
		}
		isolator.Commit()
		blk.Txs = append(blk.Txs, t)
		blk.Receipts = append(blk.Receipts, r)
		blockGasLimit -= r.GasUsage
	}
	blk.Head.Info = []byte(`{}`) // for legacy reasons. remove it in later version
	for _, t := range blk.Txs {
		provider.Drop(t, nil)
	}
	return err
}

func getLogger(enableContractLog bool) *ilog.Logger {
	if !enableContractLog {
		var l ilog.Logger
		l.Stop()
		return &l
	}
	return ilog.DefaultLogger()
}
