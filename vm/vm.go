package vm

import (
	"github.com/iost-official/go-iost/v3/core/contract"
	"github.com/iost-official/go-iost/v3/vm/host"
)

//go:generate mockgen -destination vm_mock.go -package vm github.com/iost-official/go-iost/v3/vm VM

// VM ...
type VM interface {
	Init() error
	Validate(contract *contract.Contract) error
	Compile(contract *contract.Contract) (string, error)
	LoadAndCall(host *host.Host, contract *contract.Contract, api string, args ...any) (rtn []any, cost contract.Cost, err error)
	Release()
}
