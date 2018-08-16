package new_vm

import (
	"github.com/iost-official/Go-IOS-Protocol/core/contract"
	"github.com/iost-official/Go-IOS-Protocol/new_vm/native_vm"

	"errors"

	"github.com/iost-official/Go-IOS-Protocol/new_vm/host"
	"github.com/iost-official/Go-IOS-Protocol/new_vm/v8vm"
)

var (
	ErrABINotFound    = errors.New("abi not found")
	ErrGasPriceTooBig = errors.New("gas price too big")
	ErrArgsNotEnough  = errors.New("args not enough")
	ErrArgsType       = errors.New("args type not match")
)

type Monitor struct {
	vms map[string]VM
}

func NewMonitor() *Monitor {
	m := &Monitor{
		vms: make(map[string]VM),
	}
	return m
}

func (m *Monitor) Call(h *host.Host, contractName, api string, args ...interface{}) (rtn []interface{}, cost *contract.Cost, err error) {

	c := h.DB.Contract(contractName)
	abi := c.ABI(api)
	if abi == nil {
		panic("should not reach here")
	}

	err = checkArgs(abi, args)

	if err != nil {
		return nil, contract.NewCost(0, 0, GasCheckTxFailed), err
	}

	h.Ctx = host.NewContext(h.Ctx)

	h.Ctx.Set("contract_name", contractName)
	h.Ctx.Set("abi_name", api)

	vm, ok := m.vms[c.Info.Lang]
	if !ok {
		vm = VMFactory(c.Info.Lang)
		m.vms[c.Info.Lang] = vm
		m.vms[c.Info.Lang].Init()
	}
	rtn, cost, err = vm.LoadAndCall(h, c, api, args...)

	payment, ok := h.Ctx.GValue("abi_payment").(int)
	if !ok {
		payment = int(abi.Payment)
	}
	switch payment {
	case 1:
		var gasPrice = h.Ctx.Value("gas_price").(int64)
		if abi.GasPrice < gasPrice {
			return nil, nil, ErrGasPriceTooBig
		}

		b := h.DB.Balance(host.ContractGasPrefix + contractName)
		if b > gasPrice*cost.ToGas() {
			h.PayCost(cost, host.ContractGasPrefix+contractName)
			cost = contract.Cost0()
		}

	default:
	}

	h.Ctx = h.Ctx.Base()

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
//func (m *Monitor) Destory(contractName string) error {
//	m.ho.db.DelContract(contractName)
//	return nil
//}

func (m *Monitor) Compile(con *contract.Contract) (string, error) {
	switch con.Info.Lang {
	case "native":
		return "", nil
	case "javascript":
		jsvm := m.vms["javascript"]
		return jsvm.Compile(con)
	}
	return "", errors.New("vm unsupported")
}

func checkArgs(abi *contract.ABI, args []interface{}) error {
	if len(abi.Args) > len(args) {
		return ErrArgsNotEnough
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
			return ErrArgsType
		}
	}
	return nil
}

func VMFactory(lang string) VM {
	switch lang {
	case "native":
		return &native_vm.VM{}
	case "javascript":
		vm := v8.NewVMPool(10)
		vm.SetJSPath("./v8vm/v8/libjs/")
		return vm
	}
	return nil
}
