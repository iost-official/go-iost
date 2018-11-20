package host

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"encoding/json"

	"github.com/iost-official/go-iost/common"
	"github.com/iost-official/go-iost/core/contract"
	"github.com/iost-official/go-iost/ilog"
	"github.com/iost-official/go-iost/vm/database"
)

// Monitor monitor interface
type Monitor interface {
	Call(host *Host, contractName, api string, jarg string) (rtn []interface{}, cost contract.Cost, err error)
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
func (h *Host) Call(cont, api, jarg string, withAuth ...bool) ([]interface{}, contract.Cost, error) {

	// save stack
	record := cont + "-" + api

	height := h.ctx.Value("stack_height").(int)

	for i := 0; i < height; i++ {
		key := "stack" + strconv.Itoa(i)
		if h.ctx.Value(key).(string) == record {
			return nil, contract.Cost{}, ErrReenter
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
	rtn, cost, err := h.monitor.Call(h, cont, api, jarg)
	cost.AddAssign(CommonOpCost(height))

	h.ctx = h.ctx.Base()

	return rtn, cost, err
}

// CallWithAuth  call a new contract with permission of current contract
func (h *Host) CallWithAuth(contract, api, jarg string) ([]interface{}, contract.Cost, error) {
	return h.Call(contract, api, jarg, true)
}

func (h *Host) checkAmountLimitValid(c *contract.Contract) (contract.Cost, error) {
	cost := contract.Cost0()
	for _, abi := range c.Info.Abi {
		for _, limit := range abi.AmountLimit {
			cost.AddAssign(CommonOpCost(1))
			decimal := h.db.Decimal(limit.Token)
			if decimal == -1 {
				return cost, ErrAmountLimitTokenNotExists
			}
			_, err := common.NewFixed(limit.Val, decimal)
			if err != nil {
				return cost, err
			}
		}
	}
	return cost, nil
}

// SetCode set code to storage
func (h *Host) SetCode(c *contract.Contract, owner string) (contract.Cost, error) {
	code, err := h.monitor.Compile(c)
	if err != nil {
		return Costs["CompileCost"], err
	}
	c.Code = code

	cost, err := h.checkAmountLimitValid(c)
	if err != nil {
		return cost, err
	}

	initABI := contract.ABI{
		Name: "init",
		Args: []string{},
	}

	c.Info.Abi = append(c.Info.Abi, &initABI)

	l := len(c.Encode()) // todo multi Encode call
	h.PayCost(contract.NewCost(int64(l), 0, 0), owner)

	h.db.SetContract(c)

	_, cost0, err := h.Call(c.ID, "init", "[]")
	cost.AddAssign(cost0)

	return cost, err
}

// UpdateCode update code
func (h *Host) UpdateCode(c *contract.Contract, id database.SerializedJSON) (contract.Cost, error) {
	oc := h.db.Contract(c.ID)
	if oc == nil {
		return Costs["GetCost"], ErrContractNotFound
	}
	abi := oc.ABI("can_update")
	if abi == nil {
		return Costs["GetCost"], ErrUpdateRefused
	}

	oldL := len(oc.Encode())

	rtn, cost, err := h.Call(c.ID, "can_update", `["`+string(id)+`"]`)

	if err != nil {
		return cost, fmt.Errorf("call can_update: %v", err)
	}

	if t, ok := rtn[0].(string); !ok || t != "true" {
		return cost, ErrUpdateRefused
	}

	// set code  without invoking init
	code, err := h.monitor.Compile(c)
	cost.AddAssign(Costs["CompileCost"])
	if err != nil {
		return cost, err
	}
	c.Code = code

	h.db.SetContract(c)

	owner, co := h.GlobalMapGet("system.iost", "contract_owner", c.ID)
	cost.AddAssign(co)
	l := len(c.Encode()) // todo multi Encode call
	h.PayCost(contract.NewCost(int64(l-oldL), 0, 0), owner.(string))

	return cost, nil
}

// DestroyCode delete code
func (h *Host) DestroyCode(contractName string) (contract.Cost, error) {
	// todo free kv

	oc := h.db.Contract(contractName)
	if oc == nil {
		return Costs["GetCost"], ErrContractNotFound
	}
	abi := oc.ABI("can_destroy")
	if abi == nil {
		return Costs["GetCost"], ErrDestroyRefused
	}

	oldL := len(oc.Encode())

	rtn, cost, err := h.Call(contractName, "can_destroy", "[]")

	if err != nil {
		return cost, err
	}

	if t, ok := rtn[0].(string); !ok || t != "true" {
		return cost, ErrDestroyRefused
	}

	owner, co := h.GlobalMapGet("system.iost", "contract_owner", oc.ID)
	cost.AddAssign(co)
	h.PayCost(contract.NewCost(int64(-oldL), 0, 0), owner.(string))

	h.db.MDel("system.iost-contract_owner", oc.ID)

	h.db.DelContract(contractName)
	return Costs["PutCost"], nil
}

// CancelDelaytx deletes delaytx hash.
func (h *Host) CancelDelaytx(txHash string) (contract.Cost, error) {

	if !h.db.HasDelaytx(txHash) {
		return Costs["DelaytxNotFoundCost"], ErrDelaytxNotFound
	}

	h.db.DelDelaytx(txHash)
	return Costs["DelDelaytxCost"], nil
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

// IsValidAccount check whether account exists
func (h *Host) IsValidAccount(name string) bool {
	if h.Context().Value("number") == int64(0) {
		return true
	}
	return strings.HasPrefix(name, "Contract") || strings.HasSuffix(name, ".iost") || database.Unmarshal(h.DB().Get("auth.iost@"+name+"-auth")) != nil
}

// ReadSettings read settings from db
func (h *Host) ReadSettings() {
	j, _ := h.DBHandler.GlobalMapGet("system.iost", "settings", "host")
	if j == nil {
		return
	}
	var s Setting
	err := json.Unmarshal([]byte(j.(string)), &s)
	if err != nil {
		panic(err)
	}

	for k, v := range s.Costs {
		Costs[k] = v
	}

}
