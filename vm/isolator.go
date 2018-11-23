package vm

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"errors"

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

var staticMonitor = NewMonitor()

// TriggerBlockBaseMode start blockbase mode
func (i *Isolator) TriggerBlockBaseMode() {
	i.blockBaseMode = true
}

// Prepare Isolator
func (i *Isolator) Prepare(bh *block.BlockHead, db *database.Visitor, logger *ilog.Logger) error {
	if db.Contract("system.iost") == nil {
		db.SetContract(native.SystemABI())
	}
	if bh.Number == 0 {
		i.genesisMode = true
	} else {
		i.genesisMode = false
	}

	i.blockBaseCtx = host.NewContext(nil)
	i.blockBaseCtx = loadBlkInfo(i.blockBaseCtx, bh)
	i.h = host.NewHost(i.blockBaseCtx, db, staticMonitor, logger)
	i.h.ReadSettings()
	return nil
}

// PrepareTx read tx and ready to run
func (i *Isolator) PrepareTx(t *tx.Tx, limit time.Duration) error {
	i.t = t
	i.h.SetDeadline(time.Now().Add(limit))
	i.publisherID = t.Publisher
	l := len(t.Encode())
	i.h.PayCost(contract.NewCost(0, int64(l), 0), t.Publisher)

	if !i.genesisMode && !i.blockBaseMode {
		err := checkTxParams(t)
		if err != nil {
			return err
		}

		gas, _ := i.h.CurrentGas(i.publisherID)
		price := &common.Fixed{Value: t.GasPrice, Decimal: 2}
		ilog.Infof("publisher %v current gas %v\n", i.publisherID, gas.ToString())
		if gas.LessThan(price.Times(t.GasLimit)) {
			ilog.Infof("publisher's gas less than price * limit: publisher %v current gas %v price %v limit %v\n", i.publisherID, gas.ToString(), t.GasPrice, t.GasLimit)
			return fmt.Errorf("%v gas less than price * limit %v < %v * %v", i.publisherID, gas.ToString(), price.ToString(), t.GasLimit)
		}
	}
	loadTxInfo(i.h, t, i.publisherID)
	if !i.genesisMode && !i.blockBaseMode {
		err := i.checkAuth(t)
		if err != nil {
			return err
		}
	}
	return nil
}

func (i *Isolator) checkAuth(t *tx.Tx) error {
	for _, item := range t.Signers {
		ss := strings.Split(item, "@")
		if len(ss) != 2 {
			return fmt.Errorf("illegal signer: %v", item)
		}
		b, c := i.h.RequireAuth(ss[0], ss[1])
		if !b {
			return fmt.Errorf("unauthorized signer: %v", item)
		}
		i.h.PayCost(c, t.Publisher)
	}
	b, c := i.h.RequireAuth(t.Publisher, "active")
	if !b {
		return fmt.Errorf("unauthorized publisher: %v", t.Publisher)
	}
	i.h.PayCost(c, t.Publisher)
	// check amount limit
	for _, limit := range t.AmountLimit {
		decimal := i.h.DB().Decimal(limit.Token)
		if decimal == -1 {
			return errors.New("token in amountLimit not exists, " + limit.Token)
		}
		_, err := common.NewFixed(limit.Val, decimal)
		if err != nil {
			return err
		}
	}
	return nil
}

func (i *Isolator) runAction(action tx.Action) (cost contract.Cost, status *tx.Status, ret string, receipts []*tx.Receipt, err error) {
	receipts = make([]*tx.Receipt, 0)

	i.h.PushCtx()
	defer func() {
		i.h.PopCtx()
	}()

	i.h.Context().Set("stack0", "direct_call")
	i.h.Context().Set("stack_height", 1) // record stack trace

	var rtn []interface{}

	rtn, cost, err = staticMonitor.Call(i.h, action.Contract, action.ActionName, action.Data)

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

	receipts = append(receipts, i.h.Context().GValue("receipts").([]*tx.Receipt)...)

	status = &tx.Status{
		Code:    tx.Success,
		Message: "",
	}
	return
}

