package v8

/*
#include <stdlib.h>
#include "v8/vm.h"
#cgo LDFLAGS: -L./v8/libv8 -lvm -lv8 -Wl,-rpath=./v8/libv8
*/
import "C"
import (
	"context"

	"github.com/iost-official/Go-IOS-Protocol/new_vm"
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

// LoadAndCall load contract code with provide contract, and call api with args
func (e *VM) LoadAndCall(ctx context.Context, contract *new_vm.Contract, api string, args ...string) (rtn []string, err error) {
	code := contract.Code

	preparedCode := e.sandbox.Prepare(code, api, args)

	rs, err := e.sandbox.Execute(preparedCode)

	return []string{rs}, err
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
