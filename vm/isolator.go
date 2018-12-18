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
		if i.h.GasPayed(t.Publisher)*t.GasRatio >= t.GasLimit {
			return fmt.Errorf("gas limit should be larger")
		}
		gas := i.h.TotalGas(i.publisherID)
		err = CheckTxGasLimitValid(t, gas, i.h.DB())
		if err != nil {
			return err
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
func (i *Isolator) Run() (*tx.TxReceipt, error) { // nolint
	vmGasLimit := i.t.GasLimit/i.t.GasRatio - i.h.GasPayed()
	if vmGasLimit <= 0 {
		ilog.Fatalf("vmGasLimit < 0. It should not happen. %v / %v < %v", i.t.GasLimit, i.t.GasRatio, i.h.GasPayed())
	}
	i.h.Context().GSet("gas_limit", vmGasLimit)
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

	for _, action := range i.t.Actions {
		actionCost, status, ret, receipts, err := i.runAction(*action)
		ilog.Debugf("run action : %v, result is %v\n", action, status.Code)
		ilog.Debugf("used cost %v\n", actionCost)
		ilog.Debugf("status %v\n", status)
		ilog.Debugf("return value: %v\n", ret)
		if err != nil {
			return nil, err
		}

		i.tr.Status = status
		actionCost.AddAssign(contract.NewCost(0, int64(len(ret)), 0))
		if (status.Code == tx.ErrorRuntime && status.Message == "out of gas") ||
			(vmGasLimit < actionCost.ToGas()) ||
			(!i.genesisMode && !i.blockBaseMode && i.h.TotalGas(i.t.Publisher).Value/i.t.GasRatio < i.h.GasPayed()+vmGasLimit) {
			ilog.Errorf("out of gas vmGasLimit %v actionCost %v totalGas %v gasPayed %v", vmGasLimit, actionCost.ToGas(), i.h.TotalGas(i.t.Publisher).ToString(), i.h.GasPayed())
			status.Code = tx.ErrorRuntime
			status.Message = "out of gas"
			actionCost.CPU = vmGasLimit
			actionCost.Net = 0
			ret = ""
		} else if status.Code == tx.ErrorTimeout {
			actionCost.CPU = vmGasLimit
			actionCost.Net = 0
			ret = ""
		}

		i.h.PayCost(actionCost, i.publisherID)

		if status.Code != tx.Success {
			ilog.Warnf("isolator run action %v failed, status %v, will rollback", action, status)
			i.tr.Receipts = nil
			i.h.DB().Rollback()
			i.h.ClearRAMCosts()
			i.tr.RAMUsage = make(map[string]int64)
			break
		}

		i.tr.Receipts = append(i.tr.Receipts, receipts...)
		i.tr.Returns = append(i.tr.Returns, ret)
		vmGasLimit -= actionCost.ToGas()
		i.h.Context().GSet("gas_limit", vmGasLimit)
	}
	for k, v := range i.h.Costs() {
		i.tr.RAMUsage[k] = v.Data
	}

	return i.tr, nil
}

// PayCost as name
func (i *Isolator) PayCost() (*tx.TxReceipt, error) {
	if i.t.GasLimit < i.h.GasPayed()*i.t.GasRatio {
		ilog.Fatalf("total gas cost is above limit %v < %v * %v", i.t.GasLimit, i.h.GasPayed(), i.t.GasRatio)
	}
	payedGas, err := i.h.DoPay(i.h.Context().Value("witness").(string), i.t.GasRatio)
	if err != nil {
		ilog.Errorf("DoPay failed, rollback %v", err)
		i.h.DB().Rollback()

		i.h.ClearRAMCosts()
		i.tr.RAMUsage = make(map[string]int64)
		i.tr.Status.Code = tx.ErrorBalanceNotEnough
		i.tr.Status.Message = "balance not enough after executing actions: " + err.Error()
		payedGas, err = i.h.DoPay(i.h.Context().Value("witness").(string), i.t.GasRatio)
		if err != nil {
			return nil, err
		}
	}
	i.tr.GasUsage = payedGas.Value
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
	err := t.CheckGas()
	if err != nil {
		return err
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
	return c
}

func loadTxInfo(h *host.Host, t *tx.Tx, publisherID string) {
	h.PushCtx()
	h.Context().Set("tx_time", t.Time)
	h.Context().Set("expiration", t.Expiration)
	h.Context().Set("gas_ratio", t.GasRatio)
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
