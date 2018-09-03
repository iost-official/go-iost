package native

import (
	"sort"

	"github.com/iost-official/Go-IOS-Protocol/core/contract"
)

// ABI ...
func ABI() *contract.Contract {
	c := &contract.Contract{
		ID:   "iost.system",
		Code: "codes",
		Info: &contract.Info{
			Lang:        "native",
			VersionCode: "1.0.0",
			Abis:        make([]*contract.ABI, 0),
		},
	}

	for _, v := range systemABIs {
		c.Info.Abis = append(c.Info.Abis, &contract.ABI{
			Name:     v.name,
			Args:     v.args,
			Payment:  0,
			GasPrice: int64(1000),
			Limit:    contract.NewCost(100, 100, 100),
		})
	}

	sort.Sort(abiSlice(c.Info.Abis))

	return c
}

func BonusABI() *contract.Contract {
	c := &contract.Contract{
		ID:   "iost.bonus",
		Code: "codes",
		Info: &contract.Info{
			Lang:        "native",
			VersionCode: "1.0.0",
			Abis:        make([]*contract.ABI, 0),
		},
	}

	for _, v := range bonusABIs {
		c.Info.Abis = append(c.Info.Abis, &contract.ABI{
			Name:     v.name,
			Args:     v.args,
			Payment:  0,
			GasPrice: int64(1000),
			Limit:    contract.NewCost(100, 100, 100),
		})
	}

	sort.Sort(abiSlice(c.Info.Abis))

	return c
}

type abiSlice []*contract.ABI

func (s abiSlice) Len() int {
	return len(s)
}
func (s abiSlice) Less(i, j int) bool {
	return s[i].Name < s[j].Name
}
func (s abiSlice) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}
