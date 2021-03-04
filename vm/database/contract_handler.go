package database

import (
	"github.com/iost-official/go-iost/v3/core/contract"
)

// ContractPrefix ...
const ContractPrefix = "c-"

// ContractHandler ...
type ContractHandler struct {
	db database
}

// SetContract set contract to storage, will not do check
func (m *ContractHandler) SetContract(contract *contract.Contract) {
	if contract != nil {
		m.db.Put(ContractPrefix+contract.ID, contract.Encode())
	} else {
		panic("set a nil contract")
	}
}

// Contract get contract by key
func (m *ContractHandler) Contract(key string) (c *contract.Contract) {
	str := m.db.Get(ContractPrefix + key)
	c = &contract.Contract{}
	err := c.Decode(str)
	if err != nil {
		return nil
	}
	return
}

// HasContract determine if contract existed
func (m *ContractHandler) HasContract(key string) bool {
	return m.db.Has(ContractPrefix + key)
}

// DelContract delete contract, if contract not exist, do nothing
func (m *ContractHandler) DelContract(key string) {
	m.db.Del(ContractPrefix + key)
}
