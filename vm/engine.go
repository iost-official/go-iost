package vm

import (
	"errors"
	"fmt"
	"runtime"
	"strings"

	"time"

	"github.com/bitly/go-simplejson"
	"github.com/iost-official/go-iost/common"
	"github.com/iost-official/go-iost/core/block"
	"github.com/iost-official/go-iost/core/contract"
	"github.com/iost-official/go-iost/core/tx"
	"github.com/iost-official/go-iost/ilog"
	"github.com/iost-official/go-iost/vm/database"
	"github.com/iost-official/go-iost/vm/host"
	"github.com/iost-official/go-iost/vm/native"
)

const (
	defaultCacheLength = 1000
)

var (
	errSetUpArgs = errors.New("key does not exist")
	errCannotPay = errors.New("publisher's balance less than price * limit")
)

//go:generate mockgen -destination mock/engine_mock.go -package mock github.com/iost-official/go-iost/vm Engine

// Engine the smart contract engine
type Engine interface {
	SetUp(k, v string) error
	Exec(tx0 *tx.Tx, limit time.Duration) (*tx.TxReceipt, error)
	GC()
}

var staticMonitor = NewMonitor()
var jsPath = "./v8vm/v8/libjs/"
var logLevel = ""

// SetUp setup global engine settings
func SetUp(config *common.VMConfig) error {
	jsPath = config.JsPath
	logLevel = config.LogLevel
	return nil
}

type engineImpl struct {
	ho          *host.Host
	publisherID string

	logger        *ilog.Logger
	consoleWriter *ilog.ConsoleWriter
	fileWriter    *ilog.FileWriter
}

// NewEngine ...
func NewEngine(bh *block.BlockHead, cb database.IMultiValue) Engine {
	db := database.NewVisitor(defaultCacheLength, cb)

	e := newEngine(bh, db)

	return e
}

func newEngine(bh *block.BlockHead, db *database.Visitor) Engine {
	ctx := host.NewContext(nil)

	ctx = loadBlkInfo(ctx, bh)

	//ilog.Error("iost.system is ", db.Contract("iost.system"))

	if db.Contract("iost.system") == nil {
		db.SetContract(native.SystemABI())
	}

	logger := ilog.New()
	logger.Stop()
	h := host.NewHost(ctx, db, staticMonitor, logger)

	e := &engineImpl{ho: h, logger: logger}
	runtime.SetFinalizer(e, func(e *engineImpl) {
		e.GC()
	})

	if logLevel != "" {
		e.setLogLevel(logLevel)
		e.startLog()
	}

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
		jsPath = v
	case "log_level":
		e.setLogLevel(v)
	case "log_path":
		e.setLogPath(v)
	case "log_enable":
		e.startLog()
	default:
		return errSetUpArgs
	}
	return nil
}

// nolint
func (e *engineImpl) exec(tx0 *tx.Tx, limit time.Duration) (*tx.TxReceipt, error) {
	e.ho.SetDeadline(time.Now().Add(limit))
	err := checkTxParams(tx0)
	if err != nil {
		ilog.Error(err)
		return errReceipt(tx0.Hash(), tx.ErrorTxFormat, err.Error()), err
	}

	e.publisherID = tx0.Publisher
	bl := e.ho.DB().Balance(e.publisherID)

	if bl < 0 || bl < tx0.GasPrice*tx0.GasLimit {
		ilog.Error(errCannotPay)
		return errReceipt(tx0.Hash(), tx.ErrorBalanceNotEnough, "publisher's balance less than price * limit"), errCannotPay
	}

	loadTxInfo(e.ho, tx0, e.publisherID)
	defer func() {
		e.ho.PopCtx()
	}()

	e.ho.Context().GSet("gas_limit", tx0.GasLimit)
	e.ho.Context().GSet("receipts", make([]*tx.Receipt, 0))

	txr := tx.NewTxReceipt(tx0.Hash())
	hasSetCode := false

	for _, action := range tx0.Actions {
		if hasSetCode && action.Contract == "iost.system" && action.ActionName == "SetCode" {
			txr.Receipts = nil
			txr.Status.Code = tx.ErrorDuplicateSetCode
			txr.Status.Message = "error duplicate set code in a tx"
			ilog.Debugf("rollback")
			e.ho.DB().Rollback()
			break
		}
		hasSetCode = action.Contract == "iost.system" && action.ActionName == "SetCode"

		cost, status, receipts, err := e.runAction(*action)
		ilog.Debugf("run action : %v, result is %v", action, status.Code)
		ilog.Debug("used cost > ", cost)
		ilog.Debugf("status > \n%v\n", status)

		if err != nil {
			ilog.Error(err)
			return nil, err
		}

		gasLimit := e.ho.Context().GValue("gas_limit").(int64)

		txr.Status = status
		if (status.Code == 4 && status.Message == "out of gas") || (status.Code == 5) {
			cost = contract.NewCost(0, 0, gasLimit)
		}

		txr.GasUsage += cost.ToGas()

		e.ho.Context().GSet("gas_limit", gasLimit-cost.ToGas())

		e.ho.PayCost(cost, e.publisherID)

		if status.Code != tx.Success {
			txr.Receipts = nil
			ilog.Debugf("rollback")
			e.ho.DB().Rollback()
			break
		} else {
			txr.Receipts = append(txr.Receipts, receipts...)
			txr.SuccActionNum++
		}
	}

	err = e.ho.DoPay(e.ho.Context().Value("witness").(string), tx0.GasPrice)
	if err != nil {
		e.ho.DB().Rollback()
		err = e.ho.DoPay(e.ho.Context().Value("witness").(string), tx0.GasPrice)
		if err != nil {
			ilog.Error(err.Error())
			return nil, err
		}
	}

	return txr, nil
}

