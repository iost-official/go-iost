package native_vm

import (
	"errors"

	"github.com/iost-official/Go-IOS-Protocol/core/contract"
	"github.com/iost-official/Go-IOS-Protocol/new_vm/host"
)

type VM struct {
}

func (m *VM) Init() error {
	return nil
}
func (m *VM) LoadAndCall(host *host.Host, con *contract.Contract, api string, args ...interface{}) (rtn []interface{}, cost *contract.Cost, err error) {
	//err = host.VerifyArgs(api, args...)
	//if err != nil {
	//	return nil, cost, err
	//}

	switch api {
	case "RequireAuth":
		b, cost := host.RequireAuth(args[0].(string))
		rtn = []interface{}{
			b,
		}
		return rtn, cost, nil

	case "Receipt":
		host.Receipt(args[0].(string))
		return []interface{}{}, cost, nil

	case "CallWithReceipt":
		rtn, _, err = host.CallWithReceipt(args[0].(string), args[1].(string), args[2:])
		return rtn, cost, err

	case "Transfer":
		arg2 := args[2].(int64)
		cost, err = host.Transfer(args[0].(string), args[1].(string), arg2)
		return []interface{}{}, cost, err

	case "TopUp":
		cost, err = host.TopUp(args[0].(string), args[1].(string), args[2].(int64))
		return []interface{}{}, cost, err

	case "Countermand":
		arg2 := args[2].(int64)
		cost, err = host.Countermand(args[0].(string), args[1].(string), arg2)
		return []interface{}{}, cost, err

		// 不支持在智能合约中调用, 只能放在 action 中执行, 否则会有把正在执行的智能合约更新的风险
	case "SetCode":
		con := &contract.Contract{}
		err = con.Decode(args[0].(string))
		if err != nil {
			return nil, cost, err
		}

		res, _, err := host.Call(args[0].(string), "canUpdate", args[1:])
		if err != nil {
			return nil, cost, err
		}

		if len(res) != 1 {
			return nil, cost, errors.New("return of canUpdate should have 1 argument")
		}
		if !(res[0].(bool)) {
			return nil, cost, errors.New("canUpdate return false")
		}
		cost, err := host.SetCode(args[0].(string))
		return []interface{}{}, cost, err

		// 不支持在智能合约中调用, 只能放在 action 中执行, 否则会有把正在执行的智能合约更新的风险
	case "DestroyCode":
		res, _, err := host.Call(args[0].(string), "canDestroy", args[1:])
		if err != nil {
			return nil, cost, err
		}

		if len(res) != 1 {
			return nil, cost, errors.New("return of canDestroy should have 1 argument")
		}
		if !(res[0].(bool)) {
			return nil, cost, errors.New("canDestroy return false")
		}
		host.DestroyCode(args[0].(string))
		return []interface{}{}, cost, nil

	default:
		return nil, cost, errors.New("unknown api name")

	}

	return nil, cost, errors.New("unexpected error")
}
func (m *VM) Release() {
}

func (m *VM) Compile(contract *contract.Contract) (string, error) {
	return "", nil
}
