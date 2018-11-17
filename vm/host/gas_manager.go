package host

import (
	"fmt"

	"github.com/iost-official/go-iost/common"
	"github.com/iost-official/go-iost/core/contract"
	"github.com/iost-official/go-iost/ilog"
)

// GasManager handle the logic of gas
type GasManager struct {
	h *Host
}

// NewGasManager new gas manager
func NewGasManager(h *Host) GasManager {
	return GasManager{
		h: h,
	}
}

const (
	// GasRateKey : how much gas is generated per IOST per second
	GasRateKey = "gr"
	// GasLimitKey : how much gas can be generated max per IOST
	GasLimitKey = "gl"
	// GasUpdateTimeKey : when the gas state is refreshed last time, for internal use
	GasUpdateTimeKey = "gt"
	// GasStockKey : how much gas is there when last time refreshed
	GasStockKey = "gs"
	// GasPledgeKey : who pledge how much coins for me
	GasPledgeKey = "gp"
)

// decimals of gas
const (
	GasDecimal = 2
)

// If no key exists, return 0
func (g *GasManager) getFixed(owner string, key string) (*common.Fixed, contract.Cost) {
	var err error
	result, cost := g.h.Get(key, owner)
	if result == nil {
		//ilog.Errorf("GasManager failed %v %v", owner, key)
		return nil, cost
	}
	value, err := common.UnmarshalFixed(result.(string))
	if err != nil {
		ilog.Errorf("GasManager failed %v %v %v", owner, key, result)
		return nil, cost
	}
	return value, cost
}

// putFixed ...
func (g *GasManager) putFixed(owner string, key string, value *common.Fixed) contract.Cost {
	if value.Err != nil {
		ilog.Fatalf("GasHandler putFixed %v", value)
	}
	//fmt.Printf("putFixed %v %v %v\n", owner, key, value.ToString())
	return g.h.Put(key, value.Marshal(), owner)
}

// GasRate ...
func (g *GasManager) GasRate(name string) (*common.Fixed, contract.Cost) {
	f, cost := g.getFixed(name, GasRateKey)
	if f == nil {
		return &common.Fixed{
			Value:   0,
			Decimal: GasDecimal,
		}, cost
	}
	return f, cost
}

// SetGasRate ...
func (g *GasManager) SetGasRate(name string, r *common.Fixed) contract.Cost {
	return g.putFixed(name, GasRateKey, r)
}

// GasLimit ...
func (g *GasManager) GasLimit(name string) (*common.Fixed, contract.Cost) {
	f, cost := g.getFixed(name, GasLimitKey)
	if f == nil {
		return &common.Fixed{
			Value:   0,
			Decimal: GasDecimal,
		}, cost
	}
	return f, cost
}

// SetGasLimit ...
func (g *GasManager) SetGasLimit(name string, l *common.Fixed) contract.Cost {
	return g.putFixed(name, GasLimitKey, l)
}

// GasUpdateTime ...
func (g *GasManager) GasUpdateTime(name string) (int64, contract.Cost) {
	value, cost := g.h.Get(GasUpdateTimeKey, name)
	if value == nil {
		return 0, cost
	}
	return value.(int64), cost
}

// SetGasUpdateTime ...
func (g *GasManager) SetGasUpdateTime(name string, t int64) contract.Cost {
	//ilog.Debugf("SetGasUpdateTime %v %v", name, t)
	return g.h.Put(GasUpdateTimeKey, t, name)
}

// GasStock `gasStock` means the gas amount at last update time.
func (g *GasManager) GasStock(name string) (*common.Fixed, contract.Cost) {
	f, cost := g.getFixed(name, GasStockKey)
	if f == nil {
		return &common.Fixed{
			Value:   0,
			Decimal: GasDecimal,
		}, cost
	}
	return f, cost
}

// SetGasStock ...
func (g *GasManager) SetGasStock(name string, gas *common.Fixed) contract.Cost {
	//ilog.Debugf("SetGasStock %v %v", name, g)
	return g.putFixed(name, GasStockKey, gas)
}

// GasPledge ...
func (g *GasManager) GasPledge(name string, pledger string) (*common.Fixed, contract.Cost) {
	finalCost := contract.Cost0()
	ok, cost := g.h.MapHas(GasPledgeKey, pledger, name)
	finalCost.AddAssign(cost)
	if !ok {
		return &common.Fixed{
			Value:   0,
			Decimal: 8,
		}, finalCost
	}
	result, cost := g.h.MapGet(GasPledgeKey, pledger, name)
	finalCost.AddAssign(cost)
	value, err := common.UnmarshalFixed(result.(string))
	if err != nil {
		return nil, finalCost
	}
	return value, finalCost
}

// SetGasPledge ...
func (g *GasManager) SetGasPledge(name string, pledger string, p *common.Fixed) contract.Cost {
	return g.h.MapPut(GasPledgeKey, pledger, p.Marshal(), name)
}

