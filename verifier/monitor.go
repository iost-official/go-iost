package verifier

import (
	"errors"

	"github.com/iost-official/prototype/core/state"
	"github.com/iost-official/prototype/core/tx"
	"github.com/iost-official/prototype/vm"
	"github.com/iost-official/prototype/vm/lua"
)

var ErrForbiddenCall = errors.New("forbidden call")

type vmHolder struct {
	vm.VM
	//contract vm.Contract
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

func (m *vmMonitor) StartVM(contract vm.Contract) vm.VM {
	if vm, ok := m.vms[contract.Info().Prefix]; ok {
		return vm.VM
	}
	vm := m.startVM(contract)
	m.vms[contract.Info().Prefix] = vmHolder{vm}
	return m.vms[contract.Info().Prefix].VM
}

func (m *vmMonitor) startVM(contract vm.Contract) vm.VM {
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

		return &lvm
	}
	return nil
}

func (m *vmMonitor) RestartVM(contract vm.Contract) vm.VM {
	if m.hotVM == nil {
		m.hotVM = &vmHolder{m.startVM(contract)}
		return m.hotVM
	}
	m.hotVM.Restart(contract)
	return m.hotVM
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
	if m.hotVM != nil {
		m.hotVM.Stop()
		m.hotVM = nil
	}
}

func (m *vmMonitor) GetMethod(contractPrefix, methodName string, caller vm.IOSTAccount) (vm.Method, error) {
	var contract vm.Contract
	var err error
	vmh, ok := m.vms[contractPrefix]
	if !ok {
		contract, err = FindContract(contractPrefix)
		if err != nil {
			return nil, err
		}
	} else {
		switch vmh.VM.(type) {
		case *lua.VM:
			contract = vmh.VM.(*lua.VM).Contract
		default:
			panic(errors.New("unreachable"))
		}
	}
	rtn, err := contract.API(methodName)
	if err != nil {
		return nil, err
	}
	p := vm.CheckPrivilege(contract.Info(), string(caller))
	pri := rtn.Privilege()
	switch {
	case pri == vm.Private && p > 1:
		fallthrough
	case pri == vm.Protected && p >= 0:
		fallthrough
	case pri == vm.Public:
		return rtn, nil
	default:
		return nil, ErrForbiddenCall
	}

}

func (m *vmMonitor) Call(ctx vm.Context, pool state.Pool, contractPrefix, methodName string, args ...state.Value) ([]state.Value, state.Pool, uint64, error) {

	//if m.hotVM != nil && contractPrefix == m.hotVM.contract.Info().Prefix {
	//	//fmt.Println(pool.GetHM("iost", "b"))
	//	rtn, pool2, err := m.hotVM.Call(ctx, pool, methodName, args...)
	//	//fmt.Println(pool2.GetHM("iost", "b"))
	//
	//	gas := m.hotVM.PC()
	//	return rtn, pool2, gas, err
	//} never call to hotVM
	holder, ok := m.vms[contractPrefix]
	//fmt.Println("call contract:", contractPrefix)
	if !ok {
		//fmt.Println("error vm not start up")
		contract, err := FindContract(contractPrefix)
		if err != nil {
			return nil, nil, 0, err
		}

		m.StartVM(contract)
		holder, ok = m.vms[contractPrefix]
	}
	//fmt.Println(pool.GetHM("iost", "b"))
	rtn, pool2, err := holder.Call(ctx, pool, methodName, args...)
	//fmt.Println(pool2.GetHM("iost", "b"))

	gas := holder.PC()
	return rtn, pool2, gas, err
}

// FindContract  find contract from tx database
func FindContract(contractPrefix string) (vm.Contract, error) {
	//fmt.Println("22error vm not start up:", contractPrefix)
	hash := vm.PrefixToHash(contractPrefix)

	txdb := tx.TxDbInstance()
	txx, err := txdb.Get(hash)
	if err != nil {
		panic(err)
		return nil, err
	}
	//fmt.Println("found tx hash: ", common.Base58Encode(txx.Hash()))
	return txx.Contract, nil
}
