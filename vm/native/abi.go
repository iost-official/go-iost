package native

import (
	"sort"

	"github.com/iost-official/go-iost/core/contract"
)

// SystemABI generate system.iost abi and contract
func SystemABI() *contract.Contract {
	return ABI("system.iost", systemABIs)
}

// GasABI generate gas.iost abi and contract
func GasABI() *contract.Contract {
	return ABI("gas.iost", gasABIs)
}

// TokenABI generate token.iost abi and contract
func TokenABI() *contract.Contract {
	return ABI("token.iost", tokenABIs)
}

// Token721ABI generate token.iost abi and contract
func Token721ABI() *contract.Contract {
	return ABI("token721.iost", token721ABIs)
}

// ABI generate native abis
func ABI(id string, abiSet *abiSet) *contract.Contract {
	c := &contract.Contract{
		ID:   id,
		Code: "codes",
		Info: &contract.Info{
			Lang:    "native",
			Version: "1.0.0",
			Abi:     make([]*contract.ABI, 0),
		},
	}

	for _, v := range abiSet.PublicAbis() {
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
