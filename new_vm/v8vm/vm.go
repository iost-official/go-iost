package v8

/*
#include <stdlib.h>
#include "v8/vm.h"
#cgo LDFLAGS: -lvm
*/
import "C"
import (
	"github.com/iost-official/Go-IOS-Protocol/core/contract"
	"github.com/iost-official/Go-IOS-Protocol/new_vm"
	"fmt"
)

func init() {
	C.init()
}

// Engine contains isolate instance, which is a v8 VM with its own heap.
type VM struct {
	isolate              C.IsolatePtr
	sandbox              *Sandbox
	limitsOfInstructions uint64
	limitsOfMemorySize   uint64
}

func NewVM() *VM {
	isolate := C.newIsolate()
	e := &VM{
		isolate: isolate,
	}
	e.sandbox = NewSandbox(e)
	return e
}

func (e *VM) Init() error {
	return nil
}

func (e *VM) Run(code, api string, args ...interface{}) (string, error) {
	contr := &contract.Contract{
		ID: "run_id",
		Code: code,
	}

	preparedCode, err := e.sandbox.Prepare(contr, api, args)
	if err != nil {
		return "", err
	}

	rs, err := e.sandbox.Execute(preparedCode)
	return rs, err
}

// LoadAndCall load contract code with provide contract, and call api with args
func (e *VM) LoadAndCall(host *new_vm.Host, contract *contract.Contract, api string, args ...interface{}) (rtn []interface{}, cost *contract.Cost, err error) {
	defer func() {
		if err := recover(); err != nil {
			fmt.Println("LoadAndCall recover:", err)
		}
	}()

	e.sandbox.SetHost(host)
	preparedCode, err := e.sandbox.Prepare(contract, api, args)
	if err != nil {

	}

	rs, err := e.sandbox.Execute(preparedCode)

	return []interface{}{rs}, nil, err
}

// Release release all engine associate resource
func (e *VM) Release() {
	// first release sandbox
	if e.sandbox != nil {
		e.sandbox.Release()
	}
	e.sandbox = nil

	// then release isolate
	if e.isolate != nil {
		C.releaseIsolate(e.isolate)
	}
	e.isolate = nil
}
