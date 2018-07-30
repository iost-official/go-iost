package new_vm

import (
	"context"

	"github.com/iost-official/Go-IOS-Protocol/core/new_tx"
)

type Monitor struct {
	Pool
	vms map[string]VM
}

func (m *Monitor) Call(ctx context.Context, contractName, api string, args ...string) (rtn []string, receipt tx.Receipt) {
	contract, err := m.Contract(contractName)

	if err != nil {
		panic(err)
	}

	if vm, ok := m.vms[contract.Lang]; ok {
		vm.Load(contract)
		rtn, err = vm.Call(ctx, api, args...)
	} else {
		vm = VMFactory(contract.Lang)
		m.vms[contract.Lang] = vm
		vm.Load(contract)
		rtn, err = vm.Call(ctx, api, args...)
	}
	if err != nil {
		// todo make err receipt
	}

	// todo make success receipt
	return
}

func (m *Monitor) Update(contractName string, newContract *Contract) error {
	err := m.Destory(contractName)
	if err != nil {
		return err
	}

	return SetContract(newContract)

}

func (m *Monitor) Destory(contractName string) error {
	m.Pool.DeleteContract(contractName)
	// TODO  从数据库中删除contract
	return nil
}

func VMFactory(lang string) VM {
	return nil
}
