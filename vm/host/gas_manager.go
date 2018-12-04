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
func (g *GasManager) getFixed(key string) (*common.Fixed, contract.Cost) {
	result, cost := g.h.Get(key)
	if result == nil {
		//ilog.Errorf("GasManager failed %v", key)
		return nil, cost
	}
	value, ok := result.(*common.Fixed)
	if !ok {
		ilog.Errorf("GasManager failed %v %v", key, result)
		return nil, cost
	}
	return value, cost
}

// putFixed ...
func (g *GasManager) putFixed(key string, value *common.Fixed) contract.Cost {
	if value.Err != nil {
		ilog.Fatalf("GasHandler putFixed %v", value)
	}
	//fmt.Printf("putFixed %v %v\n", key, value.ToString())
	cost, err := g.h.Put(key, value)
	if err != nil {
		panic(fmt.Errorf("GasHandler putFixed err %v", err))
	}
	return cost
}

// GasRate ...
func (g *GasManager) GasRate(name string) (*common.Fixed, contract.Cost) {
	f, cost := g.getFixed(name + database.GasRateKey)
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
	return g.putFixed(name+database.GasRateKey, r)
}

// GasLimit ...
func (g *GasManager) GasLimit(name string) (*common.Fixed, contract.Cost) {
	f, cost := g.getFixed(name + database.GasLimitKey)
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
	return g.putFixed(name+database.GasLimitKey, l)
}

// GasUpdateTime ...
func (g *GasManager) GasUpdateTime(name string) (int64, contract.Cost) {
	value, cost := g.h.Get(name + database.GasUpdateTimeKey)
	if value == nil {
		return 0, cost
	}
	return value.(int64), cost
}

// SetGasUpdateTime ...
func (g *GasManager) SetGasUpdateTime(name string, t int64) contract.Cost {
	//ilog.Debugf("SetGasUpdateTime %v %v", name, t)
	cost, err := g.h.Put(name+database.GasUpdateTimeKey, t)
	if err != nil {
		panic(fmt.Errorf("gas manager set gas update time err, %v", err))
	}
	return cost
}

// GasStock `gasStock` means the gas amount at last update time.
func (g *GasManager) GasStock(name string) (*common.Fixed, contract.Cost) {
	f, cost := g.getFixed(name + database.GasStockKey)
	if f == nil {
		return &common.Fixed{
			Value:   0,
			Decimal: database.GasDecimal,
		}, cost
	}
	return f, cost
}

// TGas ...
func (g *GasManager) TGas(name string) (*common.Fixed, contract.Cost) {
	f, cost := g.getFixed(name + database.TransferableGasKey)
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
	return g.putFixed(name+database.GasStockKey, gas)
}

// ChangeTGas ...
func (g *GasManager) ChangeTGas(name string, delta *common.Fixed) contract.Cost {
	finalCost := contract.Cost0()
	f, cost := g.TGas(name)
	finalCost.AddAssign(cost)
	cost = g.putFixed(name+database.TransferableGasKey, f.Add(delta))
	finalCost.AddAssign(cost)
	return cost
}

// GasPledge ...
func (g *GasManager) GasPledge(name string, pledger string) (*common.Fixed, contract.Cost) {
	finalCost := contract.Cost0()
	ok, cost := g.h.MapHas(name+database.GasPledgeKey, pledger)
	finalCost.AddAssign(cost)
	if !ok {
		return &common.Fixed{
			Value:   0,
			Decimal: 8,
		}, finalCost
	}
	result, cost := g.h.MapGet(name+database.GasPledgeKey, pledger)
	finalCost.AddAssign(cost)
	value, ok := result.(*common.Fixed)
	if !ok {
		return nil, finalCost
	}
	return value, finalCost
}

// SetGasPledge ...
func (g *GasManager) SetGasPledge(name string, pledger string, p *common.Fixed) contract.Cost {
	cost, err := g.h.MapPut(name+database.GasPledgeKey, pledger, p)
	if err != nil {
		panic(fmt.Errorf("gas manager set gas pledge err %v", err))
	}
	return cost
}

// DelGasPledge ...
func (g *GasManager) DelGasPledge(name string, pledger string) contract.Cost {
	if name == pledger {
		ilog.Fatalf("delGasPledge for oneself %v", name)
	}
	cost, err := g.h.MapDel(name+database.GasPledgeKey, pledger)
	if err != nil {
		panic(fmt.Errorf("gas manager del gas pledge err %v", err))
	}
	return cost
}

func (g *GasManager) refreshPGasWithValue(name string, value *common.Fixed) (contract.Cost, error) {
	finalCost := contract.Cost0()
	cost := g.SetGasStock(name, value)
	finalCost.AddAssign(cost)
	cost = g.SetGasUpdateTime(name, g.h.ctx.Value("time").(int64))
	finalCost.AddAssign(cost)
	return finalCost, nil
}

// PGas returns the current total gas of a user. It is dynamically calculated
func (g *GasManager) PGas(name string) *common.Fixed {
	t := g.h.ctx.Value("time").(int64)
	if t <= 0 {
		ilog.Fatalf("PGas invalid time %v", t)
	}
	return g.h.DB().PGasAtTime(name, t)
}

// AllGas ...
func (g *GasManager) AllGas(name string) *common.Fixed {
	return g.PGas(name).Add(g.h.DB().TGas(name))
}

// RefreshPGas update the gas status
func (g *GasManager) RefreshPGas(name string) (contract.Cost, error) {
	finalCost := contract.Cost0()
	value := g.PGas(name)
	cost, err := g.refreshPGasWithValue(name, value)
	finalCost.AddAssign(cost)
	return cost, err
}

// CostGas subtract gas of a user. It is not called in a contract. Need a better design here
func (g *GasManager) CostGas(name string, gasCost *common.Fixed) error {
	// todo modify CostGas
	oldVal := g.h.ctx.Value("contract_name")
	g.h.ctx.Set("contract_name", "gas.iost")
	_, err := g.RefreshPGas(name)
	if err != nil {
		return err
	}
	currentPGas, _ := g.GasStock(name)
	currentTGas := g.h.DB().TGas(name)
	if currentPGas.Add(currentTGas).LessThan(gasCost) {
		return fmt.Errorf("gas not enough! Now: %v(tgas:%v,pgas:%v), Need %v", currentTGas.Add(currentPGas).ToString(), currentPGas.ToString(), currentTGas.ToString(), gasCost.ToString())
	}
	if currentPGas.LessThan(gasCost) {
		g.SetGasStock(name, currentPGas.Sub(currentPGas))
		g.ChangeTGas(name, gasCost.Sub(currentPGas).Neg())
	} else {
		newPGas := currentPGas.Sub(gasCost)
		g.SetGasStock(name, newPGas)
	}
	g.h.ctx.Set("contract_name", oldVal)
	return nil
}
