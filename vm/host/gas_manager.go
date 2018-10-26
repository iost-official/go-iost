package host

import (
	"fmt"
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

// CurrentGas returns the current total gas of a user. It is dynamically calculated
func (g *GasManager) CurrentGas(name string) int64 {
	gasUpdateTime := g.h.db.BalanceHandler.GetGasUpdateTime(name)
	if gasUpdateTime == 0 {
		ilog.Errorf("user %s gasUpdateTime is 0", name)
		return 0
	}
	blockNumber := g.h.ctx.Value("number").(int64)
	rate := g.h.db.BalanceHandler.GetGasRate(name)
	limit := g.h.db.BalanceHandler.GetGasLimit(name)
	if limit == 0 {
		ilog.Errorf("user %s gasLimit is 0", name)
		return 0
	}
	gasStock := g.h.db.BalanceHandler.GetGas(name)
	timeDuration := blockNumber - gasUpdateTime
	result := timeDuration*rate + gasStock
	if result > limit {
		return limit
	}
	return result
}

func (g *GasManager) refreshGasWithValue(name string, value int64) error {
	g.h.db.BalanceHandler.SetGas(name, value)
	g.h.db.BalanceHandler.SetGasUpdateTime(name, g.h.ctx.Value("number").(int64))
	return nil
}

// RefreshGas update the gas status
func (g *GasManager) RefreshGas(name string) error {
	return g.refreshGasWithValue(name, g.CurrentGas(name))
}

// CostGas subtract gas of a user
func (g *GasManager) CostGas(name string, cost int64) error {
	currentStaticGas := g.h.db.BalanceHandler.GetGas(name)
	// the fast pass can avoid some IO in most cases
	if currentStaticGas >= cost {
		g.h.db.BalanceHandler.SetGas(name, currentStaticGas-cost)
		return nil
	}
	currentTotalGas := g.CurrentGas(name)
	if cost > currentTotalGas {
		return fmt.Errorf("Gas not enough! Now: %d, Need %d", currentTotalGas, cost)
	}
	return g.refreshGasWithValue(name, currentTotalGas-cost)
}

// ChangeGasRateAndLimit ...
func (g *GasManager) ChangeGasRateAndLimit(name string, rateDelta int64, limitDelta int64) error {
	rateOld := g.h.db.BalanceHandler.GetGasRate(name)
	rateNew := rateOld + rateDelta
	if rateNew <= 0 {
		return fmt.Errorf("change gasRate failed! current: %d, delta %d", rateOld, rateDelta)
	}
	limitOld := g.h.db.BalanceHandler.GetGasLimit(name)
	limitNew := limitOld + limitDelta
	if limitNew <= 0 {
		return fmt.Errorf("change gasLimit failed! current: %d, delta %d", limitOld, limitDelta)
	}
	g.RefreshGas(name)
	g.h.db.BalanceHandler.SetGasRate(name, rateNew)
	g.h.db.BalanceHandler.SetGasLimit(name, limitNew)
	// clear the gas above the new limit. This can also be written as 'g.RefreshGas(name)'
	if g.h.db.BalanceHandler.GetGas(name) > limitNew {
		g.h.db.BalanceHandler.SetGas(name, limitNew)
	}
	return nil
}