// Run actions in tx
func (i *Isolator) Run() (*tx.TxReceipt, error) { // nolinty
	i.h.Context().GSet("gas_limit", i.t.GasLimit)
	i.h.Context().GSet("receipts", make([]*tx.Receipt, 0))

	i.tr = tx.NewTxReceipt(i.t.Hash())

	if i.t.Delay > 0 {
		i.h.DB().StoreDelaytx(string(i.t.Hash()))
		i.tr.Status = &tx.Status{
			Code:    tx.Success,
			Message: "",
		}
		i.tr.GasUsage = i.t.Delay / 1e9 // TODO: determine the price
		return i.tr, nil
	}

	if i.t.IsDefer() {
		if !i.h.DB().HasDelaytx(string(i.t.ReferredTx)) {
			return nil, fmt.Errorf("delay tx not found, hash=%v", i.t.ReferredTx)
		}
		i.h.DB().DelDelaytx(string(i.t.ReferredTx))

		if !i.t.IsExpired(i.blockBaseCtx.Value("time").(int64)) {
			i.tr.Status = &tx.Status{
				Code:    tx.Success,
				Message: "transaction expired",
			}
			i.tr.GasUsage = 1 // TODO: determine the price
			return i.tr, nil
		}
	}

	hasSetCode := false

	for _, action := range i.t.Actions {
		if hasSetCode && action.Contract == "system.iost" && action.ActionName == "SetCode" {
			i.tr.Receipts = nil
			i.tr.Status.Code = tx.ErrorDuplicateSetCode
			i.tr.Status.Message = "error duplicate set code in a tx"
			break
		}
		hasSetCode = action.Contract == "system.iost" && action.ActionName == "SetCode"

		cost, status, ret, receipts, err := i.runAction(*action)
		//ilog.Debugf("run action : %v, result is %v", action, status.Code)
		//ilog.Debug("used cost > ", cost)
		//ilog.Debugf("status > \n%v\n", status)
		//ilog.Debug("return value: ", ret)

		if err != nil {
			return nil, err
		}

		i.tr.Status = status
		gasLimit := i.h.Context().GValue("gas_limit").(int64)
		if (status.Code == tx.ErrorRuntime && status.Message == "out of gas") || (status.Code == tx.ErrorTimeout) {
			cost.CPU = gasLimit
			cost.Net = 0
		}
		if cost.ToGas() > gasLimit {
			i.tr.Status = &tx.Status{Code: tx.ErrorRuntime, Message: host.ErrGasLimitExceeded.Error()}
			cost.CPU = gasLimit
			cost.Net = 0
		}

		i.tr.GasUsage += cost.ToGas()
		if status.Code == 0 {
			for k, v := range i.h.Costs() {
				i.tr.RAMUsage[k] = v.Data
			}
		}

		i.h.Context().GSet("gas_limit", gasLimit-cost.ToGas())

		i.h.PayCost(cost, i.publisherID)

		if status.Code != tx.Success {
			ilog.Errorf("isolator run action %v failed, status %v, will rollback", action, status)
			i.tr.Receipts = nil
			i.h.DB().Rollback()
			i.h.ClearRAMCosts()
			i.tr.RAMUsage = make(map[string]int64)
			break
		} else {
			i.tr.Receipts = append(i.tr.Receipts, receipts...)
		}
		i.tr.Returns = append(i.tr.Returns, ret)
	}

	return i.tr, nil
}

// PayCost as name
func (i *Isolator) PayCost() (*tx.TxReceipt, error) {
	err := i.h.DoPay(i.h.Context().Value("witness").(string), i.t.GasPrice)
	if err != nil {
		ilog.Errorf("DoPay failed, rollback %v", err)
		i.h.DB().Rollback()
		i.h.ClearRAMCosts()
		i.tr.RAMUsage = make(map[string]int64)
		i.tr.Status.Code = tx.ErrorBalanceNotEnough
		i.tr.Status.Message = "balance not enough after executing actions: " + err.Error()

		err = i.h.DoPay(i.h.Context().Value("witness").(string), i.t.GasPrice)
		if err != nil {
			panic(err)
		}
	}
	return i.tr, nil
}

// Commit flush changes to db
func (i *Isolator) Commit() {
	i.h.DB().Commit()
}

// ClearAll clear this isolator
func (i *Isolator) ClearAll() {
	i.h = nil
}

// ClearTx clear this tx
func (i *Isolator) ClearTx() {
	i.h.SetContext(i.blockBaseCtx)
	i.h.Context().GClear()
	i.blockBaseMode = false
	i.h.ClearCosts()
	i.h.DB().Rollback()
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
	//ilog.Debugf("loadBlkInfo set time to %v", bh.Time)
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
		x := strings.Split(v, "@")
		if len(x) != 2 {
			ilog.Error("signer format error. " + v)
			continue
		}
		signers[x[0]] = 1
	}
	signers[t.Publisher] = 2

	h.Context().Set("auth_list", authList)
	h.Context().Set("signer_list", signers)
	h.Context().Set("auth_contract_list", make(map[string]int))
}
