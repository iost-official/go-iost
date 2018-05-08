package verifier

import (
	"github.com/iost-official/prototype/vm/lua"
	"github.com/iost-official/prototype/core/state"
	"fmt"
	"github.com/iost-official/prototype/vm"
)

type VMHolder struct {
	vm.VM
}

type VMMonitor struct {
	vms map[string]VMHolder
}

func (m *VMMonitor) StartVM(contract vm.Contract) vm.VM {
	if _, ok := m.vms[contract.Info().Prefix]; ok {
		return nil
	}
	switch contract.(type) {
	case *lua.Contract:
		var lvm lua.VM
		lvm.Prepare(contract.(*lua.Contract), m)
		lvm.Start()
		m.vms[contract.Info().Prefix] = VMHolder{&lvm}
		return &lvm
	}
	return nil
}

func (m *VMMonitor) StopVm(contract vm.Contract) {
	m.vms[string(contract.Hash())].Stop()
	delete(m.vms, string(contract.Hash()))
}

func (m *VMMonitor) Stop() {
	for _, vm := range m.vms {
		vm.Stop()
	}
	m.vms = make(map[string]VMHolder)
}

func (m *VMMonitor) GetMethod(contractPrefix, methodName string) vm.Method {
	contract := FindContract(contractPrefix)
	method, _ := contract.Api(methodName)
	return method
}

func (m *VMMonitor) Call(pool state.Pool, contractPrefix, methodName string, args ...state.Value) ([]state.Value, state.Pool, error) {
	vm, ok := m.vms[contractPrefix]
	if !ok {
		contract := FindContract(contractPrefix)
		if contract == nil {
			return nil, nil, fmt.Errorf("contract not found")
		}
		vm = m.vms[contractPrefix]
	}
	return vm.Call(pool, methodName, args...)
}

func FindContract(contractPrefix string) vm.Contract {
	return nil
}
