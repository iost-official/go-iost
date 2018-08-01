package new_vm

import "context"

type VM interface {
	Init(api *Host) error
	LoadAndCall(ctx context.Context, contract *Contract, api string, args ...string) (rtn []string, err error)
	Release()
}
