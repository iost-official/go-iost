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

	"github.com/iost-official/go-iost/core/contract"
	"github.com/iost-official/go-iost/vm/host"
)

const vmRefLimit = 30

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
	limitsOfInstructions int64
	limitsOfMemorySize   int64
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
	e.sandbox = NewSandbox(e)

	return e
}

// NewVMWithChannel return new vm with release channel
func NewVMWithChannel(vmType vmPoolType, jsPath string, releaseChannel chan *VM) *VM {
	e := NewVM(vmType, jsPath)
	e.releaseChannel = releaseChannel
	return e
}

func (e *VM) init() error {
	return nil
}

func (e *VM) compile(contract *contract.Contract) (string, error) {
	return e.sandbox.Compile(contract)
}

func (e *VM) setHost(host *host.Host) {
	e.sandbox.SetHost(host)
}

func (e *VM) setContract(contract *contract.Contract, api string, args []interface{}) (string, error) {
	return e.sandbox.Prepare(contract, api, args)
}

func (e *VM) execute(code string) (rtn []interface{}, cost contract.Cost, err error) {
	rs, gasUsed, err := e.sandbox.Execute(code)
	gasCost := contract.NewCost(0, 0, gasUsed)
	return []interface{}{rs}, gasCost, err
}

func (e *VM) setJSPath(path string) {
	e.sandbox.SetJSPath(path, e.vmType)
}

func (e *VM) setReleaseChannel(releaseChannel chan *VM) {
	e.releaseChannel = releaseChannel
}

func (e *VM) recycle(poolType vmPoolType) {
	// first release sandbox
	if e.sandbox != nil {
		e.sandbox.Release()
	}

	if e.refCount >= vmRefLimit {
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
	}

	// then regen new sandbox
	e.sandbox = NewSandbox(e)
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
