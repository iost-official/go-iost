package vm

import (
	"time"

	"strings"

	"github.com/iost-official/go-iost/account"
	"github.com/iost-official/go-iost/core/block"
	"github.com/iost-official/go-iost/core/contract"
	"github.com/iost-official/go-iost/core/tx"
	"github.com/iost-official/go-iost/ilog"
	"github.com/iost-official/go-iost/vm/database"
	"github.com/iost-official/go-iost/vm/host"
	"github.com/iost-official/go-iost/vm/native"
)

// Isolator new entrance instead of Engine
type Isolator struct {
	h            *host.Host
	publisherID  string
	t            *tx.Tx
	blockBaseCtx *host.Context
}

// Prepare Isolator
func (e *Isolator) Prepare(bh *block.BlockHead, db *database.Visitor, logger *ilog.Logger) error {
	if db.Contract("iost.system") == nil {
		db.SetContract(native.SystemABI())
	}

	e.blockBaseCtx = host.NewContext(nil)
	e.blockBaseCtx = loadBlkInfo(e.blockBaseCtx, bh)
	e.h = host.NewHost(e.blockBaseCtx, db, staticMonitor, logger)
	return nil
}

// PrepareTx read tx and ready to run
func (e *Isolator) PrepareTx(t *tx.Tx, limit time.Duration) error {
	e.t = t
	e.h.SetDeadline(time.Now().Add(limit))
	err := checkTxParams(t)
	if err != nil {
		return err
	}
	e.publisherID = account.GetIDByPubkey(t.PublishSign.Pubkey)
	bl := e.h.DB().Balance(e.publisherID)
	if bl < 0 || bl < t.GasPrice*t.GasLimit {
		return errCannotPay
	}

	loadTxInfo(e.h, t, e.publisherID)
	return nil
}

func (e *Isolator) runAction(action tx.Action) (cost *contract.Cost, status tx.Status, receipts []tx.Receipt, err error) {
	receipts = make([]tx.Receipt, 0)

	e.h.PushCtx()
	defer func() {
		e.h.PopCtx()
	}()

	e.h.Context().Set("stack0", "direct_call")
	e.h.Context().Set("stack_height", 1) // record stack trace

	_, cost, err = staticMonitor.Call(e.h, action.Contract, action.ActionName, action.Data)

	if cost == nil {
		panic("cost is nil")
	}

	if err != nil {

		if strings.Contains(err.Error(), "execution killed") {
			status = tx.Status{
				Code:    tx.ErrorTimeout,
				Message: err.Error(),
			}
		} else {
			status = tx.Status{
				Code:    tx.ErrorRuntime,
				Message: err.Error(),
			}
		}

		receipt := tx.Receipt{
			Type:    tx.SystemDefined,
			Content: err.Error(),
		}
		receipts = append(receipts, receipt)

		err = nil

		return
	}

	receipts = append(receipts, e.h.Context().GValue("receipts").([]tx.Receipt)...)

	status = tx.Status{
		Code:    tx.Success,
		Message: "",
	}
	return
}

// Run actions in tx
func (e *Isolator) Run() (*tx.TxReceipt, error) {
	e.h.Context().GSet("gas_limit", e.t.GasLimit)
	e.h.Context().GSet("receipts", make([]tx.Receipt, 0))

	txr := tx.NewTxReceipt(e.t.Hash())
	hasSetCode := false

	for _, action := range e.t.Actions {
		if hasSetCode && action.Contract == "iost.system" && action.ActionName == "SetCode" {
			txr.Receipts = nil
			txr.Status.Code = tx.ErrorDuplicateSetCode
			txr.Status.Message = "error duplicate set code in a tx"
			break
		}
		hasSetCode = action.Contract == "iost.system" && action.ActionName == "SetCode"

		cost, status, receipts, err := e.runAction(*action)
		ilog.Debugf("run action : %v, result is %v", action, status.Code)
		ilog.Debug("used cost > ", cost)
		ilog.Debugf("status > \n%v\n", status)

		if err != nil {
			return nil, err
		}

		gasLimit := e.h.Context().GValue("gas_limit").(int64)

		txr.Status = status
		if (status.Code == 4 && status.Message == "out of gas") || (status.Code == 5) {
			cost = contract.NewCost(0, 0, gasLimit)
		}

		txr.GasUsage += cost.ToGas()

		e.h.Context().GSet("gas_limit", gasLimit-cost.ToGas())

		e.h.PayCost(cost, e.publisherID)

		if status.Code != tx.Success {
			txr.Receipts = nil
			break
		} else {
			txr.Receipts = append(txr.Receipts, receipts...)
			txr.SuccActionNum++
		}
	}
	return &txr, nil
}

// PayCost as name
func (e *Isolator) PayCost() error {
	err := e.h.DoPay(e.h.Context().Value("witness").(string), e.t.GasPrice)
	if err != nil {
		e.h.DB().Rollback()
		err = e.h.DoPay(e.h.Context().Value("witness").(string), e.t.GasPrice)
		if err != nil {
			return err
		}
	}
	return nil
}

// Commit flush changes to db
func (e *Isolator) Commit() {
	e.h.DB().Commit()
}

// ClearAll clear this isolator
func (e *Isolator) ClearAll() {
	e.h = nil
}

// ClearTx clear this tx
func (e *Isolator) ClearTx() {
	e.h.SetContext(e.blockBaseCtx)
	e.h.Context().GClear()
}
func checkTxParams(t *tx.Tx) error {
	if t.GasPrice < 0 || t.GasPrice > 10000 {
		return errGasPriceIllegal
	}
	return nil
}
