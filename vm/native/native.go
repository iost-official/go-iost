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

// Impl .
type Impl struct {
}

func register(m map[string]*abi, a *abi) {
	m[a.name] = a
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
		a, ok = systemABIs[api]
	case "domain.iost":
		a, ok = DomainABIs[api]
	case "gas.iost":
		a, ok = gasABIs[api]
		if !ok {
			a, ok = gasInnerABIs[api]
		}
	case "token.iost":
		a, ok = tokenABIs[api]
		if !ok {
			a, ok = tokenInnerABIs[api]
		}
	case "token721.iost":
		a, ok = token721ABIs[api]
		if !ok {
			a, ok = token721InnerABIs[api]
		}
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
