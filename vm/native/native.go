package native

import (
	"errors"

	"github.com/iost-official/Go-IOS-Protocol/core/contract"
	"github.com/iost-official/Go-IOS-Protocol/vm/host"
)

type abi struct {
	name string
	do   func(h *host.Host, args ...interface{}) (rtn []interface{}, cost *contract.Cost, err error)
}

// Impl .
type Impl struct {
	abis map[string]func(h *host.Host, args ...interface{}) (rtn []interface{}, cost *contract.Cost, err error)
}

func (i *Impl) register(a *abi) {
	i.abis[a.name] = a.do
}

// Init .
func (i *Impl) Init() error {
	i.abis = make(map[string]func(h *host.Host, args ...interface{}) (rtn []interface{}, cost *contract.Cost, err error))
	i.register(requireAuth)
	i.register(receipt)
	i.register(callWithReceipt)
	i.register(transfer)
	i.register(topUp)
	i.register(countermand)
	i.register(setCode)
	i.register(updateCode)
	i.register(destroyCode)
	i.register(issueIOST)
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
	doer, ok := i.abis[api]
	if !ok {
		return nil, host.CommonErrorCost(1), errors.New("unknown api name")
	}
	return doer(h, args...)
}
