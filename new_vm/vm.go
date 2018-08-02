package new_vm

import (
	"context"

	"github.com/iost-official/Go-IOS-Protocol/core/contract"
)

type VM interface {
	Init(api *Host) error
	LoadAndCall(ctx context.Context, contract *contract.Contract, api string, args ...string) (rtn []string, cost *contract.Cost, err error)
	Release()
}
