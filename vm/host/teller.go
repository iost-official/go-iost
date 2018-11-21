package host

import (
	"fmt"
	"strings"

	"github.com/iost-official/go-iost/common"
	"github.com/iost-official/go-iost/core/contract"
)

// const prefixs
const (
	ContractAccountPrefix = "CA"
	ContractGasPrefix     = "CG"
)

// Teller handler of iost
type Teller struct {
	h    *Host
	cost map[string]contract.Cost
	negcost map[string]contract.Cost
}

// NewTeller new teller
func NewTeller(h *Host) Teller {
	return Teller{
		h:    h,
		cost: make(map[string]contract.Cost),
	}
}

// Costs ...
func (h *Teller) Costs() map[string]contract.Cost {
	return h.cost
}

// ClearCosts ...
func (h *Teller) ClearCosts() {
	h.cost = make(map[string]contract.Cost)
}

// PayCost ...
func (h *Teller) PayCost(c contract.Cost, who string) {
	if oc, ok := h.cost[who]; ok {
		oc.AddAssign(c)
		h.cost[who] = oc
	} else {
		h.cost[who] = c
	}
}

// DoPay ...
func (h *Teller) DoPay(witness string, gasPrice int64, isPayRAM bool) error {
	//if gasPrice < 100 {
	//	panic("gas_price error")
	//}
	for k, c := range h.cost {
		fee := gasPrice * c.ToGas()
		if fee != 0 {
			gas := &common.Fixed{
				Value:   fee,
				Decimal: 2,
			}
			_, err := h.h.CostGas(k, gas)
			if err != nil {
				return fmt.Errorf("pay cost failed: %v, %v", k, err)
			}
		}
		// contracts in "iost" domain will not pay for ram
		if isPayRAM && c.Data > 0 && !strings.HasSuffix(k, ".iost") {
			var payer string
			if strings.HasPrefix(k, "Contract") {
				p, _ := h.h.GlobalMapGet("system.iost", "contract_owner", k)
				payer = p.(string)
			} else {
				payer = k
			}

			ram := c.Data
			currentRAM := h.h.db.TokenBalance("ram", payer)
			if currentRAM-ram < 0 {
				return fmt.Errorf("pay ram failed. id: %v need %v, actual %v", payer, ram, currentRAM)
			}
			h.h.db.SetTokenBalance("ram", payer, currentRAM-ram)
		}
	}
	return nil
}

// Privilege ...
func (h *Teller) Privilege(id string) int {
	am, ok := h.h.ctx.Value("auth_list").(map[string]int)
	if !ok {
		return 0
	}
	i, ok := am[id]
	if !ok {
		i = 0
	}
	return i
}
