package new_vm

import (
	"context"

	"github.com/iost-official/Go-IOS-Protocol/core/contract"
)

type Monitor struct {
	//db   *database.Visitor
	vms map[string]VM
	//host *Host
}

func NewMonitor( /*cb database.IMultiValue, cacheLength int*/ ) *Monitor {
	//visitor := database.NewVisitor(cacheLength, cb)
	m := &Monitor{
		//db: visitor,
		//host: &Host{
		//	ctx:  context.Background(),
		//	db:   visitor,
		//	cost: &contract.Cost{},
		//},
		vms: make(map[string]VM),
	}
	//m.host.monitor = m
	return m
}

func (m *Monitor) Call(host *Host, contractName, api string, args ...interface{}) (rtn []interface{}, cost *contract.Cost, err error) {

	c := host.db.Contract(contractName)
	ctx := host.Context()

	host.ctx = context.WithValue(host.ctx, "abi_config", make(map[string]*string))
	host.ctx = context.WithValue(host.ctx, "contract_name", contractName)
	host.ctx = context.WithValue(host.ctx, "abi_name", api)

	vm, ok := m.vms[c.Lang]
	if !ok {
		vm = VMFactory(c.Lang)
		m.vms[c.Lang] = vm
		m.vms[c.Lang].Init()
	}
	rtn, cost, err = vm.LoadAndCall(host, c, api, args...)

	payment := host.ctx.Value("abi_config").(map[string]*string)["payment"] // TODO 预编译
	gasPrice := host.ctx.Value("gas_price").(uint64)
	switch {
	case payment != nil && *payment == "contract_pay":
		host.PayCost(cost, contractName, gasPrice)
		cost = contract.Cost0()
	}

	host.ctx = ctx

	return
}

//func (m *Monitor) Update(contractName string, newContract *contract.Contract) error {
//	err := m.Destory(contractName)
//	if err != nil {
//		return err
//	}
//	m.host.db.SetContract(newContract)
//	return nil
//}
//
//func (m *Monitor) Destory(contractName string) error {
//	m.host.db.DelContract(contractName)
//	return nil
//}

func VMFactory(lang string) VM {
	return nil
}
