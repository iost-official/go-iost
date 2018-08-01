package new_vm

import (
	"context"

	"github.com/iost-official/Go-IOS-Protocol/core/new_tx"
	"github.com/iost-official/Go-IOS-Protocol/new_vm/database"
)

type Monitor struct {
	db  database.Visitor
	vms map[string]VM
}

func (m *Monitor) Call(ctx context.Context, contractName, api string, args ...string) (rtn []string, receipt tx.Receipt) {
	contract := m.db.GetContract(contractName)

	var err error
	if vm, ok := m.vms[contract.Lang]; ok {
		rtn, err = vm.LoadAndCall(ctx, contract, api, args...)
	} else {
		vm = VMFactory(contract.Lang)
		m.vms[contract.Lang] = vm
		rtn, err = vm.LoadAndCall(ctx, contract, api, args...)
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

	m.db.SetContract(newContract)
	return nil

}

func (m *Monitor) Destory(contractName string) error {
	m.db.DelContract(contractName)
	return nil
}

func VMFactory(lang string) VM {
	return nil
}
