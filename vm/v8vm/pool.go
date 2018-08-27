package v8

import (
	"github.com/iost-official/Go-IOS-Protocol/core/contract"
	"github.com/iost-official/Go-IOS-Protocol/vm/host"
)

// VMPool manage all V8VM instance.
type VMPool struct {
	size     int
	poolBuff chan *VM
	jsPath   string
}

// NewVMPool create new VMPool instance.
func NewVMPool(size int) *VMPool {
	return &VMPool{
		size:     size,
		poolBuff: make(chan *VM, size),
	}
}

func (vmp *VMPool) getVM() *VM {
	return <-vmp.poolBuff
}

// Init init VMPool.
func (vmp *VMPool) Init() error {
	// Fill vmPoolBuffer
	for i := 0; i < vmp.size; i++ {
		var e = NewVMWithChannel(vmp.poolBuff)
		vmp.poolBuff <- e
	}
	return nil
}

// SetJSPath set standard Javascript library path.
func (vmp *VMPool) SetJSPath(path string) {
	vmp.jsPath = path
}

// Compile compile js code to binary.
func (vmp *VMPool) Compile(contract *contract.Contract) (string, error) {
	vm := vmp.getVM()
	defer vm.recycle()

	return vm.compile(contract)
}

// LoadAndCall load compiled Javascript code and run code with specified api and args
func (vmp *VMPool) LoadAndCall(host *host.Host, contract *contract.Contract, api string, args ...interface{}) (rtn []interface{}, cost *contract.Cost, err error) {
	vm := vmp.getVM()
	defer vm.recycle()

	vm.setJSPath(vmp.jsPath)

	vm.setHost(host)
	preparedCode, _ := vm.setContract(contract, api, args)

	return vm.execute(preparedCode)
}

// Release release all V8VM instance in VMPool
func (vmp *VMPool) Release() {
	for {
		select {
		case e := <-vmp.poolBuff:
			e.release()
		default:
			break
		}
	}
}
