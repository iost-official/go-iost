package database

const (
	// IOSTPrefix ...
	IOSTPrefix = "i-"
)

// BalanceHandler ...
type BalanceHandler struct {
	db database
}

func (m *BalanceHandler) BalanceKey(to string) string {
	return IOSTPrefix + to + "-b"
}

// SetBalance ...
func (m *BalanceHandler) SetBalance(to string, delta int64) {
	ib := m.Balance(to)
	nb := ib + delta
	m.db.Put(m.BalanceKey(to), MustMarshal(nb))
}

// Balance ...
func (m *BalanceHandler) Balance(name string) int64 {
	currentRaw := m.db.Get(m.BalanceKey(name))
	balance := Unmarshal(currentRaw)
	ib, ok := balance.(int64)
	if !ok {
		return 0
	}
	return ib
}

func (m *BalanceHandler) ServiKey(to string) string {
	return IOSTPrefix + to + "-s"
}

// SetServi add delta to servi of to
func (m *BalanceHandler) SetServi(to string, delta int64) {
	ib := m.Balance(to)
	nb := ib + delta
	m.db.Put(m.ServiKey(to), MustMarshal(nb))
}

// Servi get service of name, return 0 if not exists
func (m *BalanceHandler) Servi(name string) int64 {
	currentRaw := m.db.Get(m.ServiKey(name))
	balance := Unmarshal(currentRaw)
	ib, ok := balance.(int64)
	if !ok {
		return 0
	}
	return ib
}
