package native

import (
	"sort"

	"github.com/iost-official/go-iost/core/contract"
)

// SystemABI generate iost.system abi and contract
func SystemABI() *contract.Contract {
	return ABI("iost.system", systemABIs)
}

// GasABI generate iost.gas abi and contract
func GasABI() *contract.Contract {
	return ABI("iost.gas", gasABIs)
}

// TokenABI generate iost.token abi and contract
func TokenABI() *contract.Contract {
	return ABI("iost.token", tokenABIs)
}

// ABI generate native abis
func ABI(id string, abi map[string]*abi) *contract.Contract {
	c := &contract.Contract{
		ID:   id,
		Code: "codes",
		Info: &contract.Info{
			Lang:    "native",
			Version: "1.0.0",
			Abi:     make([]*contract.ABI, 0),
		},
	}

	for _, v := range abi {
		c.Info.Abi = append(c.Info.Abi, &contract.ABI{
			Name: v.name,
			Args: v.args,
		})
	}

	sort.Sort(abiSlice(c.Info.Abi))

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
