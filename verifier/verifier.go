package verifier

import (
	"errors"
	"fmt"
	"time"

	"github.com/iost-official/go-iost/v3/core/blockcache"
	"github.com/iost-official/go-iost/v3/core/global"

	"github.com/iost-official/go-iost/v3/common"
	"github.com/iost-official/go-iost/v3/core/block"
	"github.com/iost-official/go-iost/v3/core/tx"
	"github.com/iost-official/go-iost/v3/core/txpool"
	"github.com/iost-official/go-iost/v3/ilog"
	"github.com/iost-official/go-iost/v3/vm"
	"github.com/iost-official/go-iost/v3/vm/database"
)

// values
var (
	ErrExpiredTx    = errors.New("expired tx")
	ErrNotArrivedTx = errors.New("not arrived tx")
	ErrInvalidMode  = errors.New("invalid mode")
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

type Provider interface {
	Tx() *tx.Tx
	Return(*tx.Tx)
	Drop(t *tx.Tx, err error)
	Close()
}

// Exec exec single tx and flush changes to db
func (v *Verifier) Exec(bh *block.BlockHead, db database.IMultiValue, t *tx.Tx, limit time.Duration) (*tx.TxReceipt, error) {
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
func (v *Verifier) Try(bh *block.BlockHead, db database.IMultiValue, t *tx.Tx, limit time.Duration) (*tx.TxReceipt, error) {
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
	return r, err
}

// Gen gen block
func (v *Verifier) Gen(blk, parent *block.Block, witnessList *blockcache.WitnessList, db database.IMultiValue, iter *txpool.SortedTxMap, c *Config) (droplist []*tx.Tx, errs []error, err error) {
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
		if t.Delay > 0 {
			ilog.Debug("Ignore delay tx.")
			continue L
		}
		if tx.CheckBadTx(t) != nil {
			ilog.Errorf("bad tx %v", t)
			continue L
		}
		if t.IsExpired(blk.Head.Time) && !t.IsDefer() {
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

// Verify verify block generated by Verifier
func (v *Verifier) Verify(blk, parent *block.Block, witnessList *blockcache.WitnessList, db database.IMultiValue, c *Config) error {
	err := verifyBlockBase(blk, parent, witnessList, db, c)
	if err != nil {
		return err
	}
	isolator := vm.Isolator{}
	vi, _ := database.NewBatchVisitor(database.NewBatchVisitorRoot(100, db, blk.Head.Rules()))
	isolator.Prepare(blk.Head, vi, getLogger(false))
	return baseVerify(isolator, c, blk.Txs[1:], blk.Receipts[1:], blk)
}
func verifyBlockBase(blk, parent *block.Block, witnessList *blockcache.WitnessList, db database.IMultiValue, c *Config) error {
	if len(blk.Txs) < 1 || len(blk.Receipts) < 1 {
		return fmt.Errorf("block did not contain block base tx")
	}
	baseTx, err := NewBaseTx(blk, parent, witnessList)
	if err != nil {
		return err
	}
	for i, a := range blk.Txs[0].Actions {
		if a.ActionName != baseTx.Actions[i].ActionName ||
			a.Contract != baseTx.Actions[i].Contract ||
			a.Data != baseTx.Actions[i].Data {
			ilog.Warnf("witnessList: %+v", witnessList)
			return fmt.Errorf("block base tx not match, verifyBaseTxAction: %+v\n, localBaseTxAction: %+v", a, baseTx.Actions[i])
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
	err = checkReceiptEqual(blk.Receipts[0], r)
	if err != nil {
		return err
	}

	return nil
}

func verify(isolator vm.Isolator, t *tx.Tx, r *tx.TxReceipt, timeout time.Duration, isBlockBase bool, blk *block.Block) error { // nolint
	if !t.IsCreatedBefore(blk.Head.Time) {
		return ErrNotArrivedTx
	}
	if t.IsExpired(blk.Head.Time) && !t.IsDefer() {
		return ErrExpiredTx
	}
	isolator.ClearTx()
	if isBlockBase {
		isolator.TriggerBlockBaseMode()
	}
	var to time.Duration
	if r.Status.Code == tx.ErrorTimeout {
		if blk.Head.Rules().IsFork3_0_10 {
			to = 0
		} else {
			to = timeout / 2
		}
	} else {
		to = timeout * 50
	}
	err := isolator.PrepareTx(t, to)
	if err != nil {
		return err
	}
	_, err = isolator.Run()
	if err != nil {
		return err
	}
	receipt, err := isolator.PayCost()
	if err != nil {
		return err
	}
	err = checkReceiptEqual(r, receipt)
	if err != nil {
		return err
	}
	isolator.Commit()
	return nil
}

func checkReceiptEqual(r *tx.TxReceipt, receipt *tx.TxReceipt) error {
	if r.Status.Code != receipt.Status.Code {
		return fmt.Errorf("receipt not match, status not same: %v != %v \n%v\n%v", r.Status, receipt.Status, r, receipt)
	}
	if r.Status.Code == tx.Success {
		if r.Status.Message != receipt.Status.Message {
			return fmt.Errorf("receipt not match, status not same: %v != %v \n%v\n%v", r.Status, receipt.Status, r, receipt)
		}
	}
	if r.GasUsage != receipt.GasUsage {
		return fmt.Errorf("receipt not match, gas usage not same: %v != %v \n%v\n%v", r.GasUsage, receipt.GasUsage, r, receipt)
	}
	if len(r.RAMUsage) != len(receipt.RAMUsage) {
		return fmt.Errorf("receipt not match, ram usage length not same: %v != %v \n%v\n%v", len(r.RAMUsage), len(receipt.RAMUsage), r, receipt)
	}
	for k, v := range r.RAMUsage {
		if v != receipt.RAMUsage[k] {
			return fmt.Errorf("receipt not match, ram usage not same: %v != %v \n%v\n%v", v, receipt.RAMUsage[k], r, receipt)
		}
	}
	if len(r.Receipts) != len(receipt.Receipts) {
		return fmt.Errorf("receipt not match, receipts length not same: %v != %v \n%v\n%v", len(r.Receipts), len(receipt.Receipts), r, receipt)
	}
	for i, br := range r.Receipts {
		if br.FuncName != receipt.Receipts[i].FuncName {
			return fmt.Errorf("receipt not match, funcname not same: %v != %v \n%v\n%v", br.FuncName, receipt.Receipts[i].FuncName, r, receipt)
		}
		if br.Content != receipt.Receipts[i].Content {
			return fmt.Errorf("receipt not match, content not same: %v != %v \n%v\n%v", br.Content, receipt.Receipts[i].Content, r, receipt)
		}
	}
	if len(r.Returns) != len(receipt.Returns) {
		return fmt.Errorf("receipt not match, returns length not same: %v != %v \n%v\n%v", len(r.Returns), len(receipt.Returns), r, receipt)
	}
	for i, br := range r.Returns {
		if br != receipt.Returns[i] {
			return fmt.Errorf("receipt not match, returns not same: %v != %v \n%v\n%v", br, receipt.Returns[i], r, receipt)
		}
	}
	return nil
}

func baseVerify(engine vm.Isolator, c *Config, txs []*tx.Tx, receipts []*tx.TxReceipt, blk *block.Block) error {
	blockGasLimit := common.MaxBlockGasLimit
	blockGas := int64(0)
	for _, r := range receipts {
		blockGas += r.GasUsage
	}
	if blockGas > blockGasLimit {
		return fmt.Errorf(
			"Block %v include gas %v, exceeds maximum limit %v",
			common.Base58Encode(blk.HeadHash()),
			blockGas/100,
			blockGasLimit/100,
		)
	}

	for k, t := range txs {
		err := verify(engine, t, receipts[k], c.TxTimeLimit, false, blk)
		if err != nil {
			return err
		}
	}
	return nil
}
