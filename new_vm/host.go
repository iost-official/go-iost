package new_vm

type Host struct {
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
