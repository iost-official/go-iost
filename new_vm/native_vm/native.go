package native_vm

import (
	"errors"

	"github.com/bitly/go-simplejson"
	"github.com/iost-official/Go-IOS-Protocol/core/contract"
	"github.com/iost-official/Go-IOS-Protocol/new_vm/host"
)

var (
	ErrIssueInNormalBlock = errors.New("issueIOST should never used outsides of genesis")
)

type VM struct {
}

func (m *VM) Init() error {
	return nil
}
func (m *VM) LoadAndCall(host *host.Host, con *contract.Contract, api string, args ...interface{}) (rtn []interface{}, cost *contract.Cost, err error) {
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
		json, err := simplejson.NewJson(args[2].([]byte))
		arr, err := json.Array()
		if err != nil {
			return nil, cost, err
		}
		rtn, cost, err = host.CallWithReceipt(args[0].(string), args[1].(string), arr...)
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
		cost := contract.NewCost(1, 1, 1)
		con := &contract.Contract{}
		err = con.B64Decode(args[0].(string))
		if err != nil {
			return nil, cost, err
		}

		info, cost1 := host.TxInfo()
		cost.AddAssign(cost1)
		json, err := simplejson.NewJson(info)
		if err != nil {
			return nil, cost, err
		}

		id, err := json.Get("hash").String()
		if err != nil {
			return nil, cost, err
		}
		actId := "Contract" + id
		con.ID = actId

		cost2, err := host.SetCode(con)
		cost.AddAssign(cost2)
		return []interface{}{actId}, cost, err

		// 不支持在智能合约中调用, 只能放在 action 中执行, 否则会有把正在执行的智能合约更新的风险
	case "UpdateCode":
		cost := contract.NewCost(1, 1, 1)
		con := &contract.Contract{}
		err = con.Decode(args[0].(string))
		if err != nil {
			return nil, cost, err
		}

		cost1, err := host.UpdateCode(con, []byte(args[1].(string)))
		cost.AddAssign(cost1)
		return []interface{}{}, cost, err

		// 不支持在智能合约中调用, 只能放在 action 中执行, 否则会有把正在执行的智能合约更新的风险
	case "DestroyCode":
		cost, err = host.DestroyCode(args[0].(string))
		return []interface{}{}, cost, err

	case "IssueIOST":
		if host.Context().Value("number").(int64) != 0 {
			return []interface{}{}, contract.Cost0(), ErrIssueInNormalBlock
		}
		host.DB().SetBalance(args[0].(string), args[1].(int64))
		return []interface{}{}, contract.Cost0(), nil
	default:
		return nil, contract.NewCost(1, 1, 1), errors.New("unknown api name")

	}

	return nil, contract.NewCost(1, 1, 1), errors.New("unexpected error")
}
func (m *VM) Release() {
}

func (m *VM) Compile(contract *contract.Contract) (string, error) {
	return "", nil
}

func NativeABI() *contract.Contract {
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
