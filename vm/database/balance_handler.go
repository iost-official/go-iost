package database

const (
	// IOSTPrefix ...
	IOSTPrefix = "i-"
)

// BalanceHandler ...
type BalanceHandler struct {
	db database
}

func (m *BalanceHandler) balanceKey(to string) string {
	return IOSTPrefix + to + "-b"
}

// SetBalance ...
func (m *BalanceHandler) SetBalance(to string, delta int64) {
	ib := m.Balance(to)
	nb := ib + delta
	m.db.Put(m.balanceKey(to), MustMarshal(nb))
}

// Balance ...
func (m *BalanceHandler) Balance(name string) int64 {
	currentRaw := m.db.Get(m.balanceKey(name))
	balance := Unmarshal(currentRaw)
	ib, ok := balance.(int64)
	if !ok {
		return 0
	}
	return ib
}

func (m *BalanceHandler) serviKey(to string) string {
	return IOSTPrefix + to + "-s"
}

// SetServi add delta to servi of to
func (m *BalanceHandler) SetServi(to string, delta int64) {
	ib := m.Servi(to)
	nb := ib + delta
	m.db.Put(m.serviKey(to), MustMarshal(nb))

	// add delta to total servi
	ib = m.Servi("total")
	nb = ib + delta
	m.db.Put(m.serviKey("total"), MustMarshal(nb))
}

// Servi get servi of name, return 0 if not exists
func (m *BalanceHandler) Servi(name string) int64 {
	currentRaw := m.db.Get(m.serviKey(name))
	balance := Unmarshal(currentRaw)
	ib, ok := balance.(int64)
	if !ok {
		return 0
	}
	return ib
}

// TotalServi get total servi of name, return 0 if not exists
func (m *BalanceHandler) TotalServi() int64 {
	return m.Servi("total")
}
