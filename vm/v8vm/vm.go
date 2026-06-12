package v8

import (
	"math/rand"

	"github.com/iost-official/go-iost/v3/core/contract"
	"github.com/iost-official/go-iost/v3/vm/host"
)

const vmRefLimit = 60

// VM contains a goja runtime and sandbox.
type VM struct {
	sandbox              *Sandbox
	releaseChannel       chan *VM
	vmType               vmPoolType
	jsPath               string
	refCount             int
	limitsOfInstructions int64 // nolint
	limitsOfMemorySize   int64 // nolint
}

// NewVM return new vm with sandbox
func NewVM(poolType vmPoolType, jsPath string) *VM {
	e := &VM{
		vmType: poolType,
		jsPath: jsPath,
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
		// release runtime completely
		e.refCount = 0
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
}
