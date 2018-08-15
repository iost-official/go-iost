package host

import (
	"errors"

	"strconv"

	"github.com/iost-official/Go-IOS-Protocol/core/contract"
	"github.com/iost-official/Go-IOS-Protocol/core/new_tx"
	"github.com/iost-official/Go-IOS-Protocol/new_vm/database"
)

var (
	ErrBalanceNotEnough = errors.New("balance not enough")
	ErrTransferNegValue = errors.New("trasfer amount less than zero")
	ErrReenter          = errors.New("re-entering")
)

type Monitor interface {
	Call(host *Host, contractName, api string, args ...interface{}) (rtn []interface{}, cost *contract.Cost, err error)
	Compile(con *contract.Contract) (string, error)
}

type Host struct {
	DBHandler
	Info
	Teller
	APIDelegate

	Ctx     *Context
	DB      *database.Visitor
	monitor Monitor
}

func NewHost(ctx *Context, db *database.Visitor, monitor Monitor) *Host {
	return &Host{
		Ctx:     ctx,
		DB:      db,
		monitor: monitor,

		DBHandler:   NewDBHandler(db, ctx),
		Info:        NewInfo(ctx),
		Teller:      NewTeller(db, ctx),
		APIDelegate: NewAPI(ctx),
	}

}

func (h *Host) Context() *Context {
	return h.Ctx
}

func (h *Host) SetContext(ctx *Context) {
	h.Ctx = ctx

}

func (h *Host) Call(contract, api string, args ...interface{}) ([]interface{}, *contract.Cost, error) {

	// save stack
	record := contract + "-" + api

	height := h.Ctx.Value("stack_height").(int)

	for i := 0; i < height; i++ {
		key := "stack" + strconv.Itoa(i)
		if h.Ctx.Value(key).(string) == record {
			return nil, nil, ErrReenter
		}
	}

	key := "stack" + strconv.Itoa(height)

	h.Ctx = NewContext(h.Ctx)

	h.Ctx.Set("stack_height", height+1)
	h.Ctx.Set(key, record)
	rtn, cost, err := h.monitor.Call(h, contract, api, args...)

	h.Ctx = h.Ctx.Base()

	return rtn, cost, err
}

func (h *Host) CallWithReceipt(contract, api string, args ...interface{}) ([]interface{}, *contract.Cost, error) {
	rtn, cost, err := h.Call(contract, api, args...)

	var s string
	if err != nil {
		s = err.Error()
	} else {
		s = "success"
	}
	h.receipt(tx.SystemDefined, s)

	return rtn, cost, err

}

func (h *Host) SetCode(c *contract.Contract) (*contract.Cost, error) {
	code, err := h.monitor.Compile(c)
	if err != nil {
		return contract.NewCost(0, 0, 100), err
	}
	c.Code = code

	l := int64(len(c.Encode()) / 100)

	h.DB.SetContract(c)

	_, cost, err := h.monitor.Call(h, c.ID, "constructor")

	cost.AddAssign(contract.NewCost(0, l, 100))

	return cost, err // todo check set cost
}

func (h *Host) UpdateCode(c *contract.Contract, id database.SerializedJSON) (*contract.Cost, error) {

	rtn, cost, err := h.monitor.Call(h, c.ID, "can_update")

	if err != nil {
		return contract.NewCost(0, 0, 100), err
	}

	if t, ok := rtn[0].(bool); !ok || !t {
		return cost, errors.New("update refused")
	}

	c2, err := h.SetCode(c)

	if err != nil {
		cost.AddAssign(contract.NewCost(0, 0, 100))
		return cost, err
	}

	c2.AddAssign(cost)
	return c2, err
}

func (h *Host) DestroyCode(contractName string) (*contract.Cost, error) {
	// todo 释放kv

	rtn, cost, err := h.monitor.Call(h, contractName, "can_destroy")

	if err != nil {
		return contract.NewCost(0, 0, 100), err
	}

	if t, ok := rtn[0].(bool); !ok || !t {
		return cost, errors.New("destroy refused")
	}

	h.DB.DelContract(contractName)
	return contract.NewCost(1, 2, 3), nil
}
