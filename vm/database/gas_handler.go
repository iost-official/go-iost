package database

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

// GetInt64 return the value as int64. If no key exists, return 0
func (m *GasHandler) GetInt64(key string) (value int64) {
	value, _ = Unmarshal(m.db.Get(key)).(int64)
	return
}

// GetFloat64 return the value as float64. If no key exists, return 0
func (m *GasHandler) GetFloat64(key string) (value float64) {
	value, _ = Unmarshal(m.db.Get(key)).(float64)
	return
}

// PutInt64 ...
func (m *GasHandler) PutInt64(key string, value int64) {
	m.db.Put(key, MustMarshal(value))
}

// PutFloat64 ...
func (m *GasHandler) PutFloat64(key string, value float64) {
	m.db.Put(key, MustMarshal(value))
}

func (m *GasHandler) gasRateKey(name string) string {
	return IOSTPrefix + name + gasRateSuffix
}

// GetGasRate ...
func (m *GasHandler) GetGasRate(name string) int64 {
	return m.GetInt64(m.gasRateKey(name))
}

// SetGasRate ...
func (m *GasHandler) SetGasRate(name string, r int64) {
	m.PutInt64(m.gasRateKey(name), r)
}

func (m *GasHandler) gasLimitKey(name string) string {
	return IOSTPrefix + name + gasLimitSuffix
}

// GetGasLimit ...
func (m *GasHandler) GetGasLimit(name string) int64 {
	return m.GetInt64(m.gasLimitKey(name))
}

// SetGasLimit ...
func (m *GasHandler) SetGasLimit(name string, l int64) {
	m.PutInt64(m.gasLimitKey(name), l)
}

func (m *GasHandler) gasUpdateTimeKey(name string) string {
	return IOSTPrefix + name + gasUpdateTimeSuffix
}

// GetGasUpdateTime ...
func (m *GasHandler) GetGasUpdateTime(name string) int64 {
	return m.GetInt64(m.gasUpdateTimeKey(name))
}

// SetGasUpdateTime ...
func (m *GasHandler) SetGasUpdateTime(name string, t int64) {
	m.PutInt64(m.gasUpdateTimeKey(name), t)
}

func (m *GasHandler) gasStockKey(name string) string {
	return IOSTPrefix + name + gasStockSuffix
}

// GetGasStock `gasStock` means the gas amount at last update time.
func (m *GasHandler) GetGasStock(name string) int64 {
	return m.GetInt64(m.gasStockKey(name))
}

// SetGasStock ...
func (m *GasHandler) SetGasStock(name string, g int64) {
	m.PutInt64(m.gasStockKey(name), g)
}

func (m *GasHandler) gasPledgeKey(name string) string {
	return IOSTPrefix + name + gasPledgeSuffix
}

// GetGasPledge ...
func (m *GasHandler) GetGasPledge(name string) int64 {
	return m.GetInt64(m.gasPledgeKey(name))
}

// SetGasPledge ...
func (m *GasHandler) SetGasPledge(name string, p int64) {
	m.PutInt64(m.gasPledgeKey(name), p)
}

// CurrentTotalGas return current total gas. It is min(limit, last_updated_gas + time_since_last_updated * increase_speed)
func (m *GasHandler) CurrentTotalGas(name string, now int64) (result int64) {
	result = m.GetGasStock(name)
	gasUpdateTime := m.GetGasUpdateTime(name)
	var timeDuration int64
	if gasUpdateTime > 0 {
		timeDuration = now - gasUpdateTime
	}
	rate := m.GetGasRate(name)
	limit := m.GetGasLimit(name)
	result += timeDuration * rate
	if result > limit {
		result = limit
	}
	return
}
