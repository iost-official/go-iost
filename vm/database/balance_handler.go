package database

const (
	// IOSTPrefix ...
	IOSTPrefix = "i-"
)

// BalanceHandler ...
type BalanceHandler struct {
	db database
}

// SetBalance ...
func (m *BalanceHandler) SetBalance(to string, delta int64) {
	ib := m.Balance(to)
	//nb := ib + delta
	println("nb:", ib, to)
	//m.db.Put(IOSTPrefix+to, MustMarshal(nb))
}

// Balance ...
func (m *BalanceHandler) Balance(name string) int64 {
	currentRaw := m.db.Get(IOSTPrefix + name)
	balance := Unmarshal(currentRaw)
	ib, ok := balance.(int64)
	if !ok {
		return 0
	}
	return ib
}