// DelGasPledge ...
func (g *GasManager) DelGasPledge(name string, pledger string) contract.Cost {
	if name == pledger {
		ilog.Fatalf("delGasPledge for oneself %v", name)
	}
	return g.h.MapDel(GasPledgeKey, pledger, name)
}

// PledgerInfo ...
type PledgerInfo struct {
	Pledger string
	Amount  *common.Fixed
}

// PledgerInfo get who pledged how much coins for me
func (g *GasManager) PledgerInfo(name string) ([]PledgerInfo, contract.Cost) {
	contractName, _ := g.h.ctx.Value("contract_name").(string)
	g.h.ctx.Set("contract_name", "iost.gas")
	finalCost := contract.Cost0()
	pledgers, cost := g.h.MapKeys(GasPledgeKey, name)
	//ilog.Errorf("pledge keys %v %v", pledgers, name)
	finalCost.AddAssign(cost)
	result := make([]PledgerInfo, 0)
	for _, pledger := range pledgers {
		v, cost := g.h.MapGet(GasPledgeKey, pledger, name)
		finalCost.AddAssign(cost)
		pledge, err := common.UnmarshalFixed(v.(string))
		if err != nil {
			return make([]PledgerInfo, 0), finalCost
		}
		result = append(result, PledgerInfo{pledger, pledge})
	}
	g.h.ctx.Set("contract_name", contractName)
	return result, finalCost
}

// CurrentTotalGas return current total gas. It is min(limit, last_updated_gas + time_since_last_updated * increase_speed)
func (g *GasManager) CurrentTotalGas(name string, now int64) (result *common.Fixed, finalCost contract.Cost) {
	if now <= 0 {
		ilog.Fatalf("CurrentTotalGas failed. invalid now time %v", now)
	}
	contractName, _ := g.h.ctx.Value("contract_name").(string)
	g.h.ctx.Set("contract_name", "iost.gas")
	result, cost := g.GasStock(name)
	finalCost.AddAssign(cost)
	gasUpdateTime, cost := g.GasUpdateTime(name)
	finalCost.AddAssign(cost)
	var durationSeconds int64
	if gasUpdateTime > 0 {
		durationSeconds = (now - gasUpdateTime) / 1e9
	}
	if durationSeconds < 0 {
		ilog.Fatalf("CurrentTotalGas durationSeconds invalid %v = %v - %v", durationSeconds, now, gasUpdateTime)
	}
	rate, cost := g.GasRate(name)
	finalCost.AddAssign(cost)
	limit, cost := g.GasLimit(name)
	finalCost.AddAssign(cost)
	//fmt.Printf("CurrentTotalGas user %v stock %v rate %v limit %v\n", name, result, rate, limit)
	delta := rate.Times(durationSeconds)
	if delta == nil {
		ilog.Errorf("CurrentTotalGas may overflow rate %v durationSeconds %v", rate, durationSeconds)
		return
	}
	result = result.Add(delta)
	if limit.LessThan(result) {
		result = limit
	}
	g.h.ctx.Set("contract_name", contractName)
	return
}

// CurrentGas returns the current total gas of a user. It is dynamically calculated
func (g *GasManager) CurrentGas(name string) (*common.Fixed, contract.Cost) {
	t := g.h.ctx.Value("time").(int64)
	if t <= 0 {
		ilog.Fatalf("CurrentGas invalid time %v", t)
	}
	return g.CurrentTotalGas(name, t)
}

func (g *GasManager) refreshGasWithValue(name string, value *common.Fixed) (contract.Cost, error) {
	finalCost := contract.Cost0()
	cost := g.SetGasStock(name, value)
	finalCost.AddAssign(cost)
	cost = g.SetGasUpdateTime(name, g.h.ctx.Value("time").(int64))
	finalCost.AddAssign(cost)
	return finalCost, nil
}

// RefreshGas update the gas status
func (g *GasManager) RefreshGas(name string) (contract.Cost, error) {
	finalCost := contract.Cost0()
	value, cost := g.CurrentGas(name)
	finalCost.AddAssign(cost)
	cost, err := g.refreshGasWithValue(name, value)
	finalCost.AddAssign(cost)
	return cost, err
}

// CostGas subtract gas of a user
func (g *GasManager) CostGas(name string, gasCost *common.Fixed) (contract.Cost, error) {
	finalCost := contract.Cost0()
	cost, err := g.RefreshGas(name)
	finalCost.AddAssign(cost)
	if err != nil {
		return finalCost, err
	}
	currentGas, cost := g.GasStock(name)
	finalCost.AddAssign(cost)
	b := currentGas.LessThan(gasCost)
	if b {
		return finalCost, fmt.Errorf("gas not enough! Now: %d, Need %d", currentGas, gasCost)
	}
	ret := currentGas.Sub(gasCost)
	cost = g.SetGasStock(name, ret)
	finalCost.AddAssign(cost)
	return finalCost, nil
}
