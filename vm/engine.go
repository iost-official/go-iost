package vm

import (
	"sync"

	"errors"

	"runtime"

	"github.com/bitly/go-simplejson"
	"github.com/iost-official/Go-IOS-Protocol/account"
	"github.com/iost-official/Go-IOS-Protocol/common"
	"github.com/iost-official/Go-IOS-Protocol/core/contract"
	"github.com/iost-official/Go-IOS-Protocol/core/new_block"
	"github.com/iost-official/Go-IOS-Protocol/core/new_tx"
	"github.com/iost-official/Go-IOS-Protocol/ilog"
	"github.com/iost-official/Go-IOS-Protocol/vm/database"
	"github.com/iost-official/Go-IOS-Protocol/vm/host"
	"github.com/iost-official/Go-IOS-Protocol/vm/native"
)

const (
	defaultCacheLength = 1000
)

const (
	gasCheckTxFailed = int64(100)
)

var (
	errContractNotFound = errors.New("contract not found")
	errSetUpArgs        = errors.New("key does not exist")
)

// Engine the smart contract engine
type Engine interface {
	SetUp(k, v string) error
	Exec(tx0 *tx.Tx) (*tx.TxReceipt, error)
	GC()
}

var staticMonitor *Monitor

var once sync.Once

type engineImpl struct {
	ho *host.Host

	jsPath string

	logger        *ilog.Logger
	consoleWriter *ilog.ConsoleWriter
	fileWriter    *ilog.FileWriter
}

// NewEngine ...
func NewEngine(bh *block.BlockHead, cb database.IMultiValue) Engine {
	db := database.NewVisitor(defaultCacheLength, cb)

	return newEngine(bh, db)
}

func newEngine(bh *block.BlockHead, db *database.Visitor) Engine {
	if staticMonitor == nil {
		once.Do(func() {
			staticMonitor = NewMonitor()
		})
	}

	ctx := host.NewContext(nil)

	ctx = loadBlkInfo(ctx, bh)

	if bh.Number == 0 && db.Contract("iost.system") == nil {
		db.SetContract(native.ABI())
	}

	logger := ilog.New()
	logger.Stop()
	h := host.NewHost(ctx, db, staticMonitor, logger)

	e := &engineImpl{ho: h, logger: logger}
	runtime.SetFinalizer(e, func(e *engineImpl) {
		e.GC()
	})

	return e
}

