package database

import "github.com/iost-official/Go-IOS-Protocol/core/contract"

const ContractPrefix = "c-"

type ContractHandler struct {
	db Database
}

func (m *ContractHandler) SetContract(contract *contract.Contract) {
	if contract != nil {
		m.db.Put(ContractPrefix+contract.Name, contract.Encode())
	} else {
		panic("set a nil contract")
	}
}

func (m *ContractHandler) Contract(key string) (contract *contract.Contract) {
	str := m.db.Get(ContractPrefix + key)
	contract.Decode(str)
	return
}

func (m *ContractHandler) HasContract(key string) bool {
	return m.db.Has(ContractPrefix + key)
}

func (m *ContractHandler) DelContract(key string) {
	m.db.Del(ContractPrefix + key)
}
