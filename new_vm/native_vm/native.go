package native_vm

import (
	"errors"
	"strconv"

	"github.com/iost-official/Go-IOS-Protocol/core/contract"
	"github.com/iost-official/Go-IOS-Protocol/new_vm/host"
)

type VM struct {
}

func (m *VM) Init() error {
	return nil
}
func (m *VM) LoadAndCall(host host.IHost, contract *contract.Contract, api string, args ...interface{}) (rtn []interface{}, cost *contract.Cost, err error) {
	//err = host.VerifyArgs(api, args...)
	//if err != nil {
	//	return nil, host.Cost(), err
	//}
	switch api {
	case "RequireAuth":
		b := strconv.FormatBool(host.RequireAuth(args[0].(string)))
		rtn = []interface{}{
			b,
		}
		return rtn, host.Cost(), nil

	case "Receipt":
		host.Receipt(args[0].(string))
		return []interface{}{}, host.Cost(), nil

	case "CallWithReceipt":
		//rtn = []string{
		//	// todo CallWithReceipt return value
		//	// strconv.FormatBool(host.CallWithReceipt(args[0], args[1], args[2:])),
		//}
		return rtn, host.Cost(), nil

	case "Transfer":
		arg2 := args[2].(int64)
		err = host.Transfer(args[0].(string), args[1].(string), arg2)
		return []interface{}{}, host.Cost(), err

	case "TopUp":
		arg2, err := strconv.ParseInt(args[2].(string), 10, 64)
		if err != nil {
			return nil, host.Cost(), err
		}
		err = host.TopUp(args[0].(string), args[1].(string), arg2)
		return []interface{}{}, host.Cost(), err

	case "Countermand":
		arg2 := args[2].(int64)
		if err != nil {
			return nil, host.Cost(), err
		}
		err = host.Countermand(args[0].(string), args[1].(string), arg2)
		return []interface{}{}, host.Cost(), err

	case "SetCode":
		// todo set code

	default:
		return nil, host.Cost(), errors.New("unknown api name")

	}

	return nil, host.Cost(), errors.New("unexpected error")
}
func (m *VM) Release() {
}
