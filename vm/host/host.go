package host

import (
	"errors"

	"strconv"

	"encoding/json"

	"github.com/iost-official/Go-IOS-Protocol/core/contract"
	"github.com/iost-official/Go-IOS-Protocol/core/new_tx"
	"github.com/iost-official/Go-IOS-Protocol/ilog"
	"github.com/iost-official/Go-IOS-Protocol/vm/database"
)

// var ...
var (
	ErrBalanceNotEnough = errors.New("balance not enough")
	ErrTransferNegValue = errors.New("trasfer amount less than zero")
	ErrReenter          = errors.New("re-entering")
	ErrPermissionLost   = errors.New("transaction has no permission")
)

// Monitor ...
type Monitor interface {
	Call(host *Host, contractName, api string, args ...interface{}) (rtn []interface{}, cost *contract.Cost, err error)
	Compile(con *contract.Contract) (string, error)
}

// Host ...
type Host struct {
	DBHandler
	Info
	Teller
	APIDelegate
	EventPoster

	logger  *ilog.Logger
	ctx     *Context
	db      *database.Visitor
	monitor Monitor
}

// NewHost ...
func NewHost(ctx *Context, db *database.Visitor, monitor Monitor, logger *ilog.Logger) *Host {
	return &Host{
		ctx:     ctx,
		db:      db,
		monitor: monitor,
		logger:  logger,

		DBHandler:   NewDBHandler(db, ctx),
		Info:        NewInfo(ctx),
		Teller:      NewTeller(db, ctx),
		APIDelegate: NewAPI(ctx),
	}

}

// Context ...
func (h *Host) Context() *Context {
	return h.ctx
}

// SetContext ...
func (h *Host) SetContext(ctx *Context) {
	h.ctx = ctx

}

// Call  ...
func (h *Host) Call(contract, api string, args ...interface{}) ([]interface{}, *contract.Cost, error) {

	// save stack
	record := contract + "-" + api

	height := h.ctx.Value("stack_height").(int)

	for i := 0; i < height; i++ {
		key := "stack" + strconv.Itoa(i)
		if h.ctx.Value(key).(string) == record {
			return nil, nil, ErrReenter
		}
	}

	key := "stack" + strconv.Itoa(height)

	h.ctx = NewContext(h.ctx)

	h.ctx.Set("stack_height", height+1)
	h.ctx.Set(key, record)
	rtn, cost, err := h.monitor.Call(h, contract, api, args...)

	h.ctx = h.ctx.Base()

	return rtn, cost, err
}

// CallWithReceipt ...
func (h *Host) CallWithReceipt(contractName, api string, args ...interface{}) ([]interface{}, *contract.Cost, error) {
	rtn, cost, err := h.Call(contractName, api, args...)

	cost.AddAssign(contract.NewCost(0, 0, 100))

	var sarr []interface{}
	sarr = append(sarr, api)
	sarr = append(sarr, args)

	if err != nil {
		sarr = append(sarr, err.Error())
	} else {
		sarr = append(sarr, "success")
	}
	s, err := json.Marshal(sarr)
	if err != nil {
		return rtn, cost, err
	}
	h.receipt(tx.SystemDefined, string(s))

	return rtn, cost, err

}

// SetCode ...
func (h *Host) SetCode(c *contract.Contract) (*contract.Cost, error) {
	code, err := h.monitor.Compile(c)
	if err != nil {
		return contract.NewCost(0, 0, 100), err
	}
	c.Code = code

	l := int64(len(c.Encode()) / 100)
	//ilog.Debugf("length is : %v", l)

	h.db.SetContract(c)

	_, cost, err := h.monitor.Call(h, c.ID, "constructor")

	cost.AddAssign(contract.NewCost(0, l, 100))

	//ilog.Debugf("set gas is : %v", cost.ToGas())

	return cost, err // todo check set cost
}

// UpdateCode ...
func (h *Host) UpdateCode(c *contract.Contract, id database.SerializedJSON) (*contract.Cost, error) {
	oc := h.db.Contract(c.ID)
	if oc == nil {
		return contract.NewCost(0, 0, 100), errors.New("contract not exists")
	}
	abi := oc.ABI("can_update")
	if abi == nil {
		return contract.NewCost(0, 0, 100), errors.New("update refused")
	}

	rtn, cost, err := h.monitor.Call(h, c.ID, "can_update", []byte(id))

	if err != nil {
		return contract.NewCost(0, 0, 100), err
	}

	// todo return 返回类型应该是 bool
	if t, ok := rtn[0].(string); !ok || t != "true" {
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

// DestroyCode ...
func (h *Host) DestroyCode(contractName string) (*contract.Cost, error) {
	// todo 释放kv

	oc := h.db.Contract(contractName)
	if oc == nil {
		return contract.NewCost(0, 0, 100), errors.New("contract not exists")
	}
	abi := oc.ABI("can_destroy")
	if abi == nil {
		return contract.NewCost(0, 0, 100), errors.New("destroy refused")
	}

	rtn, cost, err := h.monitor.Call(h, contractName, "can_destroy")

	if err != nil {
		return contract.NewCost(0, 0, 100), err
	}

	// todo return 返回类型应该是 bool
	if t, ok := rtn[0].(string); !ok || t != "true" {
		return cost, errors.New("destroy refused")
	}

	h.db.DelContract(contractName)
	return contract.NewCost(1, 2, 3), nil
}

// Logger ...
func (h *Host) Logger() *ilog.Logger {
	return h.logger
}

// DB ...
func (h *Host) DB() *database.Visitor {
	return h.db
}

// PushCtx ...
func (h *Host) PushCtx() {

	ctx := NewContext(h.ctx)
	h.ctx = ctx

	h.DBHandler.ctx = ctx
	h.Info.ctx = ctx
	h.Teller.ctx = ctx
	h.APIDelegate.ctx = ctx
}

// PopCtx ...
func (h *Host) PopCtx() {
	ctx := h.ctx.Base()
	h.ctx = ctx
	h.DBHandler.ctx = ctx
	h.Info.ctx = ctx
	h.Teller.ctx = ctx
	h.APIDelegate.ctx = ctx
}
