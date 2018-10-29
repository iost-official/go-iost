package vm

import (
	"errors"

	"fmt"

	"github.com/iost-official/go-iost/core/contract"
	"github.com/iost-official/go-iost/vm/host"
	"github.com/iost-official/go-iost/vm/native"
	"github.com/iost-official/go-iost/vm/v8vm"
)

var (
	errABINotFound     = errors.New("abi not found")
	errGasPriceIllegal = errors.New("gas price too big")
)

// Monitor ...
type Monitor struct {
	vms map[string]VM
}

// NewMonitor ...
func NewMonitor() *Monitor {
	m := &Monitor{
		vms: make(map[string]VM),
	}
	jsvm := Factory("javascript")
	m.vms["javascript"] = jsvm
	return m
}

func (m *Monitor) prepareContract(h *host.Host, contractName, api, jarg string) (c *contract.Contract, abi *contract.ABI, args []interface{}, err error) {
	var cid string
	if h.IsDomain(contractName) {
		cid = h.ContractID(contractName)
	} else {
		cid = contractName
	}

	c = h.DB().Contract(cid)
	if c == nil {
		return nil, nil, nil, errContractNotFound
	}

	abi = c.ABI(api)

	if abi == nil {
		return nil, nil, nil, errABINotFound
	}

	args, err = unmarshalArgs(abi, jarg)

	return
}

// Call ...
// nolint
func (m *Monitor) Call(h *host.Host, contractName, api string, jarg string) (rtn []interface{}, cost *contract.Cost, err error) {

	c, abi, args, err := m.prepareContract(h, contractName, api, jarg)

	if err != nil {
		return nil, host.ABINotFoundCost, fmt.Errorf("\nprepare contract: %v", err)
	}

	h.PushCtx()

	h.Context().Set("contract_name", c.ID)
	h.Context().Set("abi_name", api)

	vm, ok := m.vms[c.Info.Lang]
	if !ok {
		vm = Factory(c.Info.Lang)
		m.vms[c.Info.Lang] = vm
		err := m.vms[c.Info.Lang].Init()
		if err != nil {
			panic(err)
		}
	}
	rtn, cost, err = vm.LoadAndCall(h, c, api, args...)

	payment, ok := h.Context().GValue("abi_payment").(int)
	if !ok {
		payment = int(abi.Payment)
	}
	var gasPrice = h.Context().Value("gas_price").(int64)

	if payment == 1 &&
		abi.GasPrice > gasPrice &&
		!cost.IsOverflow(abi.Limit) {
		b := h.DB().Balance(host.ContractGasPrefix + contractName)
		if b > gasPrice*cost.ToGas() {
			h.PayCost(cost, host.ContractGasPrefix+contractName)
			cost = contract.Cost0()
		}
	}

	h.PopCtx()

	return
}

// Compile ...
func (m *Monitor) Compile(con *contract.Contract) (string, error) {
	switch con.Info.Lang {
	case "native":
		return "", nil
	case "javascript":
		jsvm, _ := m.vms["javascript"]
		return jsvm.Compile(con)
	}
	return "", errors.New("vm unsupported")
}

// Factory ...
func Factory(lang string) VM {
	switch lang {
	case "native":
		vm := native.Impl{}
		vm.Init()
		return &vm
	case "javascript":
		vm := v8.NewVMPool(10, 200)
		vm.Init()
		//vm.SetJSPath(jsPath)
		return vm
	}
	return nil
}
