package database

import vm "github.com/iost-official/Go-IOS-Protocol/new_vm"

const ContractPrefix = "c-"

type ContractHandler struct {
	db Database
}

func (m *ContractHandler) SetContract(contract *vm.Contract) {
	if contract != nil {
		m.db.Put(ContractPrefix+contract.Name, contract.Encode())
	} else {
		panic("set a nil contract")
	}
}

func (m *ContractHandler) GetContract(key string) (contract *vm.Contract) {
	str := m.db.Get(ContractPrefix + key)
	contract = vm.DecodeContract(str)
	return
}

func (m *ContractHandler) HasContract(key string) bool {
	return m.db.Has(ContractPrefix + key)
}

func (m *ContractHandler) DelContract(key string) {
	m.db.Del(ContractPrefix + key)
}
