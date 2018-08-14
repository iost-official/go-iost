package new_vm

import (
	"sync"

	"encoding/json"

	"strconv"

	"github.com/bitly/go-simplejson"
	"github.com/iost-official/Go-IOS-Protocol/account"
	"github.com/iost-official/Go-IOS-Protocol/core/contract"
	"github.com/iost-official/Go-IOS-Protocol/core/new_block"
	"github.com/iost-official/Go-IOS-Protocol/core/new_tx"
	"github.com/iost-official/Go-IOS-Protocol/new_vm/database"
	"github.com/iost-official/Go-IOS-Protocol/new_vm/host"
	"github.com/pkg/errors"
)

const (
	defaultCacheLength = 1000
)

var (
	ErrContractNotFound = errors.New("contract not found")
)

const (
	GasCheckTxFailed = int64(100)
)

type Engine interface {
	//Init()
	//SetEnv(bh *block.BlockHead, cb database.IMultiValue) Engine
	Exec(tx0 *tx.Tx) (*tx.TxReceipt, error)
	GC()
}

var staticMonitor *Monitor

var once sync.Once

type EngineImpl struct {
	host *host.Host
}

func NewEngine(bh *block.BlockHead, cb database.IMultiValue) Engine {
	if staticMonitor == nil {
		once.Do(func() {
			staticMonitor = NewMonitor()
		})
	}

	ctx := host.NewContext(nil)

	blkInfo := make(map[string]string)

	blkInfo["parent_hash"] = string(bh.ParentHash)
	blkInfo["number"] = strconv.FormatInt(bh.Number, 10)
	blkInfo["witness"] = string(bh.Witness)
	blkInfo["time"] = strconv.FormatInt(bh.Time, 10)

	bij, err := json.Marshal(blkInfo)
	if err != nil {
		panic(err)
	}

	ctx.Set("block_info", database.SerializedJSON(bij))
	ctx.Set("witness", blkInfo["witness"])

	db := database.NewVisitor(defaultCacheLength, cb)
	h := host.NewHost(ctx, db, staticMonitor)
	return &EngineImpl{host: h}
}
func (e *EngineImpl) Exec(tx0 *tx.Tx) (*tx.TxReceipt, error) {
	err := checkTx(tx0)
	if err != nil {
		return errReceipt(tx0.Hash(), tx.ErrorTxFormat, err.Error()), nil
	}

	// todo 检查发布者余额是否能支撑gaslimit

	txInfo, err := json.Marshal(tx0)
	if err != nil {
		panic(err) // should not get error
	}

	authList := make(map[string]int)
	for _, v := range tx0.Signers {
		authList[string(v)] = 1
	}

	authList[account.GetIdByPubkey(tx0.Publisher.Pubkey)] = 2

	e.host.Ctx = host.NewContext(e.host.Ctx)
	defer func() {
		e.host.Ctx = e.host.Ctx.Base()
	}()

	e.host.Ctx.Set("tx_info", database.SerializedJSON(txInfo))
	e.host.Ctx.Set("auth_list", authList)
	e.host.Ctx.Set("gas_price", int64(tx0.GasPrice))

	e.host.Ctx.GSet("gas_limit", tx0.GasLimit)
	e.host.Ctx.GSet("receipts", make([]tx.Receipt, 0))

	txr := tx.NewTxReceipt(tx0.Hash())

	for _, action := range tx0.Actions {

		cost, status, receipts, err := e.runAction(action)
		if err != nil {
			return nil, err
		}

		if cost == nil {
			panic("cost is nil")
		}

		txr.Status = status
		txr.GasUsage += uint64(cost.ToGas())

		if status.Code != tx.Success {
			txr.Receipts = nil
			e.host.DB.Rollback()
			break
		}

		txr.Receipts = append(txr.Receipts, receipts...)
		txr.SuccActionNum++

		e.host.Ctx.GSet("gas_limit", tx0.GasLimit-uint64(cost.ToGas()))

		e.host.PayCost(cost, account.GetIdByPubkey(tx0.Publisher.Pubkey))
	}

	e.host.DoPay(e.host.Ctx.Value("witness").(string), int64(tx0.GasPrice))
	e.host.DB.Commit()

	return &txr, nil
}
func (e *EngineImpl) GC() {

}

func checkTx(tx0 *tx.Tx) error {
	if tx0.GasPrice <= 0 || tx0.GasPrice > 10000 {
		return ErrGasPriceTooBig
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

	if len(arr) < len(abi.Args) {
		panic("less args ")
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
			rtn = append(rtn, database.SerializedJSON(s))
		}
	}

	return rtn, nil
	//return nil, errors.New("unsupported yet")

}

func errReceipt(hash []byte, code tx.StatusCode, message string) *tx.TxReceipt {
	return &tx.TxReceipt{
		TxHash:   hash,
		GasUsage: uint64(GasCheckTxFailed),
		Status: tx.Status{
			Code:    code,
			Message: message,
		},
		SuccActionNum: 0,
		Receipts:      make([]tx.Receipt, 0),
	}
}

func (e *EngineImpl) runAction(action tx.Action) (cost *contract.Cost, status tx.Status, receipts []tx.Receipt, err error) {
	receipts = make([]tx.Receipt, 0)
	cost = contract.Cost0()

	e.host.Ctx = host.NewContext(e.host.Ctx)
	defer func() {
		e.host.Ctx = e.host.Ctx.Base()
	}()

	e.host.Ctx.Set("stack0", "direct_call")
	e.host.Ctx.Set("stack_height", 1) // record stack trace

	c := e.host.DB.Contract(action.Contract)

	if c.Info == nil {
		cost = contract.NewCost(0, 0, GasCheckTxFailed)
		status = tx.Status{
			Code:    tx.ErrorParamter,
			Message: ErrContractNotFound.Error() + action.Contract,
		}

		return
	}

	abi := c.ABI(action.ActionName)

	if abi == nil {
		cost = contract.NewCost(0, 0, GasCheckTxFailed)
		status = tx.Status{
			Code:    tx.ErrorParamter,
			Message: ErrABINotFound.Error() + action.Contract,
		}
		return
	}

	args, err := unmarshalArgs(abi, action.Data)
	if err != nil {
		panic(err)
	}

	_, cost, err = staticMonitor.Call(e.host, action.Contract, action.ActionName, args...)

	if cost == nil {
		panic("here")
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

	receipts = append(receipts, e.host.Ctx.GValue("receipts").([]tx.Receipt)...)

	return
}
