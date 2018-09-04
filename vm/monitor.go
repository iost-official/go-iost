package vm

import (
	"strings"

	"errors"

	"github.com/iost-official/Go-IOS-Protocol/core/contract"

	"github.com/iost-official/Go-IOS-Protocol/ilog"
	"github.com/iost-official/Go-IOS-Protocol/vm/host"
	"github.com/iost-official/Go-IOS-Protocol/vm/native"
	"github.com/iost-official/Go-IOS-Protocol/vm/v8vm"
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
	return m
}

// Call ...
// nolint
func (m *Monitor) Call(h *host.Host, contractName, api string, args ...interface{}) (rtn []interface{}, cost *contract.Cost, err error) {

	c := h.DB().Contract(contractName)
	abi := c.ABI(api)
	if abi == nil {
		return nil, host.ContractNotFoundCost, errABINotFound
	}

	err = checkArgs(abi, args)

	if err != nil {
		return nil, host.ABINotFoundCost, err
	}

	h.PushCtx()

	h.Context().Set("contract_name", contractName)
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
	if cost == nil {
		if strings.HasPrefix(contractName, "Contract") {
			ilog.Fatalf("will return nil cost : %v.%v", contractName, api)
		} else {
			ilog.Debugf("will return nil cost : %v.%v", contractName, api)
		}
		cost = contract.NewCost(100, 100, 100)
	}

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

//func (m *Monitor) Update(contractName string, newContract *contract.Contract) error {
//	err := m.Destory(contractName)
//	if err != nil {
//		return err
//	}
//	m.ho.db.SetContract(newContract)
//	return nil
//}
//
//func (m *Monitor) Destroy(contractName string) error {
//	m.ho.db.DelContract(contractName)
//	return nil
//}

// Compile ...
func (m *Monitor) Compile(con *contract.Contract) (string, error) {
	switch con.Info.Lang {
	case "native":
		return "", nil
	case "javascript":
		jsvm, ok := m.vms["javascript"]
		if !ok {
			jsvm = Factory(con.Info.Lang)
			m.vms[con.Info.Lang] = jsvm
			err := m.vms[con.Info.Lang].Init()
			if err != nil {
				panic(err)
			}
		}
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
		vm := v8.NewVMPool(10, 10)
		vm.SetJSPath(jsPath)
		return vm
	}
	return nil
}
