package native

import "github.com/iost-official/Go-IOS-Protocol/core/contract"

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
