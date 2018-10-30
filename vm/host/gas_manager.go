package host

import (
	"fmt"
	"github.com/iost-official/go-iost/common"
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
func (g *GasManager) CurrentGas(name string) *common.FixPointNumber {
	blockNumber := g.h.ctx.Value("number").(int64)
	return g.h.db.GasHandler.CurrentTotalGas(name, blockNumber)
}

func (g *GasManager) refreshGasWithValue(name string, value *common.FixPointNumber) error {
	g.h.db.GasHandler.SetGasStock(name, value)
	g.h.db.GasHandler.SetGasUpdateTime(name, g.h.ctx.Value("number").(int64))
	return nil
}

// RefreshGas update the gas status
func (g *GasManager) RefreshGas(name string) error {
	return g.refreshGasWithValue(name, g.CurrentGas(name))
}

// CostGas subtract gas of a user
func (g *GasManager) CostGas(name string, cost *common.FixPointNumber) error {
	err := g.RefreshGas(name)
	if err != nil {
		return err
	}
	currentGas := g.h.db.GasHandler.GetGasStock(name)
	if currentGas.LessThen(cost) {
		return fmt.Errorf("Gas not enough! Now: %d, Need %d", currentGas, cost)
	}
	g.h.db.GasHandler.SetGasStock(name, currentGas.Sub(cost))
	return nil
}
