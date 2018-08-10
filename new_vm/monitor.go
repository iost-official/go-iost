package new_vm

import (
	"context"

	"github.com/iost-official/Go-IOS-Protocol/core/contract"
	"github.com/iost-official/Go-IOS-Protocol/new_vm/native_vm"

	"errors"
)

var (
	ErrABINotFound    = errors.New("abi not found")
	ErrGasPriceTooBig = errors.New("gas price too big")
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
	abi := c.ABI(api)
	if abi == nil {
		return nil, nil, ErrABINotFound
	}
	ctx := host.Context()

	host.ctx = context.WithValue(host.ctx, "abi_config", abi)
	host.ctx = context.WithValue(host.ctx, "contract_name", contractName)
	host.ctx = context.WithValue(host.ctx, "abi_name", api)

	vm, ok := m.vms[c.Info.Lang]
	if !ok {
		vm = VMFactory(c.Info.Lang)
		m.vms[c.Info.Lang] = vm
		m.vms[c.Info.Lang].Init()
	}
	rtn, cost, err = vm.LoadAndCall(host, c, api, args...)

	payment := host.ctx.Value("abi_config").(*contract.ABI).Payment // TODO 预编译
	switch payment {
	case 1:
		var gasPrice = host.ctx.Value("gas_price").(int64) // TODO 判断大于0
		if abi.GasPrice < gasPrice {
			return nil, nil, ErrGasPriceTooBig
		}
		host.PayCost(cost, contractName, gasPrice)
		cost = contract.Cost0()
	default:
		//fmt.Println("user paid for", args[0])
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
	switch lang {
	case "native":
		return &native_vm.VM{}
	}
	return nil
}
