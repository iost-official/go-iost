package native

import (
	"errors"

	"github.com/iost-official/Go-IOS-Protocol/core/contract"
	"github.com/iost-official/Go-IOS-Protocol/ilog"
	"github.com/iost-official/Go-IOS-Protocol/vm/host"
)

type abi struct {
	name string
	args []string
	do   func(h *host.Host, args ...interface{}) (rtn []interface{}, cost *contract.Cost, err error)
}

// Impl .
type Impl struct {
}

func register(m *map[string]*abi, a *abi) {
	(*m)[a.name] = a
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

func (i *Impl) LoadAndCall(h *host.Host, con *contract.Contract, api string, args ...interface{}) (rtn []interface{}, cost *contract.Cost, err error) {
	var (
		a  *abi
		ok bool
	)

	switch con.ID {
	case "iost.system":
		a, ok = systemABIs[api]
	case "iost.domain":
		a, ok = domainABIs[api]
	}
	if !ok {
		ilog.Fatal("error", con.ID, api, systemABIs)
		return nil, host.CommonErrorCost(1), errors.New("unknown api name")
	}

	return a.do(h, args...)
}
