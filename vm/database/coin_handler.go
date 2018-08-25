package database

import "math"

const (
	// CoinPrefix ...
	CoinPrefix = "o-"
)

// CoinHandler ...
type CoinHandler struct {
	db database
}

func (m *CoinHandler) getKey(coinName, name string) string {
	return CoinPrefix + coinName + "-" + name
}

// SetCoin ...
func (m *CoinHandler) SetCoin(coinName string, to string, delta int64) {
	ib := m.Coin(coinName, to)
	var nb int64
	if delta > 0 && math.MaxInt64-delta > ib {
		nb = math.MaxInt64
	} else if delta < 0 && math.MinInt64-delta > ib {
		nb = math.MinInt64
	} else {
		nb = ib + delta
	}
	m.db.Put(m.getKey(coinName, to), MustMarshal(nb))
}

// Coin ...
func (m *CoinHandler) Coin(coinName string, name string) int64 {
	currentRaw := m.db.Get(m.getKey(coinName, name))
	balance := Unmarshal(currentRaw)
	ib, ok := balance.(int64)
	if !ok {
		return 0
	}
	return ib
}
