package new_vm

import (
	"context"

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
	GasCheckTxFailed = uint64(100)
)

type Rollbacker interface {
	Rollback()
}

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

	ctx := context.Background()

	blkInfo := make(map[string]string)

	blkInfo["parent_hash"] = string(bh.ParentHash)
	blkInfo["number"] = strconv.FormatInt(bh.Number, 10)
	blkInfo["witness"] = string(bh.Witness)
	blkInfo["time"] = strconv.FormatInt(bh.Time, 10)

	bij, err := json.Marshal(blkInfo)
	if err != nil {
		panic(err)
	}

	ctx = context.WithValue(ctx, "block_info", database.SerializedJSON(bij))
	ctx = context.WithValue(ctx, "witness", blkInfo["witness"])

	db := database.NewVisitor(defaultCacheLength, cb)
	host := host.NewHost(ctx, db, staticMonitor)
	return &EngineImpl{host: host}
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

	ptxr := tx.NewTxReceipt(tx0.Hash())

	e.host.Ctx = context.WithValue(e.host.Ctx, "tx_info", database.SerializedJSON(txInfo))
	e.host.Ctx = context.WithValue(e.host.Ctx, "auth_list", authList)
	e.host.Ctx = context.WithValue(e.host.Ctx, "gas_price", int64(tx0.GasPrice))
	e.host.Ctx = context.WithValue(e.host.Ctx, "tx_receipt", &ptxr)

	for _, action := range tx0.Actions {

		e.host.Ctx = context.WithValue(e.host.Ctx, "stack0", "direct_call")
		e.host.Ctx = context.WithValue(e.host.Ctx, "stack_height", 1) // record stack trace

		e.host.Ctx = context.WithValue(e.host.Ctx, "gas_limit", tx0.GasLimit)

		c := e.host.DB.Contract(action.Contract)

		if c.Info == nil {

			ptxr.GasUsage += GasCheckTxFailed
			ptxr.Status = tx.Status{
				Code:    tx.ErrorParamter,
				Message: ErrContractNotFound.Error() + action.Contract,
			}

			return &ptxr, nil
		}

		abi := c.ABI(action.ActionName)

		if abi == nil {
			ptxr.GasUsage += GasCheckTxFailed
			ptxr.Status = tx.Status{
				Code:    tx.ErrorParamter,
				Message: ErrABINotFound.Error() + action.Contract,
			}
			return &ptxr, nil
		}

		args, err := unmarshalArgs(abi, action.Data)
		if err != nil {
			panic(err)
		}

		_, cost, err := staticMonitor.Call(e.host, action.Contract, action.ActionName, args...)
		if err != nil {
			ptxr.Status = tx.Status{
				Code:    tx.ErrorRuntime,
				Message: err.Error(),
			}
			return &ptxr, nil
		}

		e.host.Ctx = context.WithValue(e.host.Ctx, "gas_limit", tx0.GasLimit-uint64(cost.ToGas()))

		e.host.PayCost(cost, account.GetIdByPubkey(tx0.Publisher.Pubkey))
	}

	txr := e.host.Ctx.Value("tx_receipt").(*tx.TxReceipt)

	e.host.DoPay(e.host.Ctx.Value("witness").(string), int64(tx0.GasPrice))

	return txr, nil
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
		GasUsage: GasCheckTxFailed,
		Status: tx.Status{
			Code:    code,
			Message: message,
		},
		SuccActionNum: 0,
		Receipts:      make([]tx.Receipt, 0),
	}
}
