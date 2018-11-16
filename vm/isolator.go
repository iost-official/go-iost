package vm

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/iost-official/go-iost/account"
	"github.com/iost-official/go-iost/common"
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
	h             *host.Host
	publisherID   string
	t             *tx.Tx
	tr            *tx.TxReceipt
	blockBaseCtx  *host.Context
	genesisMode   bool
	blockBaseMode bool
}

// TriggerBlockBaseMode start blockbase mode
func (e *Isolator) TriggerBlockBaseMode() {
	e.blockBaseMode = true
}

// Prepare Isolator
func (e *Isolator) Prepare(bh *block.BlockHead, db *database.Visitor, logger *ilog.Logger) error {
	if db.Contract("iost.system") == nil {
		db.SetContract(native.SystemABI())
	}
	if bh.Number == 0 {
		e.genesisMode = true
	} else {
		e.genesisMode = false
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
	e.publisherID = t.Publisher
	l := len(t.Encode())
	e.h.PayCost(contract.NewCost(0, int64(l), 0), t.Publisher)

	if !e.genesisMode && !e.blockBaseMode {
		err := checkTxParams(t)
		if err != nil {
			return err
		}
		gas, _ := e.h.CurrentGas(e.publisherID)
		price := &common.Fixed{Value: t.GasPrice, Decimal: 2}
		if gas.LessThan(price.Times(t.GasLimit)) {
			ilog.Infof("err %v publisher %v current gas %v price %v limit %v\n", errCannotPay, e.publisherID, gas.ToString(), t.GasPrice, t.GasLimit)
			return fmt.Errorf("%v gas less than price * limit %v < %v * %v", e.publisherID, gas.ToString(), price.ToString(), t.GasLimit)
		}
	}
	loadTxInfo(e.h, t, e.publisherID)
	return nil
}

func (e *Isolator) runAction(action tx.Action) (cost contract.Cost, status *tx.Status, ret string, receipts []*tx.Receipt, err error) {
	receipts = make([]*tx.Receipt, 0)

	e.h.PushCtx()
	defer func() {
		e.h.PopCtx()
	}()

	e.h.Context().Set("stack0", "direct_call")
	e.h.Context().Set("stack_height", 1) // record stack trace

	var rtn []interface{}

	rtn, cost, err = staticMonitor.Call(e.h, action.Contract, action.ActionName, action.Data)

	if err != nil {
		if strings.Contains(err.Error(), "execution killed") {
			status = &tx.Status{
				Code:    tx.ErrorTimeout,
				Message: err.Error(),
			}
		} else {
			status = &tx.Status{
				Code:    tx.ErrorRuntime,
				Message: err.Error(),
			}
		}

		receipt := &tx.Receipt{
			FuncName: action.Contract + "/" + action.ActionName,
			Content:  err.Error(),
		}
		receipts = append(receipts, receipt)

		err = nil

		return
	}

	rj, errj := json.Marshal(rtn)
	if errj != nil {
		panic(errj)
	}

	ret = string(rj)

	receipts = append(receipts, e.h.Context().GValue("receipts").([]*tx.Receipt)...)

	status = &tx.Status{
		Code:    tx.Success,
		Message: "",
	}
	return
}

// Run actions in tx
func (e *Isolator) Run() (*tx.TxReceipt, error) { // nolinty
	e.h.Context().GSet("gas_limit", e.t.GasLimit)
	e.h.Context().GSet("receipts", make([]*tx.Receipt, 0))

	e.tr = tx.NewTxReceipt(e.t.Hash())

	if e.t.Delay > 0 {
		e.h.DB().StoreDelaytx(string(e.t.Hash()))
		e.tr.Status = &tx.Status{
			Code:    tx.Success,
			Message: "",
		}
		e.tr.GasUsage = e.t.Delay / 1e9 // TODO: determine the price
		return e.tr, nil
	}

	if e.t.IsDefer() {
		if !e.h.DB().HasDelaytx(string(e.t.ReferredTx)) {
			return nil, fmt.Errorf("delay tx not found, hash=%v", e.t.ReferredTx)
		}
		e.h.DB().DelDelaytx(string(e.t.ReferredTx))

		if !e.t.IsExpired(e.blockBaseCtx.Value("time").(int64)) {
			e.tr.Status = &tx.Status{
				Code:    tx.Success,
				Message: "transaction expired",
			}
			e.tr.GasUsage = 1 // TODO: determine the price
			return e.tr, nil
		}
	}

	hasSetCode := false

	for _, action := range e.t.Actions {
		if hasSetCode && action.Contract == "iost.system" && action.ActionName == "SetCode" {
			e.tr.Receipts = nil
			e.tr.Status.Code = tx.ErrorDuplicateSetCode
			e.tr.Status.Message = "error duplicate set code in a tx"
			break
		}
		hasSetCode = action.Contract == "iost.system" && action.ActionName == "SetCode"

		cost, status, ret, receipts, err := e.runAction(*action)
		ilog.Debugf("run action : %v, result is %v", action, status.Code)
		ilog.Debug("used cost > ", cost)
		ilog.Debugf("status > \n%v\n", status)
		ilog.Debug("return value: ", ret)

		if err != nil {
			return nil, err
		}

		gasLimit := e.h.Context().GValue("gas_limit").(int64)

		e.tr.Status = status
		if (status.Code == tx.ErrorRuntime && status.Message == "out of gas") || (status.Code == tx.ErrorTimeout) {
			cost = contract.NewCost(0, 0, gasLimit)
		}

		e.tr.GasUsage += cost.ToGas()
		if status.Code == 0 {
			for k, v := range e.h.Costs() {
				e.tr.RAMUsage[k] = v.Data
			}
		}

		e.h.Context().GSet("gas_limit", gasLimit-cost.ToGas())

		e.h.PayCost(cost, e.publisherID)

		if status.Code != tx.Success {
			ilog.Errorf("isolator run action %v failed, status %v, will rollback", action, status)
			e.tr.Receipts = nil
			e.h.DB().Rollback()
			break
		} else {
			e.tr.Receipts = append(e.tr.Receipts, receipts...)
		}
		e.tr.Returns = append(e.tr.Returns, ret)
	}

	return e.tr, nil
}

// PayCost as name
func (e *Isolator) PayCost() (*tx.TxReceipt, error) {
	err := e.h.DoPay(e.h.Context().Value("witness").(string), e.t.GasPrice, true)
	if err != nil {
		e.h.DB().Rollback()
		e.tr.RAMUsage = make(map[string]int64)
		e.tr.Status.Code = tx.ErrorBalanceNotEnough
		e.tr.Status.Message = "balance not enough after executing actions: " + err.Error()

		err = e.h.DoPay(e.h.Context().Value("witness").(string), e.t.GasPrice, false)
		if err != nil {
			return nil, err
		}
	}
	return e.tr, nil
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
	e.blockBaseMode = false
	e.h.ClearCosts()
	e.h.DB().Rollback()
}
func checkTxParams(t *tx.Tx) error {
	if t.GasPrice < 100 || t.GasPrice > 10000 {
		return errGasPriceIllegal
	}
	if t.GasLimit < 500 {
		return errGasLimitIllegal
	}
	return nil
}

func loadBlkInfo(ctx *host.Context, bh *block.BlockHead) *host.Context {
	c := host.NewContext(ctx)
	c.Set("parent_hash", common.Base58Encode(bh.ParentHash))
	c.Set("number", bh.Number)
	c.Set("witness", bh.Witness)
	c.Set("time", bh.Time)
	if bh.Time <= 1 {
		panic(fmt.Sprintf("invalid blockhead time %v", bh.Time))
	}
	ilog.Debugf("loadBlkInfo set time to %v", bh.Time)
	return c
}

func loadTxInfo(h *host.Host, t *tx.Tx, publisherID string) {
	h.PushCtx()
	h.Context().Set("tx_time", t.Time)
	h.Context().Set("expiration", t.Expiration)
	h.Context().Set("gas_price", t.GasPrice)
	h.Context().Set("tx_hash", common.Base58Encode(t.Hash()))
	h.Context().Set("publisher", publisherID)
	h.Context().Set("amount_limit", t.AmountLimit)

	authList := make(map[string]int)
	for _, v := range t.Signs {
		authList[account.GetIDByPubkey(v.Pubkey)] = 1
	}
	for _, v := range t.PublishSigns {
		authList[account.GetIDByPubkey(v.Pubkey)] = 2
	}

	signers := make(map[string]int)
	for _, v := range t.Signers {
		signers[v] = 1
	}
	signers[t.Publisher] = 2

	h.Context().Set("auth_list", authList)
	h.Context().Set("signer_list", signers)
	h.Context().Set("auth_contract_list", make(map[string]int))
}
