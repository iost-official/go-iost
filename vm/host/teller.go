package host

import (
	"fmt"
	"strings"

	"github.com/iost-official/go-iost/common"
	"github.com/iost-official/go-iost/core/contract"
	"github.com/iost-official/go-iost/ilog"
)

// const prefixs
const (
	ContractAccountPrefix = "CA"
	ContractGasPrefix     = "CG"
)

// Teller handler of iost
type Teller struct {
	h    *Host
	cost map[string]*contract.Cost
}

// NewTeller new teller
func NewTeller(h *Host) Teller {
	return Teller{
		h:    h,
		cost: make(map[string]*contract.Cost),
	}
}

// TransferRawNew ...
func (h *Teller) TransferRawNew(from, to string, amount int64) error {
	tokenName := "iost"
	srcBalance := h.h.db.TokenBalance(tokenName, from)
	dstBalance := h.h.db.TokenBalance(tokenName, to)
	if strings.HasPrefix(from, ContractAccountPrefix) && srcBalance >= amount || srcBalance > amount {
		h.h.db.SetTokenBalance(tokenName, from, srcBalance-amount)
		h.h.db.SetTokenBalance(tokenName, to, dstBalance+amount)
		ilog.Debugf("TransferRaw src %v: %v -> %v and dst %v: %v -> %v\n", from, srcBalance, srcBalance-amount, to, dstBalance, dstBalance+amount)
		return nil
	}
	ilog.Debugf("TransferRaw balance not enough %v has %v < %v", from, srcBalance, amount)
	return ErrBalanceNotEnough
}

// todo deprecated GrantCoin issue coin
func (h *Teller) GrantCoin(coinName, to string, amountStr string) (*contract.Cost, error) {
	amount, _ := common.NewFixed(amountStr, 8)
	if amount.Value <= 0 {
		return CommonErrorCost(1), ErrTransferNegValue
	}
	cn := h.h.ctx.Value("contract_name").(string)
	if !strings.HasPrefix(cn, "iost.") {
		return CommonErrorCost(2), ErrPermissionLost
	}
	h.h.db.SetCoin(coinName, to, amount.Value)
	return TransferCost, nil
}

// todo deprecated ConsumeCoin consume coin from
func (h *Teller) ConsumeCoin(coinName, from string, amountStr string) (cost *contract.Cost, err error) {
	amount, _ := common.NewFixed(amountStr, 8)
	if amount.Value <= 0 {
		return CommonErrorCost(1), ErrTransferNegValue
	}
	if h.Privilege(from) < 1 {
		return CommonErrorCost(1), ErrPermissionLost
	}
	bl := h.h.db.Coin(coinName, from)
	if bl < amount.Value {
		return CommonErrorCost(2), ErrBalanceNotEnough
	}
	h.h.db.SetCoin(coinName, from, -1*amount.Value)
	return TransferCost, nil
}

// todo deprecated GrantServi ...
func (h *Teller) GrantServi(to string, amountStr string) (*contract.Cost, error) {
	amount, _ := common.NewFixed(amountStr, 8)
	if amount.Value <= 0 {
		return CommonErrorCost(1), ErrTransferNegValue
	}
	//cn := h.h.ctx.Value("contract_name").(string) todo privilege of system contracts
	//if !strings.HasPrefix(cn, "iost.") {
	//	return CommonErrorCost(2), ErrPermissionLost
	//}
	h.h.db.SetServi(to, amount.Value)
	return TransferCost, nil
}

// todo deprecated ConsumeServi ...
func (h *Teller) ConsumeServi(from string, amountStr string) (cost *contract.Cost, err error) {
	amount, _ := common.NewFixed(amountStr, 8)
	if amount.Value <= 0 {
		return CommonErrorCost(1), ErrTransferNegValue
	}
	if h.Privilege(from) < 1 {
		return CommonErrorCost(1), ErrPermissionLost
	}
	bl := h.h.db.Servi(from)
	if bl < amount.Value {
		return CommonErrorCost(2), ErrBalanceNotEnough
	}
	h.h.db.SetServi(from, -1*amount.Value)
	return TransferCost, nil
}

// todo deprecated TotalServi ...
func (h *Teller) TotalServi() (ts string, cost *contract.Cost) {
	fpn := common.Fixed{Value: h.h.db.TotalServi(), Decimal: 8}
	ts = fpn.ToString()
	cost = GetCost
	return
}

// PayCost ...
func (h *Teller) PayCost(c *contract.Cost, who string) {
	h.cost[who] = c
}

// DoPay ...
func (h *Teller) DoPay(witness string, gasPrice int64) error {
	if gasPrice < 100 {
		panic("gas_price error")
	}

	for k, c := range h.cost {
		fee := gasPrice * c.ToGas()
		if fee != 0 {
			gas := &common.Fixed{
				Value:   fee * 1000000,
				Decimal: 8, // TODO magic number
			}
			err := h.h.CostGas(k, gas)
			if err != nil {
				return fmt.Errorf("pay cost failed: %v, %v", k, err)
			}
		}
		//ram := c.Data // todo activate ram
		//currentRam := h.h.db.TokenBalance("ram", k)
		//h.h.db.SetTokenBalance("ram", k, currentRam+ram)
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
