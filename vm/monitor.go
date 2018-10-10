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
	errArgsNotEnough   = errors.New("args not enough")
	errArgsType        = errors.New("args type not match")
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
		cid = h.URL(contractName)
	} else {
		cid = contractName
	}

	c = h.DB().Contract(cid)
	//ilog.Debugf("after get Contract")
	//h.DB().PrintCache()
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
	//ilog.Debugf("before prepareContract")
	c, abi, args, err := m.prepareContract(h, contractName, api, jarg)
	//ilog.Debugf("after prepareContract")
	if err != nil {
		return nil, host.ABINotFoundCost, fmt.Errorf("\nprepare contract: %v", err)
	}

	h.PushCtx()
	//ilog.Debug("after pushctx")

	h.Context().Set("contract_name", contractName)
	h.Context().Set("abi_name", api)
	//ilog.Debug("after set abi_name")

	vm, ok := m.vms[c.Info.Lang]
	if !ok {
		vm = Factory(c.Info.Lang)
		m.vms[c.Info.Lang] = vm
		err := m.vms[c.Info.Lang].Init()
		if err != nil {
			panic(err)
		}
	}
	//ilog.Debug("before loadandcall")
	rtn, cost, err = vm.LoadAndCall(h, c, api, args...)
	//ilog.Debug("after loadandcall")

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

func checkArgs(abi *contract.ABI, args []interface{}) error {
	if len(abi.Args) > len(args) {
		return errArgsNotEnough
	}

	for i, t := range abi.Args {
		var ok bool
		switch t {
		case "string":
			_, ok = args[i].(string)
		case "number":
			_, ok = args[i].(int64)
		case "bool":
			_, ok = args[i].(bool)
		case "json":
			_, ok = args[i].([]byte)
		}
		if !ok {
			return errArgsType
		}
	}
	return nil
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
