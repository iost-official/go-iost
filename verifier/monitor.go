package verifier

import (
	"fmt"

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

func (m *vmMonitor) GetMethod(contractPrefix, methodName string) vm.Method {
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

func (m *vmMonitor) Call(pool state.Pool,
	contractPrefix,
	methodName string,
	args ...state.Value) ([]state.Value, state.Pool, uint64, error) { // todo 权限检查
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

// FindContract  find contract from tx database
func FindContract(contractPrefix string) vm.Contract {
	return nil
}