func (e *engineImpl) Exec(tx0 *tx.Tx, limit time.Duration) (*tx.TxReceipt, error) {
	r, err := e.exec(tx0, limit)
	e.ho.DB().Commit()
	return r, err
}
func (e *engineImpl) GC() {
	e.logger.Stop()
}

// nolint
func unmarshalArgs(abi *contract.ABI, data string) ([]interface{}, error) {
	if strings.HasSuffix(data, ",]") {
		data = data[:len(data)-2] + "]"
	}
	js, err := simplejson.NewJson([]byte(data))
	if err != nil {
		return nil, fmt.Errorf("error in abi file: %v", err)
	}

	rtn := make([]interface{}, 0)
	arr, err := js.Array()
	if err != nil {
		ilog.Error(js.EncodePretty())
		return nil, err
	}

	if len(arr) != len(abi.Args) {
		return nil, errors.New("args length unmatched to abi " + abi.Name)
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
			// make sure s is a valid json
			_, err = simplejson.NewJson(s)
			if err != nil {
				ilog.Error(string(s))
				return nil, err
			}
			rtn = append(rtn, s)
		}
	}

	return rtn, nil
}
func errReceipt(hash []byte, code tx.StatusCode, message string) *tx.TxReceipt {
	return &tx.TxReceipt{
		TxHash:   hash,
		GasUsage: 0,
		Status: &tx.Status{
			Code:    code,
			Message: message,
		},
		SuccActionNum: 0,
		Receipts:      make([]*tx.Receipt, 0),
	}
}
func (e *engineImpl) runAction(action tx.Action) (cost *contract.Cost, status *tx.Status, receipts []*tx.Receipt, err error) {
	receipts = make([]*tx.Receipt, 0)

	e.ho.PushCtx()
	defer func() {
		e.ho.PopCtx()
	}()

	e.ho.Context().Set("stack0", "direct_call")
	e.ho.Context().Set("stack_height", 1) // record stack trace

	_, cost, err = staticMonitor.Call(e.ho, action.Contract, action.ActionName, action.Data)

	if cost == nil {
		panic("cost is nil")
	}

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
			Type:    tx.SystemDefined,
			Content: err.Error(),
		}
		receipts = append(receipts, receipt)

		err = nil

		return
	}

	receipts = append(receipts, e.ho.Context().GValue("receipts").([]*tx.Receipt)...)

	status = &tx.Status{
		Code:    tx.Success,
		Message: "",
	}
	return
}

func (e *engineImpl) setLogLevel(level string) {
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
}
func (e *engineImpl) setLogPath(path string) {
	e.fileWriter = ilog.NewFileWriter(path)
}
func (e *engineImpl) startLog() {
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
		e.logger.HideLocation()
		e.logger.AsyncWrite()
		e.logger.Start()
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

func loadTxInfo(h *host.Host, t *tx.Tx, publisherID string) {
	h.PushCtx()
	h.Context().Set("time", t.Time)
	h.Context().Set("expiration", t.Expiration)
	h.Context().Set("gas_price", t.GasPrice)
	h.Context().Set("tx_hash", common.Base58Encode(t.Hash()))
	h.Context().Set("publisher", publisherID)

	authList := make(map[string]int)
	for _, v := range t.Signers {
		authList[v] = 1
	}

	authList[publisherID] = 2

	h.Context().Set("auth_list", authList)
	h.Context().Set("auth_contract_list", make(map[string]int))
}