/*
SetUp keys:
	js_path   	path to libjs/
	log_level 	log level of this engine(debug, info, warning, error, fatal), unset to silent log
	log_path	path to log file, unset to disable saving logs
	log_enable	enable log, log_level should set
*/
func (e *engineImpl) SetUp(k, v string) error {
	switch k {
	case "js_path":
		e.jsPath = v
	case "log_level":
		e.setLogger(v, "", false)
	case "log_path":
		e.setLogger("", v, false)
	case "log_enable":
		e.setLogger("", "", true)
	default:
		return errSetUpArgs
	}
	return nil
}
func (e *engineImpl) Exec(tx0 *tx.Tx) (*tx.TxReceipt, error) {
	err := checkTx(tx0)
	if err != nil {
		return errReceipt(tx0.Hash(), tx.ErrorTxFormat, err.Error()), nil
	}

	bl := e.ho.DB().Balance(account.GetIDByPubkey(tx0.Publisher.Pubkey))
	if bl <= 0 || bl < tx0.GasPrice*tx0.GasLimit {
		return errReceipt(tx0.Hash(), tx.ErrorBalanceNotEnough, "publisher's balance less than price * limit"), nil
	}

	loadTxInfo(e.ho, tx0)
	defer func() {
		e.ho.PopCtx()
	}()

	e.ho.Context().GSet("gas_limit", tx0.GasLimit)
	e.ho.Context().GSet("receipts", make([]tx.Receipt, 0))

	txr := tx.NewTxReceipt(tx0.Hash())

	for _, action := range tx0.Actions {

		cost, status, receipts, err2 := e.runAction(action)
		e.logger.Info("run action : %v, result is %v", action, status.Code)
		e.logger.Debug("used cost > %v", cost)
		e.logger.Debug("status > \n%v\n", status)
		e.logger.Debug("receipts > \n%v\n", receipts)

		if err2 != nil {
			return nil, err2
		}

		if cost == nil {
			panic("cost is nil")
		}

		txr.Status = status
		txr.GasUsage += cost.ToGas()
		//ilog.Debug("action status: %v", status)

		if status.Code != tx.Success {
			txr.Receipts = nil
			e.logger.Debug("rollback")
			e.ho.DB().Rollback()
		} else {
			txr.Receipts = append(txr.Receipts, receipts...)
			txr.SuccActionNum++
		}

		gasLimit := e.ho.Context().GValue("gas_limit").(int64)
		e.ho.Context().GSet("gas_limit", gasLimit-cost.ToGas())

		e.ho.PayCost(cost, account.GetIDByPubkey(tx0.Publisher.Pubkey))
	}

	err = e.ho.DoPay(e.ho.Context().Value("witness").(string), tx0.GasPrice)
	if err != nil {
		e.ho.DB().Rollback()
		err = e.ho.DoPay(e.ho.Context().Value("witness").(string), tx0.GasPrice)
		if err != nil {
			ilog.Debug(err.Error())
			return nil, err
		}
	} else {
		e.ho.DB().Commit()
	}

	return &txr, nil
}
func (e *engineImpl) GC() {
	e.logger.Stop()
}

func checkTx(tx0 *tx.Tx) error {
	if tx0.GasPrice < 0 || tx0.GasPrice > 10000 {
		return errGasPriceIllegal
	}
	return nil
}
func unmarshalArgs(abi *contract.ABI, data string) ([]interface{}, error) {
	js, err := simplejson.NewJson([]byte(data))
	if err != nil {
		return nil, err
	}

	rtn := make([]interface{}, 0)
	arr, err := js.Array()
	if err != nil {
		return nil, err
	}

	if len(arr) != len(abi.Args) {
		return nil, errors.New("args unmatched to abi")
	}
	for i := range arr {
		switch abi.Args[i] {
		case "string":
			s, err := js.GetIndex(i).String()
			if err != nil {
				return nil, err
			}
			rtn = append(rtn, s)
		case "bool":
			s, err := js.GetIndex(i).Bool()
			if err != nil {
				return nil, err
			}
			rtn = append(rtn, s)
		case "number":
			s, err := js.GetIndex(i).Int64()
			if err != nil {
				return nil, err
			}
			rtn = append(rtn, s)
		case "json":
			s, err := js.GetIndex(i).Encode()
			if err != nil {
				return nil, err
			}
			rtn = append(rtn, s)
		}
	}

	return rtn, nil
	//return nil, errors.New("unsupported yet")

}
func errReceipt(hash []byte, code tx.StatusCode, message string) *tx.TxReceipt {
	return &tx.TxReceipt{
		TxHash:   hash,
		GasUsage: gasCheckTxFailed,
		Status: tx.Status{
			Code:    code,
			Message: message,
		},
		SuccActionNum: 0,
		Receipts:      make([]tx.Receipt, 0),
	}
}
func (e *engineImpl) runAction(action tx.Action) (cost *contract.Cost, status tx.Status, receipts []tx.Receipt, err error) {
	receipts = make([]tx.Receipt, 0)

	e.ho.PushCtx()
	defer func() {
		e.ho.PopCtx()
	}()

	e.ho.Context().Set("stack0", "direct_call")
	e.ho.Context().Set("stack_height", 1) // record stack trace

	c := e.ho.DB().Contract(action.Contract)
	if c == nil || c.Info == nil {
		cost = contract.NewCost(0, 0, gasCheckTxFailed)
		status = tx.Status{
			Code:    tx.ErrorParamter,
			Message: errContractNotFound.Error() + action.Contract,
		}
		return
	}

	abi := c.ABI(action.ActionName)
	if abi == nil {
		cost = contract.NewCost(0, 0, gasCheckTxFailed)
		status = tx.Status{
			Code:    tx.ErrorParamter,
			Message: errABINotFound.Error() + action.Contract,
		}
		return
	}

	args, err := unmarshalArgs(abi, action.Data)
	if err != nil {
		cost = contract.NewCost(0, 0, gasCheckTxFailed)
		status = tx.Status{
			Code:    tx.ErrorParamter,
			Message: err.Error(),
		}
		return
	}
	//var rtn []interface{}
	//rtn, cost, err = staticMonitor.Call(e.ho, action.Contract, action.ActionName, args...)
	//ilog.Debug("action %v > %v", action.Contract+"."+action.ActionName, rtn)

	_, cost, err = staticMonitor.Call(e.ho, action.Contract, action.ActionName, args...)
	e.logger.Debug("cost is %v", cost)

	if cost == nil {
		panic("cost is nil")
	}

	if err != nil {
		status = tx.Status{
			Code:    tx.ErrorRuntime,
			Message: err.Error(),
		}
		receipt := tx.Receipt{
			Type:    tx.SystemDefined,
			Content: err.Error(),
		}
		receipts = append(receipts, receipt)

		err = nil

		return
	}

	receipts = append(receipts, e.ho.Context().GValue("receipts").([]tx.Receipt)...)

	status = tx.Status{
		Code:    tx.Success,
		Message: "",
	}
	return
}

