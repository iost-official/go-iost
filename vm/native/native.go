package native

import (
	"errors"

	"github.com/bitly/go-simplejson"
	"github.com/iost-official/Go-IOS-Protocol/core/contract"
	"github.com/iost-official/Go-IOS-Protocol/vm/host"
)

// var ...
var (
	ErrIssueInNormalBlock = errors.New("issueIOST should never used outsides of genesis")
)

// VM ...
type VM struct {
}

// Init ...
func (m *VM) Init() error {
	return nil
}

// LoadAndCall ...
// nolint
func (m *VM) LoadAndCall(h *host.Host, con *contract.Contract, api string, args ...interface{}) (rtn []interface{}, cost *contract.Cost, err error) {
	switch api {
	case "RequireAuth":
		var b bool
		b, cost = h.RequireAuth(args[0].(string))
		rtn = []interface{}{
			b,
		}
		return rtn, cost, nil

	case "Receipt":
		cost = h.Receipt(args[0].(string))
		return []interface{}{}, cost, nil

	case "CallWithReceipt":
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

	case "Transfer":
		arg2 := args[2].(int64)
		cost, err = h.Transfer(args[0].(string), args[1].(string), arg2)
		return []interface{}{}, cost, err

	case "TopUp":
		cost, err = h.TopUp(args[0].(string), args[1].(string), args[2].(int64))
		return []interface{}{}, cost, err

	case "Countermand":
		arg2 := args[2].(int64)
		cost, err = h.Countermand(args[0].(string), args[1].(string), arg2)
		return []interface{}{}, cost, err

		// 不支持在智能合约中调用, 只能放在 action 中执行, 否则会有把正在执行的智能合约更新的风险
	case "SetCode":
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

		// 不支持在智能合约中调用, 只能放在 action 中执行, 否则会有把正在执行的智能合约更新的风险
	case "UpdateCode":
		cost := contract.Cost0()
		con := &contract.Contract{}
		err = con.Decode(args[0].(string))
		if err != nil {
			return nil, host.CommonErrorCost(1), err
		}

		cost1, err := h.UpdateCode(con, []byte(args[1].(string)))
		cost.AddAssign(cost1)
		return []interface{}{}, cost, err

		// 不支持在智能合约中调用, 只能放在 action 中执行, 否则会有把正在执行的智能合约更新的风险
	case "DestroyCode":
		cost, err = h.DestroyCode(args[0].(string))
		return []interface{}{}, cost, err

	case "IssueIOST":
		if h.Context().Value("number").(int64) != 0 {
			return []interface{}{}, contract.Cost0(), ErrIssueInNormalBlock
		}
		h.DB().SetBalance(args[0].(string), args[1].(int64))
		return []interface{}{}, contract.Cost0(), nil
	default:
		return nil, host.CommonErrorCost(1), errors.New("unknown api name")

	}
}

// Release ...
func (m *VM) Release() {
}

// Compile ...
func (m *VM) Compile(contract *contract.Contract) (string, error) {
	return "", nil
}

// ABI ...
func ABI() *contract.Contract {
	return &contract.Contract{
		ID:   "iost.system",
		Code: "codes",
		Info: &contract.Info{
			Lang:        "native",
			VersionCode: "1.0.0",
			Abis: []*contract.ABI{
				{
					Name:     "RequireAuth",
					Args:     []string{"string"},
					Payment:  0,
					GasPrice: int64(1000),
					Limit:    contract.NewCost(100, 100, 100),
				},
				{
					Name:     "Receipt",
					Args:     []string{"string"},
					Payment:  0,
					GasPrice: int64(1000),
					Limit:    contract.NewCost(100, 100, 100),
				},
				{
					Name:     "CallWithReceipt",
					Args:     []string{"string", "string", "json"},
					Payment:  0,
					GasPrice: int64(1000),
					Limit:    contract.NewCost(100, 100, 100),
				},
				{
					Name:     "Transfer",
					Args:     []string{"string", "string", "number"},
					Payment:  0,
					GasPrice: int64(1000),
					Limit:    contract.NewCost(100, 100, 100),
				},
				{
					Name:     "TopUp",
					Args:     []string{"string", "string", "number"},
					Payment:  0,
					GasPrice: int64(1000),
					Limit:    contract.NewCost(100, 100, 100),
				},
				{
					Name:     "Countermand",
					Args:     []string{"string", "string", "number"},
					Payment:  0,
					GasPrice: int64(1000),
					Limit:    contract.NewCost(100, 100, 100),
				},
				{
					Name:     "SetCode",
					Args:     []string{"string"},
					Payment:  0,
					GasPrice: int64(1000),
					Limit:    contract.NewCost(100, 100, 100),
				},
				{
					Name:     "UpdateCode",
					Args:     []string{"string"},
					Payment:  0,
					GasPrice: int64(1000),
					Limit:    contract.NewCost(100, 100, 100),
				},
				{
					Name:     "DestroyCode",
					Args:     []string{"string"},
					Payment:  0,
					GasPrice: int64(1000),
					Limit:    contract.NewCost(100, 100, 100),
				},
				{
					Name:     "IssueIOST",
					Args:     []string{"string", "number"},
					Payment:  0,
					GasPrice: int64(1000),
					Limit:    contract.NewCost(100, 100, 100),
				},
			},
		},
	}
}
