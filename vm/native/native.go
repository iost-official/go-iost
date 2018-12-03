package native

import (
	"errors"

	"github.com/iost-official/go-iost/core/contract"
	"github.com/iost-official/go-iost/ilog"
	"github.com/iost-official/go-iost/vm/host"
)

type abi struct {
	name string
	args []string
	do   func(h *host.Host, args ...interface{}) (rtn []interface{}, cost contract.Cost, err error)
}

type abiSet struct {
	abi        map[string]*abi
	privateMap map[string]bool
}

func newAbiSet() *abiSet {
	return &abiSet{
		abi:        make(map[string]*abi),
		privateMap: make(map[string]bool),
	}
}

// Register is register an abi to abiSet
func (as *abiSet) Register(a *abi, isPrivate ...bool) {
	as.abi[a.name] = a
	if len(isPrivate) > 0 && isPrivate[0] {
		as.privateMap[a.name] = true
	}
}

// Get is get abi from an abiSet
func (as *abiSet) Get(name string) (a *abi, ok bool) {
	a, ok = as.abi[name]
	return
}

// PublicAbis is get public abis from an abiSet
func (as *abiSet) PublicAbis() map[string]*abi {
	abis := make(map[string]*abi)
	for name, a := range as.abi {
		if _, ok := as.privateMap[name]; !ok {
			abis[name] = a
		}
	}
	return abis
}

// Impl .
type Impl struct {
}

// Init .
func (i *Impl) Init() error {
	return nil
}

// Release .
func (i *Impl) Release() {

}

// Compile ...
func (i *Impl) Compile(contract *contract.Contract) (string, error) {
	return "", nil
}

// LoadAndCall implement
func (i *Impl) LoadAndCall(h *host.Host, con *contract.Contract, api string, args ...interface{}) (rtn []interface{}, cost contract.Cost, err error) {
	var (
		a  *abi
		ok bool
	)

	switch con.ID {
	case "system.iost":
		a, ok = systemABIs.Get(api)
	case "domain.iost":
		a, ok = DomainABIs.Get(api)
	case "gas.iost":
		a, ok = gasABIs.Get(api)
	case "token.iost":
		a, ok = tokenABIs.Get(api)
	case "token721.iost":
		a, ok = token721ABIs.Get(api)
	}
	if !ok {
		ilog.Fatal("error", con.ID, api, systemABIs)
		return nil, host.CommonErrorCost(1), errors.New("unknown api name")
	}

	return a.do(h, args...)
}

// CheckCost check if cost exceed gas_limit
func CheckCost(h *host.Host, cost contract.Cost) bool {
	gasLimit := h.Context().GValue("gas_limit").(int64)
	if cost.ToGas() > gasLimit {
		return false
	}
	return true
}
