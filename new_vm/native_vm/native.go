package native_vm

import (
	"errors"

	"github.com/iost-official/Go-IOS-Protocol/core/contract"
	"github.com/iost-official/Go-IOS-Protocol/new_vm/host"
	"github.com/bitly/go-simplejson"
	"github.com/iost-official/Go-IOS-Protocol/common"
)

type VM struct {
}

func (m *VM) Init() error {
	return nil
}
func (m *VM) LoadAndCall(host *host.Host, con *contract.Contract, api string, args ...interface{}) (rtn []interface{}, cost *contract.Cost, err error) {
	// todo cost add base cost
	switch api {
	case "RequireAuth":
		b, cost := host.RequireAuth(args[0].(string))
		rtn = []interface{}{
			b,
		}
		return rtn, cost, nil

	case "Receipt":
		cost := host.Receipt(args[0].(string))
		return []interface{}{}, cost, nil

	case "CallWithReceipt":
		rtn, cost, err = host.CallWithReceipt(args[0].(string), args[1].(string), args[2:])
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

		info, cost := host.TxInfo()
		json, err := simplejson.NewJson(info)
		if err != nil {
			return nil, cost, err
		}
		id, err := json.Get("hash").Bytes()
		if err != nil {
			return nil, cost, err
		}
		con.ID = common.Base58Encode(id)

		cost1, err := host.SetCode(con)
		cost.AddAssign(cost1)
		return []interface{}{}, cost, err

		// 不支持在智能合约中调用, 只能放在 action 中执行, 否则会有把正在执行的智能合约更新的风险
	case "UpdateCode":
		con := &contract.Contract{}
		err = con.Decode(args[0].(string))
		if err != nil {
			return nil, cost, err
		}

		cost, err := host.UpdateCode(con, []byte(args[1].(string)))
		return []interface{}{}, cost, err

		// 不支持在智能合约中调用, 只能放在 action 中执行, 否则会有把正在执行的智能合约更新的风险
	case "DestroyCode":
		cost, err = host.DestroyCode(args[0].(string))
		return []interface{}{}, cost, err

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
