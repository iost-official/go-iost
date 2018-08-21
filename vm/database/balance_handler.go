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
	m.db.Put(IOSTPrefix+to, MustMarshal(nb))
}

func (m *BalanceHandler) Balance(name string) int64 {
	currentRaw := m.db.Get(IOSTPrefix + name)
	balance := Unmarshal(currentRaw)
	ib, ok := balance.(int64)
	if !ok {
		return 0
	}
	return ib
}
