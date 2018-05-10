package verifier

import (
	"fmt"

	"github.com/iost-official/prototype/core/state"
	"github.com/iost-official/prototype/vm"
	"github.com/iost-official/prototype/vm/lua"
)

type VMHolder struct {
	vm.VM
	contract vm.Contract
}

type VMMonitor struct {
	vms map[string]VMHolder
}

func NewVMMonitor() VMMonitor {
	return VMMonitor{
		vms: make(map[string]VMHolder),
	}
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
		m.vms[contract.Info().Prefix] = VMHolder{&lvm, contract}
		return &lvm
	}
	return nil
}

func (m *VMMonitor) StopVm(contract vm.Contract) {
	m.vms[contract.Info().Prefix].Stop()
	delete(m.vms, string(contract.Hash()))
}

func (m *VMMonitor) Stop() {
	for _, vv := range m.vms {
		vv.Stop()
	}
	m.vms = make(map[string]VMHolder)
}

func (m *VMMonitor) GetMethod(contractPrefix, methodName string) vm.Method {
	var contract vm.Contract
	vmh, ok := m.vms[contractPrefix]
	if !ok {
		contract = FindContract(contractPrefix)
	} else {
		contract = vmh.contract
	}
	method, _ := contract.Api(methodName)
	return method
}

func (m *VMMonitor) Call(pool state.Pool, contractPrefix, methodName string, args ...state.Value) ([]state.Value, state.Pool, uint64, error) {
	holder, ok := m.vms[contractPrefix]
	if !ok {
		contract := FindContract(contractPrefix)
		if contract == nil {
			return nil, nil, 0, fmt.Errorf("contract not found")
		}
		holder = m.vms[contractPrefix]
	}
	rtn, pool, err := holder.Call(pool, methodName, args...)
	gas := holder.PC()
	return rtn, pool, gas, err
}

func FindContract(contractPrefix string) vm.Contract {
	return nil
}
