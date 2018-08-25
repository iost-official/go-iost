package host

import (
	"strings"

	"github.com/iost-official/Go-IOS-Protocol/core/contract"
	"github.com/iost-official/Go-IOS-Protocol/ilog"
)

// const ...
const (
	ContractAccountPrefix = "CA"
	ContractGasPrefix     = "CG"
)

// Teller ...
type Teller struct {
	h    *Host
	cost map[string]*contract.Cost
}

// NewTeller ...
func NewTeller(h *Host) Teller {
	return Teller{
		h:    h,
		cost: make(map[string]*contract.Cost),
	}
}

func (h *Teller) transfer(from, to string, amount int64) error {
	bf := h.h.db.Balance(from)
	//ilog.Debugf("%v's balance : %v", from, bf)
	if bf > amount {
		h.h.db.SetBalance(from, -1*amount)
		h.h.db.SetBalance(to, amount)
		return nil
	}
	return ErrBalanceNotEnough
}

// GrantCoin ...
func (h *Teller) GrantCoin(coinName, to string, amount int64) (*contract.Cost, error) {
	if amount <= 0 {
		return CommonErrorCost(1), ErrTransferNegValue
	}
	h.h.db.SetCoin(coinName, to, amount)
	return TransferCost, nil
}

// ConsumeCoin ...
func (h *Teller) ConsumeCoin(coinName, from string, amount int64) (cost *contract.Cost, err error) {
	if amount <= 0 {
		return CommonErrorCost(1), ErrTransferNegValue
	}
	bl := h.h.db.Coin(coinName, from)
	if bl < amount {
		return CommonErrorCost(1), ErrBalanceNotEnough
	}
	h.h.db.SetCoin(coinName, from, -1*amount)
	return TransferCost, nil
}

// Transfer ...
func (h *Teller) Transfer(from, to string, amount int64) (*contract.Cost, error) {
	//ilog.Debugf("amount : %v", amount)
	if amount <= 0 {
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

	err := h.transfer(from, to, amount)
	return TransferCost, err
}

// Withdraw ...
func (h *Teller) Withdraw(to string, amount int64) (*contract.Cost, error) {
	c := h.h.ctx.Value("contract_name").(string)
	return TransferCost, h.transfer(ContractAccountPrefix+c, to, amount)
}

// Deposit ...
func (h *Teller) Deposit(from string, amount int64) (*contract.Cost, error) {
	c := h.h.ctx.Value("contract_name").(string)
	return TransferCost, h.transfer(from, ContractAccountPrefix+c, amount)

}

// TopUp ...
func (h *Teller) TopUp(c, from string, amount int64) (*contract.Cost, error) {
	return TransferCost, h.transfer(from, ContractGasPrefix+c, amount)

}

// Countermand ...
func (h *Teller) Countermand(c, to string, amount int64) (*contract.Cost, error) {
	return TransferCost, h.transfer(ContractGasPrefix+c, to, amount)
}

// PayCost ...
func (h *Teller) PayCost(c *contract.Cost, who string) {
	h.cost[who] = c
}

// DoPay ...
func (h *Teller) DoPay(witness string, gasPrice int64) error {
	if gasPrice <= 0 {
		panic("gas_price error")
	}

	for k, c := range h.cost {
		fee := gasPrice * c.ToGas()
		if fee == 0 {
			continue
		}
		if strings.HasPrefix(k, "IOST") {
			err := h.transfer(k, witness, fee)
			if err != nil {
				return err
			}
		} else if strings.HasPrefix(k, ContractGasPrefix) {
			err := h.transfer(k, witness, fee)
			if err != nil {
				return err
			}
		} else {
			ilog.Errorf("key is:", k)
			panic("prefix error")
		}
	}

	return nil
}

// Privilege ...
func (h *Teller) Privilege(id string) int {
	am := h.h.ctx.Value("auth_list").(map[string]int)
	i, ok := am[id]
	if !ok {
		i = 0
	}
	return i
}
