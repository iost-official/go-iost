package verifier

import (
	"github.com/iost-official/prototype/core/state"
	"github.com/iost-official/prototype/vm"
	"github.com/iost-official/prototype/vm/lua"
)

type vmHolder struct {
	vm.VM
	contract vm.Contract
}

type vmMonitor struct {
	vms map[string]vmHolder
}

func newVMMonitor() vmMonitor {
	return vmMonitor{
		vms: make(map[string]vmHolder),
	}
}

func (m *vmMonitor) StartVM(contract vm.Contract) vm.VM {
	if _, ok := m.vms[contract.Info().Prefix]; ok {
		return nil
	}

	switch contract.(type) {
	case *lua.Contract:
		var lvm lua.VM
		err := lvm.Prepare(contract.(*lua.Contract), m)
		if err != nil {
			panic(err)
		}
		err = lvm.Start()
		if err != nil {
			panic(err)
		}
		m.vms[contract.Info().Prefix] = vmHolder{&lvm, contract}
		return &lvm
	}
	return nil
}

func (m *vmMonitor) RestartVM(contract vm.Contract) vm.VM {
	if _, ok := m.vms[contract.Info().Prefix]; ok {
		m.StopVM(contract)
	}
	return m.StartVM(contract)
}

func (m *vmMonitor) StopVM(contract vm.Contract) {
	m.vms[contract.Info().Prefix].Stop()
	delete(m.vms, string(contract.Hash()))
}

func (m *vmMonitor) Stop() {
	for _, vv := range m.vms {
		vv.Stop()
	}
	m.vms = make(map[string]vmHolder)
}

func (m *vmMonitor) GetMethod(contractPrefix, methodName string) (vm.Method, error) {
	var contract vm.Contract
	var err error
	vmh, ok := m.vms[contractPrefix]
	if !ok {
		contract, err = FindContract(contractPrefix)
		if err != nil {
			return nil, err
		}
	} else {
		contract = vmh.contract
	}
	return contract.Api(methodName)
}

func (m *vmMonitor) Call(pool state.Pool,
	contractPrefix,
	methodName string,
	args ...state.Value) ([]state.Value, state.Pool, uint64, error) { // todo 权限检查
	holder, ok := m.vms[contractPrefix]
	if !ok {
		contract, err := FindContract(contractPrefix)
		if err != nil {
			return nil, nil, 0, err
		}
		m.StartVM(contract)
		holder = m.vms[contractPrefix]
	}
	//switch holder.VM.(type) {
	//case *lua.VM:
	//	holder.VM.(*lua.VM).L.PCount = 0 // 注意：L会复用因此会保留PCount，需要在
	//}
	rtn, pool, err := holder.Call(pool, methodName, args...)
	gas := holder.PC()
	return rtn, pool, gas, err
}

// FindContract  find contract from tx database
func FindContract(contractPrefix string) (vm.Contract, error) {
	code2 := `function sayHi(name)
	return "hi " .. name
end`
	sayHi := lua.NewMethod("sayHi", 1, 1)
	lc2 := lua.NewContract(vm.ContractInfo{Prefix: "con2", GasLimit: 1000, Price: 1, Sender: vm.IOSTAccount("ahaha")},
		code2, sayHi, sayHi)
	return &lc2, nil
}
