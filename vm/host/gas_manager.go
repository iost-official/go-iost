package host

import (
	"fmt"
	"github.com/iost-official/go-iost/vm/database"

	"github.com/iost-official/go-iost/common"
	"github.com/iost-official/go-iost/core/contract"
	"github.com/iost-official/go-iost/ilog"
)

// GasManager handle the logic of gas. It should be called in a contract
type GasManager struct {
	h *Host
}

// NewGasManager new gas manager
func NewGasManager(h *Host) GasManager {
	return GasManager{
		h: h,
	}
}

// If no key exists, return 0
func (g *GasManager) getFixed(owner string, key string) (*common.Fixed, contract.Cost) {
	result, cost := g.h.Get(key + owner)
	if result == nil {
		//ilog.Errorf("GasManager failed %v %v", owner, key)
		return nil, cost
	}
	value, ok := result.(*common.Fixed)
	if !ok {
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
	return g.h.Put(key+owner, value)
}

// GasRate ...
func (g *GasManager) GasRate(name string) (*common.Fixed, contract.Cost) {
	f, cost := g.getFixed(name, database.GasRateKey)
	if f == nil {
		return &common.Fixed{
			Value:   0,
			Decimal: database.GasDecimal,
		}, cost
	}
	return f, cost
}

// SetGasRate ...
func (g *GasManager) SetGasRate(name string, r *common.Fixed) contract.Cost {
	return g.putFixed(name, database.GasRateKey, r)
}

// GasLimit ...
func (g *GasManager) GasLimit(name string) (*common.Fixed, contract.Cost) {
	f, cost := g.getFixed(name, database.GasLimitKey)
	if f == nil {
		return &common.Fixed{
			Value:   0,
			Decimal: database.GasDecimal,
		}, cost
	}
	return f, cost
}

// SetGasLimit ...
func (g *GasManager) SetGasLimit(name string, l *common.Fixed) contract.Cost {
	return g.putFixed(name, database.GasLimitKey, l)
}

// GasUpdateTime ...
func (g *GasManager) GasUpdateTime(name string) (int64, contract.Cost) {
	value, cost := g.h.Get(database.GasUpdateTimeKey + name)
	if value == nil {
		return 0, cost
	}
	return value.(int64), cost
}

// SetGasUpdateTime ...
func (g *GasManager) SetGasUpdateTime(name string, t int64) contract.Cost {
	//ilog.Debugf("SetGasUpdateTime %v %v", name, t)
	return g.h.Put(database.GasUpdateTimeKey+name, t)
}

// GasStock `gasStock` means the gas amount at last update time.
func (g *GasManager) GasStock(name string) (*common.Fixed, contract.Cost) {
	f, cost := g.getFixed(name, database.GasStockKey)
	if f == nil {
		return &common.Fixed{
			Value:   0,
			Decimal: database.GasDecimal,
		}, cost
	}
	return f, cost
}

// SetGasStock ...
func (g *GasManager) SetGasStock(name string, gas *common.Fixed) contract.Cost {
	//ilog.Debugf("SetGasStock %v %v", name, g)
	return g.putFixed(name, database.GasStockKey, gas)
}

// GasPledge ...
func (g *GasManager) GasPledge(name string, pledger string) (*common.Fixed, contract.Cost) {
	finalCost := contract.Cost0()
	ok, cost := g.h.MapHas(database.GasPledgeKey+name, pledger)
	finalCost.AddAssign(cost)
	if !ok {
		return &common.Fixed{
			Value:   0,
			Decimal: 8,
		}, finalCost
	}
	result, cost := g.h.MapGet(database.GasPledgeKey+name, pledger)
	finalCost.AddAssign(cost)
	value, ok := result.(*common.Fixed)
	if !ok {
		return nil, finalCost
	}
	return value, finalCost
}

// SetGasPledge ...
func (g *GasManager) SetGasPledge(name string, pledger string, p *common.Fixed) contract.Cost {
	return g.h.MapPut(database.GasPledgeKey+name, pledger, p)
}

// DelGasPledge ...
func (g *GasManager) DelGasPledge(name string, pledger string) contract.Cost {
	if name == pledger {
		ilog.Fatalf("delGasPledge for oneself %v", name)
	}
	return g.h.MapDel(database.GasPledgeKey+name, pledger)
}

func (g *GasManager) refreshGasWithValue(name string, value *common.Fixed) (contract.Cost, error) {
	finalCost := contract.Cost0()
	cost := g.SetGasStock(name, value)
	finalCost.AddAssign(cost)
	cost = g.SetGasUpdateTime(name, g.h.ctx.Value("time").(int64))
	finalCost.AddAssign(cost)
	return finalCost, nil
}

// CurrentGas returns the current total gas of a user. It is dynamically calculated
func (g *GasManager) CurrentGas(name string) *common.Fixed {
	t := g.h.ctx.Value("time").(int64)
	if t <= 0 {
		ilog.Fatalf("CurrentGas invalid time %v", t)
	}
	return g.h.DB().CurrentTotalGas(name, t)
}

// RefreshGas update the gas status
func (g *GasManager) RefreshGas(name string) (contract.Cost, error) {
	finalCost := contract.Cost0()
	value := g.CurrentGas(name)
	cost, err := g.refreshGasWithValue(name, value)
	finalCost.AddAssign(cost)
	return cost, err
}

// CostGas subtract gas of a user. It is not called in a contract. Need a better design here
func (g *GasManager) CostGas(name string, gasCost *common.Fixed) error {
	// todo modify CostGas
	oldVal := g.h.ctx.Value("contract_name")
	g.h.ctx.Set("contract_name", "gas.iost")
	_, err := g.RefreshGas(name)
	if err != nil {
		return err
	}
	currentGas, _ := g.GasStock(name)
	b := currentGas.LessThan(gasCost)
	if b {
		return fmt.Errorf("gas not enough! Now: %s, Need %s", currentGas.ToString(), gasCost.ToString())
	}
	ret := currentGas.Sub(gasCost)
	g.SetGasStock(name, ret)
	g.h.ctx.Set("contract_name", oldVal)
	return nil
}
