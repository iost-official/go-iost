package native

import (
	"context"

	vm "github.com/iost-official/Go-IOS-Protocol/new_vm"
)

type VM struct {
}

func (m *VM) Init() error {
	return nil
}
func (m *VM) LoadAndCall(ctx context.Context, contract *vm.Contract, api string, args ...string) (rtn []string, err error) {
	return nil, nil
}
func (m *VM) Release() {

}
