package host

import (
	"fmt"
)

const (
	// GasMinPledge Every user must pledge a minimum amount of IOST (including GAS and RAM)
	GasMinPledge = 100

	// Each IOST you pledge, you will get `GasImmediateReward` gas immediately.
	// Then gas will be generated at a rate of `GasIncreaseRate` gas per block.
	// Your gas production will stop when it reaches the limit.
	// When you use some gas later, the total amount will be less than the limit,
	// so gas production will continue again util the limit.

	// GasImmediateReward immediate reward per IOST
	GasImmediateReward = 10
	// GasLimit gas limit per IOST
	GasLimit = 30
	// GasIncreaseRate gas increase per IOST per block
	GasIncreaseRate = 1
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
	return g.h.db.GasHandler.CurrentTotalGas(name, blockNumber)
}

func (g *GasManager) refreshGasWithValue(name string, value int64) error {
	g.h.db.GasHandler.SetGasStock(name, value)
	g.h.db.GasHandler.SetGasUpdateTime(name, g.h.ctx.Value("number").(int64))
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
	currentGas := g.h.db.GasHandler.GetGasStock(name)
	if currentGas < cost {
		return fmt.Errorf("Gas not enough! Now: %d, Need %d", currentGas, cost)
	}
	g.h.db.GasHandler.SetGasStock(name, currentGas-cost)
	return nil
}

// Pledge Change all gas related storage here. If pledgeAmount > 0. pledge. If pledgeAmount < 0, unpledge.
func (g *GasManager) Pledge(name string, pledgeAmount int64) error {
	if pledgeAmount == 0 {
		return fmt.Errorf("invalid pledge amount %d", pledgeAmount)
	}
	if pledgeAmount < 0 {
		unpledgeAmount := -pledgeAmount
		pledged := g.h.db.GasHandler.GetGasPledge(name)
		// how to deal with overflow here?
		if pledged-unpledgeAmount < GasMinPledge {
			return fmt.Errorf("%d should be pledged at least ", GasMinPledge)
		}
	}

	limitDelta := pledgeAmount * GasLimit
	rateDelta := pledgeAmount * GasIncreaseRate
	gasDelta := pledgeAmount * GasImmediateReward
	if pledgeAmount < 0 {
		// unpledge should not change current generated gas
		gasDelta = 0
	}

	// pledge first time
	if g.h.db.GasHandler.GetGasUpdateTime(name) == 0 {
		if pledgeAmount < 0 {
			return fmt.Errorf("cannot unpledge! No pledge before")
		}
		g.h.db.GasHandler.SetGasPledge(name, pledgeAmount)
		g.h.db.GasHandler.SetGasUpdateTime(name, g.h.ctx.Value("number").(int64))
		g.h.db.GasHandler.SetGasRate(name, rateDelta)
		g.h.db.GasHandler.SetGasLimit(name, limitDelta)
		g.h.db.GasHandler.SetGasStock(name, gasDelta)
		return nil
	}
	g.RefreshGas(name)
	rateOld := g.h.db.GasHandler.GetGasRate(name)
	rateNew := rateOld + rateDelta
	if rateNew <= 0 {
		return fmt.Errorf("change gasRate failed! current: %d, delta %d", rateOld, rateDelta)
	}
	limitOld := g.h.db.GasHandler.GetGasLimit(name)
	limitNew := limitOld + limitDelta
	if limitNew <= 0 {
		return fmt.Errorf("change gasLimit failed! current: %d, delta %d", limitOld, limitDelta)
	}
	gasOld := g.h.db.GasHandler.GetGasStock(name)
	gasNew := gasOld + gasDelta
	if gasNew > limitNew {
		// clear the gas above the new limit.
		gasNew = limitNew
	}

	g.h.db.GasHandler.SetGasPledge(name, g.h.db.GasHandler.GetGasPledge(name)+pledgeAmount)
	g.h.db.GasHandler.SetGasRate(name, rateNew)
	g.h.db.GasHandler.SetGasLimit(name, limitNew)
	g.h.db.GasHandler.SetGasStock(name, gasNew)
	return nil
}
