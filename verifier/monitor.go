package verifier

import (
	"errors"

	"fmt"

	"sync"

	"github.com/iost-official/prototype/core/state"
	"github.com/iost-official/prototype/core/tx"
	"github.com/iost-official/prototype/vm"
	"github.com/iost-official/prototype/vm/lua"
)

var ErrForbiddenCall = errors.New("forbidden call")

type vmHolder struct {
	vm.VM
	Lock sync.Mutex
	//contract vm.contract
}

type vmMonitor struct {
	vms   map[string]vmHolder
	hotVM *vmHolder
}

func newVMMonitor() vmMonitor {
	return vmMonitor{
		vms:   make(map[string]vmHolder),
		hotVM: nil,
	}
}

func (m *vmMonitor) StartVM(contract vm.Contract) (vm.VM, error) {
	if vmx, ok := m.vms[contract.Info().Prefix]; ok {
		return vmx.VM, nil
	}
	vmx, err := m.startVM(contract)
	if err != nil {
		return nil, err
	}
	m.vms[contract.Info().Prefix] = vmHolder{VM: vmx, Lock: sync.Mutex{}}
	return m.vms[contract.Info().Prefix].VM, nil
}

func (m *vmMonitor) startVM(contract vm.Contract) (vm.VM, error) {
	switch contract.(type) {
	case *lua.Contract:
		var lvm lua.VM
		err := lvm.Prepare(m)
		if err != nil {
			return nil, err
		}
		err = lvm.Start(contract.(*lua.Contract))
		if err != nil {
			return nil, err
		}

		return &lvm, nil
	default:
		return nil, fmt.Errorf("contract not supported")
	}
}

func (m *vmMonitor) RestartVM(contract vm.Contract) (vm.VM, error) {
	if m.hotVM == nil {
		vmx, err := m.startVM(contract)
		if err != nil {
			return nil, err
		}
		m.hotVM = &vmHolder{VM: vmx, Lock: sync.Mutex{}}
		return m.hotVM, nil
	}
	m.hotVM.Restart(contract)
	return m.hotVM, nil
}

func (m *vmMonitor) StopVM(contract vm.Contract) {
	holder, ok := m.vms[contract.Info().Prefix]
	if !ok {
		return
	}
	holder.Lock.Lock()
	holder.Stop()
	delete(m.vms, string(contract.Hash()))
	holder.Lock.Unlock()
}

func (m *vmMonitor) Stop() {
	for _, vv := range m.vms {
		vv.Lock.Lock()
		vv.Stop()
	}
	m.vms = make(map[string]vmHolder)
	if m.hotVM != nil {
		m.hotVM.Lock.Lock()
		m.hotVM.Stop()
		m.hotVM = nil
	}
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
		contract = vmh.VM.Contract()
	}
	rtn, err := contract.API(methodName)
	if err != nil {
		return nil, err
	}

	return rtn, nil

}

func (m *vmMonitor) Call(ctx *vm.Context, pool state.Pool, contractPrefix, methodName string, args ...state.Value) ([]state.Value, state.Pool, uint64, error) {

	if m.hotVM != nil && contractPrefix == m.hotVM.Contract().Info().Prefix {
		m.hotVM.Lock.Lock()
		rtn, pool2, err := m.hotVM.Call(ctx, pool, methodName, args...)
		gas := m.hotVM.PC()
		m.hotVM.Lock.Unlock()

		return rtn, pool2, gas, err
	}
	holder, ok := m.vms[contractPrefix]
	if !ok {
		contract, err := FindContract(contractPrefix)
		if err != nil {
			return nil, pool, 0, err
		}
		_, err = m.StartVM(contract)
		if err != nil {
			return nil, pool, 0, err
		}
		holder, ok = m.vms[contract.Info().Prefix] // TODO 有危险的bug
		//return nil, pool, 0, fmt.Errorf("cannot find contract %v", contractPrefix)
	}
	holder.Lock.Lock()
	rtn, pool2, err := holder.Call(ctx, pool, methodName, args...)
	gas := holder.PC()
	holder.Lock.Unlock()
	return rtn, pool2, gas, err
}

// FindContract  find contract from tx database
func FindContract(contractPrefix string) (vm.Contract, error) {
	//fmt.Println("error vm not start up:", contractPrefix)
	hash := vm.PrefixToHash(contractPrefix)

	txdb := tx.TxDbInstance()
	txx, err := txdb.Get(hash)
	if err != nil {
		return nil, err
	}
	//fmt.Println("found tx hash: ", common.Base58Encode(txx.Hash()))
	return txx.Contract, nil
}
