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
	"github.com/pkg/errors"
)

const (
	defaultCacheLength = 1000
)

var (
	ErrContractNotFound = errors.New("contract not found")
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
	host *Host
}

func NewEngine(bh *block.BlockHead, cb database.IMultiValue) Engine {
	if staticMonitor == nil {
		once.Do(func() {
			staticMonitor = NewMonitor()
		})
	}

	ctx := context.Background() // todo

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
	host := NewHost(ctx, db)
	return &EngineImpl{host: host}
}
func (e *EngineImpl) Exec(tx0 *tx.Tx) (*tx.TxReceipt, error) {
	err := checkTx(tx0)
	if err != nil {
		return nil, err // todo receipt
	}

	totalCost := contract.Cost0()

	txInfo, err := json.Marshal(tx0)
	if err != nil {
		panic(err)
	}

	authList := make(map[string]int)
	for _, v := range tx0.Signers {
		authList[string(v)] = 1
	}

	authList[account.GetIdByPubkey(tx0.Publisher.Pubkey)] = 2

	ptxr := tx.NewTxReceipt(tx0.Hash())

	e.host.ctx = context.WithValue(e.host.ctx, "tx_info", database.SerializedJSON(txInfo))
	e.host.ctx = context.WithValue(e.host.ctx, "auth_list", authList)
	e.host.ctx = context.WithValue(e.host.ctx, "gas_limit", tx0.GasLimit)
	e.host.ctx = context.WithValue(e.host.ctx, "gas_price", int64(tx0.GasPrice))
	e.host.ctx = context.WithValue(e.host.ctx, "tx_receipt", &ptxr)

	for _, action := range tx0.Actions {

		e.host.ctx = context.WithValue(e.host.ctx, "stack0", "direct_call")
		e.host.ctx = context.WithValue(e.host.ctx, "stack_height", 1) // record stack trace

		c := e.host.db.Contract(action.Contract)

		if c.Info == nil {
			return nil, ErrContractNotFound
		}

		abi := c.ABI(action.ActionName)

		if abi == nil {
			return nil, ErrABINotFound
		}

		args, err := unmarshalArgs(abi, action.Data)
		if err != nil {
			panic(err)
		}
		if err = checkArgs(args); err != nil {
			panic(err)
		}
		// todo host call check args
		_, cost, err := staticMonitor.Call(e.host, action.Contract, action.ActionName, args...)
		if err != nil {
			txr := e.host.ctx.Value("tx_receipt").(*tx.TxReceipt)
			return txr, err
		}

		totalCost.AddAssign(cost)
	}

	txr := e.host.ctx.Value("tx_receipt").(*tx.TxReceipt)

	e.host.PayCost(totalCost, account.GetIdByPubkey(tx0.Publisher.Pubkey), int64(tx0.GasPrice))

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

	//if len(arr) < len(abi.Args) {
	//	panic("less args ")
	//}
	for i := range js.MustArray() {
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

func checkArgs(args []interface{}) error {
	return nil
}
