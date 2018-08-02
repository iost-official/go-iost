package new_vm

import (
	"context"

	"github.com/iost-official/Go-IOS-Protocol/core/contract"
	"github.com/iost-official/Go-IOS-Protocol/core/new_tx"
	"github.com/iost-official/Go-IOS-Protocol/new_vm/database"
)

type Monitor struct {
	db   *database.Visitor
	vms  map[string]VM
	host *Host
}

func NewMonitor(cb database.IMultiValue, cacheLength int) *Monitor {
	visitor := database.NewVisitor(cacheLength, cb)
	return &Monitor{
		db: visitor,
		host: &Host{
			ctx: nil,
			db:  visitor,
		},
	}
}

func (m *Monitor) Call(ctx context.Context, contractName, api string, args ...string) (rtn []string, receipt *tx.Receipt, cost *contract.Cost, err error) {
	c := m.db.Contract(contractName)

	ctx2 := ctx

	switch c.ContractInfo.Payment {
	case contract.ContractPay:
		ctx2 = context.WithValue(ctx, "cost_limit", c.ContractInfo.Limit)

	}

	if vm, ok := m.vms[c.Lang]; ok {
		rtn, cost, err = vm.LoadAndCall(ctx2, c, api, args...)
	} else {
		vm = VMFactory(c.Lang)
		m.vms[c.Lang] = vm
		m.vms[c.Lang].Init(m.host)
		rtn, cost, err = vm.LoadAndCall(ctx2, c, api, args...)
	}
	if err != nil {
		receipt = &tx.Receipt{
			Type:    tx.SystemDefined,
			Content: err.Error(),
		}
	}
	receipt = &tx.Receipt{
		Type:    tx.SystemDefined,
		Content: "success",
	}
	switch c.ContractInfo.Payment { // todo move to higher layer
	case contract.ContractPay:
		m.host.LoadContext(ctx2).Transfer(contractName, ctx.Value("witness").(string), int64(c.ContractInfo.GasPrice*cost.ToGas()))
		cost = &contract.Cost{}
	}

	return
}

func (m *Monitor) Update(contractName string, newContract *contract.Contract) error {
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
