package new_vm

import (
	"context"

	"github.com/iost-official/Go-IOS-Protocol/core/contract"
)

//go:generate mockgen -destination vm_mock.go -package new_vm github.com/iost-official/Go-IOS-Protocol/new_vm VM

type VM interface {
	Init(api *Host) error
	LoadAndCall(ctx context.Context, contract *contract.Contract, api string, args ...string) (rtn []string, cost *contract.Cost, err error)
	Release()
}
