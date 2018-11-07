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

// TransferRaw ...
func (h *Teller) TransferRaw(from, to string, amount int64) error {
	srcBalance := h.h.db.Balance(from)
	if strings.HasPrefix(from, ContractAccountPrefix) && srcBalance >= amount || srcBalance > amount {
		h.h.db.SetBalance(from, -1*amount)
		h.h.db.SetBalance(to, amount)
		return nil
	}
	return ErrBalanceNotEnough
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

// GetBalance return balance of an id
func (h *Teller) GetBalance(from string) (string, *contract.Cost, error) {
	var bl int64
	if strings.HasPrefix(from, "IOST") {
		bl = h.h.db.Balance(from)
	} else {
		bl = h.h.db.Balance(ContractAccountPrefix + from)
	}
	fpn := common.Fixed{Value: bl, Decimal: 8}
	return fpn.ToString(), GetCost, nil
}

// GrantCoin issue coin
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

// ConsumeCoin consume coin from
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

// GrantServi ...
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

// ConsumeServi ...
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

// TotalServi ...
func (h *Teller) TotalServi() (ts string, cost *contract.Cost) {
	fpn := common.Fixed{Value: h.h.db.TotalServi(), Decimal: 8}
	ts = fpn.ToString()
	cost = GetCost
	return
}

// Transfer ...
func (h *Teller) Transfer(from, to string, amountStr string) (*contract.Cost, error) {
	amount, _ := common.NewFixed(amountStr, 8)
	if amount.Value <= 0 {
		return CommonErrorCost(1), ErrTransferNegValue
	}

	if strings.HasPrefix(from, ContractAccountPrefix) {
		if from != ContractAccountPrefix+h.h.ctx.Value("contract_name").(string) {
			return CommonErrorCost(2), ErrPermissionLost
		}
	} else {
		if h.Privilege(from) < 1 {
			return CommonErrorCost(2), ErrPermissionLost
		}
	}

	err := h.TransferRaw(from, to, amount.Value)
	return TransferCost, err
}

// Withdraw ...
func (h *Teller) Withdraw(to string, amountStr string) (*contract.Cost, error) {
	c := h.h.ctx.Value("contract_name").(string)
	return h.Transfer(ContractAccountPrefix+c, to, amountStr)
}

// Deposit ...
func (h *Teller) Deposit(from string, amountStr string) (*contract.Cost, error) {
	c := h.h.ctx.Value("contract_name").(string)
	return h.Transfer(from, ContractAccountPrefix+c, amountStr)

}

// TopUp ...
func (h *Teller) TopUp(c, from string, amountStr string) (*contract.Cost, error) {
	return h.Transfer(from, ContractGasPrefix+c, amountStr)
}

// Countermand ...
func (h *Teller) Countermand(c, to string, amountStr string) (*contract.Cost, error) {
	amount, _ := common.NewFixed(amountStr, 8)
	return TransferCost, h.TransferRaw(ContractGasPrefix+c, to, amount.Value)
}

func (h *Teller) Costs() map[string]*contract.Cost {
	return h.cost
}

// PayCost ...
func (h *Teller) PayCost(c *contract.Cost, who string) {
	h.cost[who] = c
}

// DoPay ...
func (h *Teller) DoPay(witness string, gasPrice int64, isPayRAM bool) error {
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
		if isPayRAM {
			ram := c.Data
			currentRam := h.h.db.TokenBalance("ram", k)
			if currentRam-ram < 0 {
				return fmt.Errorf("pay ram failed. id: %v need %v, actual %v", k, ram, currentRam)
			}
			h.h.db.SetTokenBalance("ram", k, currentRam-ram)
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
