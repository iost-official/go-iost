package new_vm

import (
	"github.com/iost-official/Go-IOS-Protocol/core/contract"
)

//go:generate mockgen -destination vm_mock.go -package new_vm github.com/iost-official/Go-IOS-Protocol/new_vm VM

type VM interface {
	Init() error
	LoadAndCall(host *Host, contract *contract.Contract, api string, args ...interface{}) (rtn []string, cost *contract.Cost, err error)
	Release()
}
