package host

import (
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/iost-official/go-iost/core/contract"
	"github.com/iost-official/go-iost/core/tx"
	"github.com/iost-official/go-iost/ilog"
	"github.com/iost-official/go-iost/vm/database"
)

// Monitor monitor interface
type Monitor interface {
	Call(host *Host, contractName, api string, jarg string) (rtn []interface{}, cost *contract.Cost, err error)
	Compile(con *contract.Contract) (string, error)
}

// Host host struct, used as isolate of vm
type Host struct {
	DBHandler
	Info
	Teller
	APIDelegate
	EventPoster
	DNS
	Authority
	GasManager

	logger  *ilog.Logger
	ctx     *Context
	db      *database.Visitor
	monitor Monitor

	deadline time.Time
}

// NewHost get a new host
func NewHost(ctx *Context, db *database.Visitor, monitor Monitor, logger *ilog.Logger) *Host {
	h := &Host{
		ctx:     ctx,
		db:      db,
		monitor: monitor,
		logger:  logger,
	}
	h.DBHandler = NewDBHandler(h)
	h.Info = NewInfo(h)
	h.Teller = NewTeller(h)
	h.APIDelegate = NewAPI(h)
	h.EventPoster = EventPoster{}
	h.DNS = NewDNS(h)
	h.Authority = Authority{h: h}
	h.GasManager = NewGasManager(h)

	return h

}

// Context get context in host
func (h *Host) Context() *Context {
	return h.ctx
}

// SetContext set a new context to host
func (h *Host) SetContext(ctx *Context) {
	h.ctx = ctx

}

// Call  call a new contract in this context
func (h *Host) Call(contract, api, jarg string, withAuth ...bool) ([]interface{}, *contract.Cost, error) {

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

	// handle withAuth
	if len(withAuth) > 0 && withAuth[0] {
		authList := h.ctx.Value("auth_contract_list").(map[string]int)
		authList[h.ctx.Value("contract_name").(string)] = 1
		h.ctx.Set("auth_contract_list", authList)
	}

	h.ctx.Set("stack_height", height+1)
	h.ctx.Set(key, record)
	rtn, cost, err := h.monitor.Call(h, contract, api, jarg)
	cost.AddAssign(CommonOpCost(height))

	h.ctx = h.ctx.Base()

	return rtn, cost, err
}

// CallWithAuth  call a new contract with permission of current contract
func (h *Host) CallWithAuth(contract, api, jarg string) ([]interface{}, *contract.Cost, error) {
	return h.Call(contract, api, jarg, true)
}

// CallWithReceipt call and generate receipt
func (h *Host) CallWithReceipt(contractName, api, jarg string) ([]interface{}, *contract.Cost, error) {
	rtn, cost, err := h.Call(contractName, api, jarg)

	var sarr []interface{}
	sarr = append(sarr, api)
	sarr = append(sarr, jarg)

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
	cost.AddAssign(ReceiptCost(len(s)))
	return rtn, cost, err

}

// SetCode set code to storage
func (h *Host) SetCode(c *contract.Contract, owner string) (*contract.Cost, error) {
	code, err := h.monitor.Compile(c)
	if err != nil {
		return CompileErrCost, err
	}
	c.Code = code

	initABI := contract.ABI{
		Name:    "init",
		Payment: 0,
		Args:    []string{},
	}

	c.Info.Abi = append(c.Info.Abi, &initABI)

	l := len(c.Encode()) // todo multi Encode call
	h.PayCost(contract.NewCost(int64(l), 0, 0), owner)

	h.db.SetContract(c)

	_, cost, err := h.Call(c.ID, "init", "[]")

	return cost, err
}

// UpdateCode update code
func (h *Host) UpdateCode(c *contract.Contract, id database.SerializedJSON) (*contract.Cost, error) {
	oc := h.db.Contract(c.ID)
	if oc == nil {
		return ContractNotFoundCost, ErrContractNotFound
	}
	abi := oc.ABI("can_update")
	if abi == nil {
		return ABINotFoundCost, ErrUpdateRefused
	}

	oldL := len(oc.Encode())

	rtn, cost, err := h.Call(c.ID, "can_update", `["`+string(id)+`"]`)

	if err != nil {
		return cost, fmt.Errorf("call can_update: %v", err)
	}

	// todo rtn[0] should be bool type
	if t, ok := rtn[0].(string); !ok || t != "true" {
		return cost, ErrUpdateRefused
	}

	// set code  without invoking init
	code, err := h.monitor.Compile(c)
	cost.AddAssign(CompileErrCost)
	if err != nil {
		return cost, err
	}
	c.Code = code

	h.db.SetContract(c)

	owner, co := h.GlobalMapGet("iost.system", "contract_owner", c.ID)
	cost.AddAssign(co)
	l := len(c.Encode()) // todo multi Encode call
	h.PayCost(contract.NewCost(int64(l-oldL), 0, 0), owner.(string))

	return cost, nil
}

// DestroyCode delete code
func (h *Host) DestroyCode(contractName string) (*contract.Cost, error) {
	// todo free kv

	oc := h.db.Contract(contractName)
	if oc == nil {
		return ContractNotFoundCost, ErrContractNotFound
	}
	abi := oc.ABI("can_destroy")
	if abi == nil {
		return ABINotFoundCost, ErrDestroyRefused
	}

	oldL := len(oc.Encode())

	rtn, cost, err := h.Call(contractName, "can_destroy", "[]")

	if err != nil {
		return cost, err
	}

	// todo rtn[0] should be bool type
	if t, ok := rtn[0].(string); !ok || t != "true" {
		return cost, ErrDestroyRefused
	}

	owner, co := h.GlobalMapGet("iost.system", "contract_owner", oc.ID)
	cost.AddAssign(co)
	h.PayCost(contract.NewCost(int64(-oldL), 0, 0), owner.(string))

	h.db.MDel("iost.system-contract_owner", oc.ID)

	h.db.DelContract(contractName)
	return DelContractCost, nil
}

// CancelDelaytx deletes delaytx hash.
func (h *Host) CancelDelaytx(txHash string) (*contract.Cost, error) {

	if !h.db.HasDelaytx(txHash) {
		return DelaytxNotFoundCost, ErrDelaytxNotFound
	}

	h.db.DelDelaytx(txHash)
	return DelDelaytxCost, nil
}

// Logger get a log in host
func (h *Host) Logger() *ilog.Logger {
	return h.logger
}

// DB get current version mvccdb
func (h *Host) DB() *database.Visitor {
	return h.db
}

// PushCtx make a new context based on current one
func (h *Host) PushCtx() {
	ctx := NewContext(h.ctx)
	h.ctx = ctx
}

// PopCtx pop current context
func (h *Host) PopCtx() {
	ctx := h.ctx.Base()
	h.ctx = ctx
}

// Deadline return this host's deadline
func (h *Host) Deadline() time.Time {
	return h.deadline
}

// SetDeadline set this host's deadline
func (h *Host) SetDeadline(t time.Time) {
	h.deadline = t
}