func (e *engineImpl) setLogger(level, path string, start bool) {
	if path == "" && !start {
		//ilog.Debug("console log accepted")
		if e.consoleWriter == nil {
			e.consoleWriter = ilog.NewConsoleWriter()
		}

		switch level {
		case "debug":
			e.consoleWriter.SetLevel(ilog.LevelDebug)
		case "info":
			e.consoleWriter.SetLevel(ilog.LevelInfo)
		case "warning":
			e.consoleWriter.SetLevel(ilog.LevelWarn)
		case "error":
			e.consoleWriter.SetLevel(ilog.LevelError)
		case "fatal":
			e.consoleWriter.SetLevel(ilog.LevelFatal)
		}

		return
	}

	if level == "" && !start {
		e.fileWriter = ilog.NewFileWriter(path)
	}

	if start {
		var ok bool
		if e.consoleWriter != nil {
			err := e.logger.AddWriter(e.consoleWriter)
			if err != nil {
				panic(err)
			}
			ok = true
		}
		if e.fileWriter != nil {
			err := e.logger.AddWriter(e.fileWriter)
			if err != nil {
				panic(err)
			}
			ok = true
		}

		if ok {
			e.logger.SetCallDepth(0)
			e.logger.Start()
		}
	}
}

func loadBlkInfo(ctx *host.Context, bh *block.BlockHead) *host.Context {
	c := host.NewContext(ctx)
	c.Set("parent_hash", common.Base58Encode(bh.ParentHash))
	c.Set("number", bh.Number)
	c.Set("witness", bh.Witness)
	c.Set("time", bh.Time)
	return c
}

func loadTxInfo(h *host.Host, t *tx.Tx) {
	h.PushCtx()
	h.Context().Set("time", t.Time)
	h.Context().Set("expiration", t.Expiration)
	h.Context().Set("gas_price", t.GasPrice)
	h.Context().Set("tx_hash", common.Base58Encode(t.Hash()))

	authList := make(map[string]int)
	for _, v := range t.Signers {
		authList[string(v)] = 1
	}

	authList[account.GetIDByPubkey(t.Publisher.Pubkey)] = 2

	h.Context().Set("auth_list", authList)

}
