package database

import (
	"github.com/iost-official/Go-IOS-Protocol/core/contract"
)

// ContractPrefix ...
const ContractPrefix = "c-"

// ContractHandler ...
type ContractHandler struct {
	db database
}

// SetContract ...
func (m *ContractHandler) SetContract(contract *contract.Contract) {
	if contract != nil {
		m.db.Put(ContractPrefix+contract.ID, contract.Encode())
	} else {
		panic("set a nil contract")
	}
}

// Contract ...
func (m *ContractHandler) Contract(key string) (c *contract.Contract) {
	str := m.db.Get(ContractPrefix + key)
	c = &contract.Contract{}
	err := c.Decode(str)
	if err != nil {
		return nil
	}
	return
}

// HasContract ...
func (m *ContractHandler) HasContract(key string) bool {
	return m.db.Has(ContractPrefix + key)
}

// DelContract ...
func (m *ContractHandler) DelContract(key string) {
	m.db.Del(ContractPrefix + key)
}
