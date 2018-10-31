package database

import (
	"github.com/iost-official/go-iost/common"
)

const (
	gasRateSuffix       = "-gr"
	gasLimitSuffix      = "-gl"
	gasUpdateTimeSuffix = "-gt"
	gasStockSuffix      = "-gs"
	gasPledgeSuffix     = "-gp"
)

// GasHandler handle gas related storage
type GasHandler struct {
	db database
}

// If no key exists, return 0
func (m *GasHandler) getFN(key string) (value *common.Fixed) {
	value, _ = common.UnmarshalFixed(m.db.Get(key))
	return
}

// putFN ...
func (m *GasHandler) putFN(key string, value *common.Fixed) {
	m.db.Put(key, value.Marshal())
}

// getInt64 return the value as int64. If no key exists, return 0
func (m *GasHandler) getInt64(key string) (value int64) {
	value, _ = Unmarshal(m.db.Get(key)).(int64)
	return
}

// putInt64 ...
func (m *GasHandler) putInt64(key string, value int64) {
	m.db.Put(key, MustMarshal(value))
}

// getFloat64 return the value as float64. If no key exists, return 0
func (m *GasHandler) getFloat64(key string) (value float64) {
	value, _ = Unmarshal(m.db.Get(key)).(float64)
	return
}

// putFloat64 ...
func (m *GasHandler) putFloat64(key string, value float64) {
	m.db.Put(key, MustMarshal(value))
}

func (m *GasHandler) gasRateKey(name string) string {
	return IOSTPrefix + name + gasRateSuffix
}

// GetGasRate ...
func (m *GasHandler) GetGasRate(name string) *common.Fixed {
	return m.getFN(m.gasRateKey(name))
}

// SetGasRate ...
func (m *GasHandler) SetGasRate(name string, r *common.Fixed) {
	m.putFN(m.gasRateKey(name), r)
}

func (m *GasHandler) gasLimitKey(name string) string {
	return IOSTPrefix + name + gasLimitSuffix
}

// GetGasLimit ...
func (m *GasHandler) GetGasLimit(name string) *common.Fixed {
	return m.getFN(m.gasLimitKey(name))
}

// SetGasLimit ...
func (m *GasHandler) SetGasLimit(name string, l *common.Fixed) {
	m.putFN(m.gasLimitKey(name), l)
}

func (m *GasHandler) gasUpdateTimeKey(name string) string {
	return IOSTPrefix + name + gasUpdateTimeSuffix
}

// GetGasUpdateTime ...
func (m *GasHandler) GetGasUpdateTime(name string) int64 {
	return m.getInt64(m.gasUpdateTimeKey(name))
}

// SetGasUpdateTime ...
func (m *GasHandler) SetGasUpdateTime(name string, t int64) {
	m.putInt64(m.gasUpdateTimeKey(name), t)
}

func (m *GasHandler) gasStockKey(name string) string {
	return IOSTPrefix + name + gasStockSuffix
}

// GetGasStock `gasStock` means the gas amount at last update time.
func (m *GasHandler) GetGasStock(name string) *common.Fixed {
	return m.getFN(m.gasStockKey(name))
}

// SetGasStock ...
func (m *GasHandler) SetGasStock(name string, g *common.Fixed) {
	m.putFN(m.gasStockKey(name), g)
}

func (m *GasHandler) gasPledgeKey(name string) string {
	return IOSTPrefix + name + gasPledgeSuffix
}

// GetGasPledge ...
func (m *GasHandler) GetGasPledge(name string) *common.Fixed {
	return m.getFN(m.gasPledgeKey(name))
}

// SetGasPledge ...
func (m *GasHandler) SetGasPledge(name string, p *common.Fixed) {
	m.putFN(m.gasPledgeKey(name), p)
}

// CurrentTotalGas return current total gas. It is min(limit, last_updated_gas + time_since_last_updated * increase_speed)
func (m *GasHandler) CurrentTotalGas(name string, now int64) (result *common.Fixed) {
	result = m.GetGasStock(name)
	gasUpdateTime := m.GetGasUpdateTime(name)
	var timeDuration int64
	if gasUpdateTime > 0 {
		timeDuration = now - gasUpdateTime
	}
	rate := m.GetGasRate(name)
	limit := m.GetGasLimit(name)
	//fmt.Printf("CurrentTotalGas stock %v rate %v limit %v", result, rate, limit)
	result = result.Add(rate.Times(timeDuration))
	if limit.LessThan(result) {
		result = limit
	}
	return
}
