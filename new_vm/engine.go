package new_vm

import (
	"context"

	"sync"

	"encoding/json"

	"strconv"

	"github.com/iost-official/Go-IOS-Protocol/account"
	"github.com/iost-official/Go-IOS-Protocol/core/contract"
	"github.com/iost-official/Go-IOS-Protocol/core/new_block"
	"github.com/iost-official/Go-IOS-Protocol/core/new_tx"
	"github.com/iost-official/Go-IOS-Protocol/new_vm/database"
)

const (
	defaultCacheLength = 1000
)

type Engine interface {
	//Init()
	//SetEnv(bh *block.BlockHead, cb database.IMultiValue) Engine
	Exec(tx0 tx.Tx) (*tx.TxReceipt, error)
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
	db := database.NewVisitor(defaultCacheLength, cb)
	host := NewHost(ctx, db)
	return &EngineImpl{host: host}
}
func (e *EngineImpl) Exec(tx0 tx.Tx) (*tx.TxReceipt, error) {

	//txr := tx.NewTxReceipt([]byte(tx0.Id))
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

	e.host.ctx = context.WithValue(e.host.ctx, "tx_info", database.SerializedJSON(txInfo))

	for _, action := range tx0.Actions {

		e.host.ctx = context.WithValue(e.host.ctx, "stack0", tx0.Id)
		e.host.ctx = context.WithValue(e.host.ctx, "stack_height", 1) // record stack trace

		_, cost, err := staticMonitor.Call(e.host, action.Contract, action.ActionName, action.Data)
		if err != nil {
			txr := e.host.ctx.Value("tx_receipt").(*tx.TxReceipt)
			return txr, err
		}

		totalCost.AddAssign(cost)
	}

	txr := e.host.ctx.Value("tx_receipt").(*tx.TxReceipt)

	e.host.PayCost(totalCost, account.GetIdByPubkey(tx0.Publisher.Pubkey), tx0.GasPrice)

	return txr, nil
}
func (e *EngineImpl) GC() {

}
