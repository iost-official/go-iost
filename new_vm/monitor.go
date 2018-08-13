package new_vm

import (
	"context"

	"github.com/iost-official/Go-IOS-Protocol/core/contract"
	"github.com/iost-official/Go-IOS-Protocol/new_vm/native_vm"

	"errors"

	"github.com/iost-official/Go-IOS-Protocol/new_vm/host"
	"github.com/iost-official/Go-IOS-Protocol/new_vm/v8vm"
)

var (
	ErrABINotFound    = errors.New("abi not found")
	ErrGasPriceTooBig = errors.New("gas price too big")
	ErrArgsNotEnough  = errors.New("args not enough")
	ErrArgsType       = errors.New("args type not match")
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

func (m *Monitor) Call(host *host.Host, contractName, api string, args ...interface{}) (rtn []interface{}, cost *contract.Cost, err error) {

	c := host.DB.Contract(contractName)
	abi := c.ABI(api)
	if abi == nil {
		return nil, nil, ErrABINotFound
	}

	err = checkArgs(abi, args)

	if err != nil {
		return nil, nil, err // todo check cost
	}

	ctx := host.Context()

	host.Ctx = context.WithValue(host.Ctx, "abi_config", abi)
	host.Ctx = context.WithValue(host.Ctx, "contract_name", contractName)
	host.Ctx = context.WithValue(host.Ctx, "abi_name", api)

	vm, ok := m.vms[c.Info.Lang]
	if !ok {
		vm = VMFactory(c.Info.Lang)
		m.vms[c.Info.Lang] = vm
		m.vms[c.Info.Lang].Init()
	}
	rtn, cost, err = vm.LoadAndCall(host, c, api, args...)

	payment := host.Ctx.Value("abi_config").(*contract.ABI).Payment
	switch payment {
	case 1:
		var gasPrice = host.Ctx.Value("gas_price").(int64) // TODO 判断大于0
		if abi.GasPrice < gasPrice {
			return nil, nil, ErrGasPriceTooBig
		}

		b := host.DB.Balance(contractName)
		if b > gasPrice*cost.ToGas() {
			host.PayCost(cost, contractName, gasPrice)
			cost = contract.Cost0()
		}

	default:
		//fmt.Println("user paid for", args[0])
	}

	host.Ctx = ctx

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

func checkArgs(abi *contract.ABI, args []interface{}) error {
	if len(abi.Args) > len(args) {
		return ErrArgsNotEnough
	}

	for i, t := range abi.Args {
		var ok bool
		switch t {
		case "string":
			_, ok = args[i].(string)
		case "number":
			_, ok = args[i].(uint64)
		case "bool":
			_, ok = args[i].(bool)
		case "json":
			_, ok = args[i].([]byte)
		}
		if !ok {
			return ErrArgsType
		}
	}
	return nil
}

func VMFactory(lang string) VM {
	switch lang {
	case "native":
		return &native_vm.VM{}
	case "javascript":
		vm := v8.NewVM()
		vm.SetJSPath("/Users/hepeijian/go/src/github.com/iost-official/Go-IOS-Protocol/new_vm/v8vm/v8/libjs/")
		return vm
	}
	return nil
}
