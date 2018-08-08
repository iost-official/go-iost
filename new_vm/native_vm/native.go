package native_vm

import (
	"github.com/iost-official/Go-IOS-Protocol/core/contract"
	"github.com/iost-official/Go-IOS-Protocol/new_vm"
	"strconv"
	"errors"
)

type VM struct {
}

func (m *VM) Init() error {
	return nil
}
func (m *VM) LoadAndCall(host *new_vm.Host, cont *contract.Contract, api string, args ...interface{}) (rtn []string, cost *contract.Cost, err error) {
	// todo 检查参数类型收gas
	switch api {
	case "RequireAuth":
		if len(args) != 1 {
			return nil, host.Cost(), errors.New("RequireAuth should have 1 arg")
		}
		rtn = []string{
			strconv.FormatBool(host.RequireAuth(args[0].(string))),
		}
		return rtn, host.Cost(), nil

	case "Receipt":
		if len(args) != 1 {
			return nil, host.Cost(), errors.New("Receipt should have 1 arg")
		}
		host.Receipt(args[0].(string))
		return []string{}, host.Cost(), nil

	case "CallWithReceipt":
		if len(args) < 2 {
			return nil, host.Cost(), errors.New("CallWithReceipt should have at least 2 args")
		}
		rtn, _, err = host.CallWithReceipt(args[0].(string), args[1].(string), args[2:])
		return rtn, host.Cost(), err

	case "Transfer":
		if len(args) != 3 {
			return nil, host.Cost(), errors.New("Transfer should have 3 args")
		}
		err = host.Transfer(args[0].(string), args[1].(string), args[2].(int64))
		return []string{}, host.Cost(), err

	case "TopUp":
		if len(args) != 3 {
			return nil, host.Cost(), errors.New("Topup should have 3 args")
		}
		err = host.TopUp(args[0].(string), args[1].(string), args[2].(int64))
		return []string{}, host.Cost(), err

	case "Countermand":
		if len(args) != 3 {
			return nil, host.Cost(), errors.New("Countermand should have 3 args")
		}
		err = host.Countermand(args[0].(string), args[1].(string), args[2].(int64))
		return []string{}, host.Cost(), err

		// 不支持在智能合约中调用, 只能放在 action 中执行, 否则会有把正在执行的智能合约更新的风险
	case "SetCode":
		// todo 预编译
		if len(args) < 1 {
			return nil, host.Cost(), errors.New("SetCode should have at least 1 args")
		}
		con := &contract.Contract{}
		err = con.Decode(args[0].(string))
		if err != nil {
			return nil, host.Cost(), err
		}

		res, _, err := host.Call(args[0].(string), "canUpdate", args[1:])
		if err != nil {
			return nil, host.Cost(), err
		}

		if len(res) != 1 {
			return nil, host.Cost(), errors.New("return of canUpdate should have 1 argument")
		}
		// todo check res[0]
		host.SetCode(args[0].(string))
		return []string{}, host.Cost(), nil

		// 不支持在智能合约中调用, 只能放在 action 中执行, 否则会有把正在执行的智能合约更新的风险
	case "DestroyCode":
		if len(args) < 1 {
			return nil, host.Cost(), errors.New("DestroyCode should have at least 1 args")
		}
		res, _, err := host.Call(args[0].(string), "canDestroy", args[1:])
		if err != nil {
			return nil, host.Cost(), err
		}

		if len(res) != 1 {
			return nil, host.Cost(), errors.New("return of canDestroy should have 1 argument")
		}

		// todo check res[0]
		host.DestroyCode(args[0].(string))
		return []string{}, host.Cost(), nil

	default:
		return nil, host.Cost(), errors.New("unknown api name")

	}

	return nil, host.Cost(), errors.New("unexpected error")
}
func (m *VM) Release() {
}
