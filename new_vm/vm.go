package new_vm

import "context"

type VM interface {
	Init() error
	Load(contract *Contract) error
	Call(ctx context.Context, api string, args ...string)(rtn []string, err error)
	Release()
}
