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
func (g *GasManager) CurrentGas(name string) *common.Fixed {
	blockNumber := g.h.ctx.Value("number").(int64)
	return g.h.db.GasHandler.CurrentTotalGas(name, blockNumber)
}

func (g *GasManager) refreshGasWithValue(name string, value *common.Fixed) error {
	g.h.db.GasHandler.SetGasStock(name, value)
	g.h.db.GasHandler.SetGasUpdateTime(name, g.h.ctx.Value("number").(int64))
	return nil
}

// RefreshGas update the gas status
func (g *GasManager) RefreshGas(name string) error {
	return g.refreshGasWithValue(name, g.CurrentGas(name))
}

// CostGas subtract gas of a user
func (g *GasManager) CostGas(name string, cost *common.Fixed) error {
	err := g.RefreshGas(name)
	if err != nil {
		return err
	}
	currentGas := g.h.db.GasHandler.GasStock(name)
	if currentGas.LessThan(cost) {
		return fmt.Errorf("gas not enough! Now: %d, Need %d", currentGas, cost)
	}
	g.h.db.GasHandler.SetGasStock(name, currentGas.Sub(cost))
	return nil
}
