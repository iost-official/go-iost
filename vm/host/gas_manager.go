package host

import (
	"fmt"
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
	blockNumber := g.h.ctx.Value("number").(int64)
	return g.h.db.BalanceHandler.CurrentTotalGas(name, blockNumber)
}

func (g *GasManager) refreshGasWithValue(name string, value int64) error {
	g.h.db.BalanceHandler.SetGasStock(name, value)
	g.h.db.BalanceHandler.SetGasUpdateTime(name, g.h.ctx.Value("number").(int64))
	return nil
}

// RefreshGas update the gas status
func (g *GasManager) RefreshGas(name string) error {
	return g.refreshGasWithValue(name, g.CurrentGas(name))
}

// CostGas subtract gas of a user
func (g *GasManager) CostGas(name string, cost int64) error {
	err := g.RefreshGas(name)
	if err != nil {
		return err
	}
	currentGas := g.h.db.BalanceHandler.GetGasStock(name)
	if currentGas < cost {
		return fmt.Errorf("Gas not enough! Now: %d, Need %d", currentGas, cost)
	}
	g.h.db.BalanceHandler.SetGasStock(name, currentGas-cost)
	return nil
}

// ChangeGas ...
func (g *GasManager) ChangeGas(name string, gasStockDelta int64, rateDelta int64, limitDelta int64) error {
	// pledge first time
	if g.h.db.BalanceHandler.GetGasUpdateTime(name) == 0 {
		g.h.db.BalanceHandler.SetGasUpdateTime(name, g.h.ctx.Value("number").(int64))
		g.h.db.BalanceHandler.SetGasRate(name, rateDelta)
		g.h.db.BalanceHandler.SetGasLimit(name, limitDelta)
		g.h.db.BalanceHandler.SetGasStock(name, gasStockDelta)
		return nil
	}
	g.RefreshGas(name)
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
	g.h.db.BalanceHandler.SetGasRate(name, rateNew)
	g.h.db.BalanceHandler.SetGasLimit(name, limitNew)
	// clear the gas above the new limit. This can also be written as 'g.RefreshGas(name)'
	if g.h.db.BalanceHandler.GetGasStock(name) > limitNew {
		g.h.db.BalanceHandler.SetGasStock(name, limitNew)
	}
	return nil
}
