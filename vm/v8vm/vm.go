package v8

/*
#include <stdlib.h>
#include "v8/vm.h"
#cgo darwin LDFLAGS: -L${SRCDIR}/v8/libv8/_darwin_amd64 -lvm
#cgo linux LDFLAGS: -L${SRCDIR}/v8/libv8/_linux_amd64 -lvm -lv8 -Wl,-rpath,${SRCDIR}/v8/libv8/_linux_amd64
*/
import "C"
import (
	"sync"

	"math/rand"

	"github.com/iost-official/go-iost/v3/core/contract"
	"github.com/iost-official/go-iost/v3/vm/host"
)

const vmRefLimit = 60

// CVMInitOnce vm init once
var CVMInitOnce = sync.Once{}
var customStartupData C.CustomStartupData
var customCompileStartupData C.CustomStartupData

// VM contains isolate instance, which is a v8 VM with its own heap.
type VM struct {
	isolate              C.IsolateWrapperPtr
	sandbox              *Sandbox
	releaseChannel       chan *VM
	vmType               vmPoolType
	jsPath               string
	refCount             int
	limitsOfInstructions int64 // nolint
	limitsOfMemorySize   int64 // nolint
}

// NewVM return new vm with isolate and sandbox
func NewVM(poolType vmPoolType, jsPath string) *VM {
	CVMInitOnce.Do(func() {
		C.init()
		customStartupData = C.createStartupData()
		customCompileStartupData = C.createCompileStartupData()
	})
	var isolateWrapperPtr C.IsolateWrapperPtr
	if poolType == CompileVMPool {
		isolateWrapperPtr = C.newIsolate(customCompileStartupData)
	} else {
		isolateWrapperPtr = C.newIsolate(customStartupData)
	}
	e := &VM{
		isolate: isolateWrapperPtr,
		vmType:  poolType,
		jsPath:  jsPath,
	}
	e.sandbox = NewSandbox(e, 1)

	return e
}

// NewVMWithChannel return new vm with release channel
func NewVMWithChannel(vmType vmPoolType, jsPath string, releaseChannel chan *VM) *VM {
	e := NewVM(vmType, jsPath)
	e.releaseChannel = releaseChannel
	return e
}

func (e *VM) validate(c *contract.Contract) error {
	return e.sandbox.Validate(c)
}

func (e *VM) compile(contract *contract.Contract) (string, error) {
	return e.sandbox.Compile(contract)
}

func (e *VM) setHost(host *host.Host) {
	e.sandbox.SetHost(host)
}

func (e *VM) setContract(contract *contract.Contract, api string, args []any) (string, error) {
	return e.sandbox.Prepare(contract, api, args)
}

func (e *VM) execute(code string) (rtn []any, cost contract.Cost, err error) {
	rs, gasUsed, err := e.sandbox.Execute(code)
	gasCost := contract.NewCost(0, 0, gasUsed)
	return []any{rs}, gasCost, err
}

// EnsureFlags make sandbox flag is same with input
func (e *VM) EnsureFlags(flags int64) {
	if e.sandbox.GetFlags() != flags {
		if e.sandbox != nil {
			e.sandbox.Release()
		}
		e.sandbox = NewSandbox(e, flags)
	}
}

func (e *VM) recycle(poolType vmPoolType) {
	// first release sandbox
	if e.sandbox != nil {
		e.sandbox.Release()
	}

	if rand.Int()%(vmRefLimit-e.refCount) == 0 {
		// release isolate
		if e.isolate != nil {
			e.refCount = 0
			C.releaseIsolate(e.isolate)
		}
		// regen isolate
		if poolType == CompileVMPool {
			e.isolate = C.newIsolate(customCompileStartupData)
		} else {
			e.isolate = C.newIsolate(customStartupData)
		}
	} else {
		C.lowMemoryNotification(e.isolate)
	}

	// then regen new sandbox
	e.sandbox = NewSandbox(e, e.sandbox.GetFlags())
	if e.releaseChannel != nil {
		e.releaseChannel <- e
	}
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
