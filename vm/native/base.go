package native

import (
	"github.com/bitly/go-simplejson"
	"github.com/iost-official/Go-IOS-Protocol/core/contract"
	"github.com/iost-official/Go-IOS-Protocol/vm/host"
	"github.com/pkg/errors"
)

// var .
var (
	requireAuth = &abi{
		name: "RequireAuth",
		do: func(h *host.Host, args ...interface{}) (rtn []interface{}, cost *contract.Cost, err error) {
			var b bool
			b, cost = h.RequireAuth(args[0].(string))
			rtn = []interface{}{
				b,
			}
			return rtn, cost, nil
		},
	}
	receipt = &abi{
		name: "Receipt",
		do: func(h *host.Host, args ...interface{}) (rtn []interface{}, cost *contract.Cost, err error) {
			cost = h.Receipt(args[0].(string))
			return []interface{}{}, cost, nil
		},
	}
	callWithReceipt = &abi{
		name: "CallWithReceipt",
		do: func(h *host.Host, args ...interface{}) (rtn []interface{}, cost *contract.Cost, err error) {
			var json *simplejson.Json
			json, err = simplejson.NewJson(args[2].([]byte))
			if err != nil {
				return nil, host.CommonErrorCost(1), err
			}
			arr, err := json.Array()
			if err != nil {
				return nil, host.CommonErrorCost(2), err
			}
			rtn, cost, err = h.CallWithReceipt(args[0].(string), args[1].(string), arr...)
			return rtn, cost, err
		},
	}
	transfer = &abi{
		name: "Transfer",
		do: func(h *host.Host, args ...interface{}) (rtn []interface{}, cost *contract.Cost, err error) {

			arg2 := args[2].(int64)
			cost, err = h.Transfer(args[0].(string), args[1].(string), arg2)
			return []interface{}{}, cost, err
		},
	}
	topUp = &abi{
		name: "TopUp",
		do: func(h *host.Host, args ...interface{}) (rtn []interface{}, cost *contract.Cost, err error) {

			cost, err = h.TopUp(args[0].(string), args[1].(string), args[2].(int64))
			return []interface{}{}, cost, err
		},
	}
	countermand = &abi{
		name: "",
		do: func(h *host.Host, args ...interface{}) (rtn []interface{}, cost *contract.Cost, err error) {

			arg2 := args[2].(int64)
			cost, err = h.Countermand(args[0].(string), args[1].(string), arg2)
			return []interface{}{}, cost, err
		},
	}
	// 不支持在智能合约中调用, 只能放在 action 中执行, 否则会有把正在执行的智能合约更新的风险
	setCode = &abi{
		name: "SetCode",
		do: func(h *host.Host, args ...interface{}) (rtn []interface{}, cost *contract.Cost, err error) {

			cost = contract.Cost0()
			con := &contract.Contract{}
			err = con.B64Decode(args[0].(string))
			if err != nil {
				return nil, host.CommonErrorCost(1), err
			}

			info, cost1 := h.TxInfo()
			cost.AddAssign(cost1)
			var json *simplejson.Json
			json, err = simplejson.NewJson(info)
			if err != nil {
				return nil, cost, err
			}

			var id string
			id, err = json.Get("hash").String()
			if err != nil {
				return nil, cost, err
			}
			actID := "Contract" + id
			con.ID = actID

			var cost2 *contract.Cost
			cost2, err = h.SetCode(con)
			cost.AddAssign(cost2)
			return []interface{}{actID}, cost, err
		},
	}
	// 不支持在智能合约中调用, 只能放在 action 中执行, 否则会有把正在执行的智能合约更新的风险
	updateCode = &abi{
		name: "UpdateCode",
		do: func(h *host.Host, args ...interface{}) (rtn []interface{}, cost *contract.Cost, err error) {
			cost = contract.Cost0()
			con := &contract.Contract{}
			err = con.Decode(args[0].(string))
			if err != nil {
				return nil, host.CommonErrorCost(1), err

			}

			cost1, err := h.UpdateCode(con, []byte(args[1].(string)))
			cost.AddAssign(cost1)
			return []interface{}{}, cost, err
		},
	}
	// 不支持在智能合约中调用, 只能放在 action 中执行, 否则会有把正在执行的智能合约更新的风险
	destroyCode = &abi{
		name: "DestroyCode",
		do: func(h *host.Host, args ...interface{}) (rtn []interface{}, cost *contract.Cost, err error) {

			cost, err = h.DestroyCode(args[0].(string))
			return []interface{}{}, cost, err
		},
	}
	issueIOST = &abi{
		name: "IssueIOST",
		do: func(h *host.Host, args ...interface{}) (rtn []interface{}, cost *contract.Cost, err error) {

			if h.Context().Value("number").(int64) != 0 {
				return []interface{}{}, contract.Cost0(), errors.New("issue IOST in normal block")
			}
			h.DB().SetBalance(args[0].(string), args[1].(int64))
			return []interface{}{}, contract.Cost0(), nil
		},
	}
)
