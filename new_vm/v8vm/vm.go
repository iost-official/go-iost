package v8

/*
#include <stdlib.h>
#include "v8/vm.h"
#cgo LDFLAGS: -lvm -lv8
*/
import "C"
import (
	"github.com/iost-official/Go-IOS-Protocol/core/contract"
	"github.com/iost-official/Go-IOS-Protocol/new_vm/host"
)

func init() {
	C.init()
}

// Engine contains isolate instance, which is a v8 VM with its own heap.
type VM struct {
	isolate              C.IsolatePtr
	sandbox              *Sandbox
	limitsOfInstructions int64
	limitsOfMemorySize   int64
}

func NewVM() *VM {
	isolate := C.newIsolate()
	e := &VM{
		isolate: isolate,
	}
	e.sandbox = NewSandbox(e)
	return e
}

func (e *VM) init() error {
	return nil
}

func (e *VM) Run(code, api string, args ...interface{}) (interface{}, error) {
	contr := &contract.Contract{
		ID:   "run_id",
		Code: code,
	}

	preparedCode, err := e.sandbox.Prepare(contr, api, args)
	if err != nil {
		return "", err
	}

	rs, _, err := e.sandbox.Execute(preparedCode)
	return rs, err
}

func (e *VM) compile(contract *contract.Contract) (string, error) {
	return contract.Code, nil
}

func (e *VM) setHost(host *host.Host) {
	e.sandbox.SetHost(host)
}

func (e *VM) setContract(contract *contract.Contract, api string, args ...interface{}) (string, error) {
	return e.sandbox.Prepare(contract, api, args)
}

func (e *VM) execute(code string) (rtn []interface{}, cost *contract.Cost, err error) {
	rs, gasUsed, err := e.sandbox.Execute(code)
	gasCost := contract.NewCost(gasUsed, 0, 0)
	return []interface{}{rs}, gasCost, err
}

func (e *VM) setJSPath(path string) {
	e.sandbox.SetJSPath(path)
}

// Release release all engine associate resource
func (e *VM) release() {
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
