package native_vm

import (
	"context"

	"github.com/iost-official/Go-IOS-Protocol/core/contract"
)

type VM struct {
}

func (m *VM) Init() error {
	return nil
}
func (m *VM) LoadAndCall(ctx context.Context, contract *contract.Contract, api string, args ...string) (rtn []string, err error) {

	return nil, nil
}
func (m *VM) Release() {

}

func setContract(ctx context.Context, contract *contract.Contract) error {
	return nil
}
