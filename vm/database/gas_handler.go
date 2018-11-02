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

const (
	DecPledge  = 8
	DecGasRate = 8
	DecGas     = 8
)

// GasHandler handle gas related storage
type GasHandler struct {
	db database
}

// If no key exists, return 0
func (m *GasHandler) getFixed(key string) (value *common.Fixed) {
	var err error
	value, err = common.UnmarshalFixed(m.db.Get(key))
	if err != nil {
		return nil
	}
	return
}

// putFixed ...
func (m *GasHandler) putFixed(key string, value *common.Fixed) {
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
	f := m.getFixed(m.gasRateKey(name))
	if f == nil {
		return &common.Fixed{
			Value:   0,
			Decimal: DecGasRate, // TODO set correct decimal of gas rate
		}
	}
	return f
}

// SetGasRate ...
func (m *GasHandler) SetGasRate(name string, r *common.Fixed) {
	m.putFixed(m.gasRateKey(name), r)
}

func (m *GasHandler) gasLimitKey(name string) string {
	return IOSTPrefix + name + gasLimitSuffix
}

// GasLimit ...
func (m *GasHandler) GasLimit(name string) *common.Fixed {
	f := m.getFixed(m.gasLimitKey(name))
	if f == nil {
		return &common.Fixed{
			Value:   0,
			Decimal: DecGas, // TODO set correct decimal of gas
		}
	}
	return f
}

// SetGasLimit ...
func (m *GasHandler) SetGasLimit(name string, l *common.Fixed) {
	m.putFixed(m.gasLimitKey(name), l)
}

func (m *GasHandler) gasUpdateTimeKey(name string) string {
	return IOSTPrefix + name + gasUpdateTimeSuffix
}

// GasUpdateTime ...
func (m *GasHandler) GasUpdateTime(name string) int64 {
	return m.getInt64(m.gasUpdateTimeKey(name))
}

// SetGasUpdateTime ...
func (m *GasHandler) SetGasUpdateTime(name string, t int64) {
	m.putInt64(m.gasUpdateTimeKey(name), t)
}

func (m *GasHandler) gasStockKey(name string) string {
	return IOSTPrefix + name + gasStockSuffix
}

// GasStock `gasStock` means the gas amount at last update time.
func (m *GasHandler) GasStock(name string) *common.Fixed {
	f := m.getFixed(m.gasStockKey(name))
	if f == nil {
		return &common.Fixed{
			Value:   0,
			Decimal: DecGas, // TODO set correct decimal of gas
		}
	}
	return f
}

// SetGasStock ...
func (m *GasHandler) SetGasStock(name string, g *common.Fixed) {
	m.putFixed(m.gasStockKey(name), g)
}

func (m *GasHandler) gasPledgeKey(name string) string {
	return IOSTPrefix + name + gasPledgeSuffix
}

// GasPledge ...
func (m *GasHandler) GasPledge(name string) *common.Fixed {
	f := m.getFixed(m.gasPledgeKey(name))
	if f == nil {
		return &common.Fixed{
			Value:   0,
			Decimal: DecPledge, // TODO set correct decimal of pledge
		}
	}
	return f
}

// SetGasPledge ...
func (m *GasHandler) SetGasPledge(name string, p *common.Fixed) {
	m.putFixed(m.gasPledgeKey(name), p)
}

// CurrentTotalGas return current total gas. It is min(limit, last_updated_gas + time_since_last_updated * increase_speed)
func (m *GasHandler) CurrentTotalGas(name string, now int64) (result *common.Fixed) {
	result = m.GasStock(name)
	gasUpdateTime := m.GasUpdateTime(name)
	var timeDuration int64
	if gasUpdateTime > 0 {
		timeDuration = now - gasUpdateTime
	}
	rate := m.GetGasRate(name)
	limit := m.GasLimit(name)
	//fmt.Printf("CurrentTotalGas stock %v rate %v limit %v", result, rate, limit)
	result = result.Add(rate.Times(timeDuration))
	if limit.LessThan(result) {
		result = limit
	}
	return
}
