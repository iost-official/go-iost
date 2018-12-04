package host

import (
	"fmt"
	"github.com/iost-official/go-iost/ilog"
	"strings"

	"github.com/iost-official/go-iost/common"
	"github.com/iost-official/go-iost/core/contract"
)

// Teller handler of iost
type Teller struct {
	h         *Host
	cost      map[string]contract.Cost
	cacheCost contract.Cost
}

// NewTeller new teller
func NewTeller(h *Host) Teller {
	return Teller{
		h:    h,
		cost: make(map[string]contract.Cost),
	}
}

// Costs ...
func (t *Teller) Costs() map[string]contract.Cost {
	return t.cost
}

// ClearCosts ...
func (t *Teller) ClearCosts() {
	t.cost = make(map[string]contract.Cost)
}

// ClearRAMCosts ...
func (t *Teller) ClearRAMCosts() {
	newCost := make(map[string]contract.Cost)
	for k, c := range t.cost {
		if c.Net != 0 || c.CPU != 0 {
			newCost[k] = contract.NewCost(0, c.Net, c.CPU)
		}
	}
	t.cost = newCost
}

// AddCacheCost ...
func (t *Teller) AddCacheCost(c contract.Cost) {
	t.cacheCost.AddAssign(c)
}

// CacheCost ...
func (t *Teller) CacheCost() contract.Cost {
	return t.cacheCost
}

// FlushCacheCost ...
func (t *Teller) FlushCacheCost() {
	t.PayCost(t.cacheCost, "")
	t.cacheCost = contract.Cost0()
}

// ClearCacheCost ...
func (t *Teller) ClearCacheCost() {
	t.cacheCost = contract.Cost0()
}

// PayCost ...
func (t *Teller) PayCost(c contract.Cost, who string) {
	costMap := make(map[string]contract.Cost)
	if c.CPU > 0 || c.Net > 0 {
		costMap[who] = contract.Cost{CPU: c.CPU, Net: c.Net}
	}
	for _, item := range c.DataList {
		if oc, ok := costMap[item.Payer]; ok {
			oc.AddAssign(contract.Cost{Data: item.Val, DataList: []contract.DataItem{item}})
			costMap[item.Payer] = oc
		} else {
			costMap[item.Payer] = contract.Cost{Data: item.Val, DataList: []contract.DataItem{item}}
		}
	}
	for who, c := range costMap {
		if oc, ok := t.cost[who]; ok {
			oc.AddAssign(c)
			t.cost[who] = oc
		} else {
			t.cost[who] = c
		}
	}
}

// DoPay ...
func (t *Teller) DoPay(witness string, gasRatio int64) error {
	for k, c := range t.cost {
		fee := gasRatio * c.ToGas()
		if fee != 0 {
			gas := &common.Fixed{
				Value:   fee,
				Decimal: 2,
			}
			err := t.h.CostGas(k, gas)
			if err != nil {
				return fmt.Errorf("pay cost failed: %v, %v", k, err)
			}
			// reward 15% gas to account referrer
			if !t.h.IsContract(k) {
				acc, _ := ReadAuth(t.h.DB(), k)
				if acc == nil {
					ilog.Fatalf("invalid account %v", k)
				}
				if acc.Referrer != "" {
					reward := gas.TimesF(0.15)
					t.h.ChangeTGas(acc.Referrer, reward)
				}
			}
		}
		// contracts in "iost" domain will not pay for ram
		if !strings.HasSuffix(k, ".iost") {
			var payer string
			if t.h.IsContract(k) {
				p, _ := t.h.GlobalMapGet("system.iost", "contract_owner", k)
				var ok bool
				payer, ok = p.(string)
				if !ok {
					return fmt.Errorf("DoPay failed: contract %v has no owner", k)
				}
			} else {
				payer = k
			}

			ram := c.Data
			currentRAM := t.h.db.TokenBalance("ram", payer)
			if currentRAM-ram < 0 {
				return fmt.Errorf("pay ram failed. id: %v need %v, actual %v", payer, ram, currentRAM)
			}
			t.h.db.SetTokenBalance("ram", payer, currentRAM-ram)
		}
	}
	return nil
}

// Privilege ...
func (t *Teller) Privilege(id string) int {
	am, ok := t.h.ctx.Value("auth_list").(map[string]int)
	if !ok {
		return 0
	}
	i, ok := am[id]
	if !ok {
		i = 0
	}
	return i
}
