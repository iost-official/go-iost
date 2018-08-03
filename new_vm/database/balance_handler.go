package database

const (
	IOSTPrefix = "i-"
)

type BalanceHandler struct {
	db Database
}

func (m *BalanceHandler) SetBalance(to string, delta int64) {
	ib := m.Balance(to)
	nb := ib + delta
	m.db.Put(IOSTPrefix+to, MustMarshall(nb))
}

func (m *BalanceHandler) Balance(name string) int64 {
	currentRaw := m.db.Get(IOSTPrefix + name)
	balance := Unmarshall(currentRaw)
	ib, ok := balance.(int64)
	if !ok {
		panic(balance)
	}
	return ib
}
