package v8

import (
	"github.com/iost-official/Go-IOS-Protocol/core/contract"
	"github.com/iost-official/Go-IOS-Protocol/vm/host"
)

// VMPool provides interface to execute JS smart contract
type VMPool struct {
	size   int
	jsPath string
}

// NewVMPool by size
func NewVMPool(size int) *VMPool {
	return &VMPool{
		size: size,
	}
}

func (vmp *VMPool) getVM() *VM {
	return NewVM()
}

// Init VMPool
func (vmp *VMPool) Init() error {
	return nil
}

// SetJSPath set path including js library
func (vmp *VMPool) SetJSPath(path string) {
	vmp.jsPath = path
}

// Compile contract before storage
func (vmp *VMPool) Compile(contract *contract.Contract) (string, error) {
	vm := vmp.getVM()
	defer vm.release()

	return vm.compile(contract)
}

// LoadAndCall load contract and call api function, return results and cost
func (vmp *VMPool) LoadAndCall(host *host.Host, contract *contract.Contract, api string, args ...interface{}) (rtn []interface{}, cost *contract.Cost, err error) {
	vm := vmp.getVM()
	defer vm.release()

	vm.setJSPath(vmp.jsPath)

	vm.setHost(host)
	preparedCode, _ := vm.setContract(contract, api, args)

	return vm.execute(preparedCode)
}

// Release invoke release when VMPool no longer in use
func (vmp *VMPool) Release() {
}
