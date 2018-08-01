package new_vm

import "context"

type Host struct {
	ctx context.Context
}

func (h *Host) LoadContext(ctx context.Context) *Host {
	return &Host{
		ctx: ctx,
	}
}

func (h *Host) Put(key, value string) error {

}

func (h *Host) Get(key string) (string, error) {
}

func (h *Host) MPut(key, field, value string) error {

}

func (h *Host) SetContract(contract *Contract) error {
	return nil
}
