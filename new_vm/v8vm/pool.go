package v8

import (
	"github.com/iost-official/Go-IOS-Protocol/core/contract"
	"github.com/iost-official/Go-IOS-Protocol/new_vm/host"
)

type VMPool struct {
	size   int
	jsPath string
}

func NewVMPool(size int) *VMPool {
	return &VMPool{
		size: size,
	}
}

func (vmp *VMPool) getVM() *VM {
	return NewVM()
}

func (vmp *VMPool) Init() error {
	return nil
}

func (vmp *VMPool) SetJSPath(path string) {
	vmp.jsPath = path
}

func (vmp *VMPool) Compile(contract *contract.Contract) (string, error) {
	vm := vmp.getVM()
	defer vm.release()

	return vm.compile(contract)
}

func (vmp *VMPool) LoadAndCall(host *host.Host, contract *contract.Contract, api string, args ...interface{}) (rtn []interface{}, cost *contract.Cost, err error) {
	vm := vmp.getVM()
	defer vm.release()

	vm.setJSPath(vmp.jsPath)

	vm.setHost(host)
	preparedCode, _ := vm.setContract(contract, api, args)

	return vm.execute(preparedCode)
}

func (vmp *VMPool) Release() {
}
