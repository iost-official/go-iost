package database

import (
	"fmt"
)

const (
	// IOSTPrefix prefix of iost
	IOSTPrefix = "i-"
)

// BalanceHandler handler of balace storage
type BalanceHandler struct {
	db database
}

// GetInt64 return the value as int64. If no key exists, return 0
func (m *BalanceHandler) GetInt64(key string) (value int64) {
	value, _ = Unmarshal(m.db.Get(key)).(int64)
	return
}

// GetFloat64 return the value as float64. If no key exists, return 0
func (m *BalanceHandler) GetFloat64(key string) (value float64) {
	value, _ = Unmarshal(m.db.Get(key)).(float64)
	return
}

// PutInt64 ...
func (m *BalanceHandler) PutInt64(key string, value int64) {
	m.db.Put(key, MustMarshal(value))
}

// PutFloat64 ...
func (m *BalanceHandler) PutFloat64(key string, value float64) {
	m.db.Put(key, MustMarshal(value))
}

func (m *BalanceHandler) balanceKey(to string) string {
	return IOSTPrefix + to + "-b"
}

// SetBalance set balance to id
func (m *BalanceHandler) SetBalance(to string, delta int64) {
	ib := m.Balance(to)
	nb := ib + delta
	m.db.Put(m.balanceKey(to), MustMarshal(nb))
}

// Balance get balance to id
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

const (
	gasRateSuffix       = "-gr"
	gasLimitSuffix      = "-gl"
	gasUpdateTimeSuffix = "-gt"
	gasStockSuffix      = "-gs"
	gasPledgeSuffix     = "-gp"
)

func (m *BalanceHandler) gasRateKey(name string) string {
	return IOSTPrefix + name + gasRateSuffix
}

// GetGasRate ...
func (m *BalanceHandler) GetGasRate(name string) int64 {
	return m.GetInt64(m.gasRateKey(name))
}

// SetGasRate ...
func (m *BalanceHandler) SetGasRate(name string, r int64) {
	m.PutInt64(m.gasRateKey(name), r)
}

func (m *BalanceHandler) gasLimitKey(name string) string {
	return IOSTPrefix + name + gasLimitSuffix
}

// GetGasLimit ...
func (m *BalanceHandler) GetGasLimit(name string) int64 {
	return m.GetInt64(m.gasLimitKey(name))
}

// SetGasLimit ...
func (m *BalanceHandler) SetGasLimit(name string, l int64) {
	m.PutInt64(m.gasLimitKey(name), l)
}

func (m *BalanceHandler) gasUpdateTimeKey(name string) string {
	return IOSTPrefix + name + gasUpdateTimeSuffix
}

// GetGasUpdateTime ...
func (m *BalanceHandler) GetGasUpdateTime(name string) int64 {
	return m.GetInt64(m.gasUpdateTimeKey(name))
}

// SetGasUpdateTime ...
func (m *BalanceHandler) SetGasUpdateTime(name string, t int64) {
	fmt.Printf("set update time %d\n", t)
	m.PutInt64(m.gasUpdateTimeKey(name), t)
}

func (m *BalanceHandler) gasStockKey(name string) string {
	return IOSTPrefix + name + gasStockSuffix
}

// GetGasStock ...
func (m *BalanceHandler) GetGasStock(name string) int64 {
	return m.GetInt64(m.gasStockKey(name))
}

// SetGasStock ...
func (m *BalanceHandler) SetGasStock(name string, g int64) {
	fmt.Printf("set stock %d\n", g)
	m.PutInt64(m.gasStockKey(name), g)
}

func (m *BalanceHandler) gasPledgeKey(name string) string {
	return IOSTPrefix + name + gasPledgeSuffix
}

// GetGasPledge ...
func (m *BalanceHandler) GetGasPledge(name string) int64 {
	return m.GetInt64(m.gasPledgeKey(name))
}

// SetGasPledge ...
func (m *BalanceHandler) SetGasPledge(name string, p int64) {
	m.PutInt64(m.gasPledgeKey(name), p)
}

// CurrentTotalGas return current total gas
func (m *BalanceHandler) CurrentTotalGas(name string, now int64) (result int64) {
	result = m.GetGasStock(name)
	fmt.Printf("stock %d\n", result)
	gasUpdateTime := m.GetGasUpdateTime(name)
	var timeDuration int64
	if gasUpdateTime > 0 {
		timeDuration = now - gasUpdateTime
	}
	rate := m.GetGasRate(name)
	limit := m.GetGasLimit(name)
	result += timeDuration * rate
	fmt.Printf("result limit %d %d \n", result, limit)
	if result > limit {
		result = limit
	}
	return
}
