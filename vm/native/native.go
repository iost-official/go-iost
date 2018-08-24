package native

import (
	"errors"

	"github.com/iost-official/Go-IOS-Protocol/core/contract"
	"github.com/iost-official/Go-IOS-Protocol/vm/host"
)

type abi struct {
	name string
	args []string
	do   func(h *host.Host, args ...interface{}) (rtn []interface{}, cost *contract.Cost, err error)
}

var abis map[string]*abi

// Impl .
type Impl struct {
}

func register(a *abi) {
	abis[a.name] = a
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
	abi, ok := abis[api]
	if !ok {
		return nil, host.CommonErrorCost(1), errors.New("unknown api name")
	}
	return abi.do(h, args...)
}

func init() {
	abis = make(map[string]*abi)
	register(requireAuth)
	register(receipt)
	register(callWithReceipt)
	register(transfer)
	register(topUp)
	register(countermand)
	register(setCode)
	register(updateCode)
	register(destroyCode)
	register(issueIOST)
}
