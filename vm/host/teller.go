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
	if strings.HasPrefix(from, ContractAccountPrefix) && bf >= amount || bf > amount {
		h.h.db.SetBalance(from, -1*amount)
		h.h.db.SetBalance(to, amount)
		return nil
	}
	return ErrBalanceNotEnough
}

func (h *Teller) GetBalance(from string) (int64, *contract.Cost, error) {
	bl := int64(0)
	if strings.HasPrefix(from, "IOST") {
		bl = h.h.db.Balance(from)
	} else {
		bl = h.h.db.Balance(ContractAccountPrefix + from)
	}
	return bl, GetCost, nil
}

// GrantCoin ...
func (h *Teller) GrantCoin(coinName, to string, amount int64) (*contract.Cost, error) {
	if amount <= 0 {
		return CommonErrorCost(1), ErrTransferNegValue
	}
	cn := h.h.ctx.Value("contract_name").(string)
	if !strings.HasPrefix(cn, "iost.") {
		return CommonErrorCost(2), ErrPermissionLost
	}
	h.h.db.SetCoin(coinName, to, amount)
	return TransferCost, nil
}

// ConsumeCoin ...
func (h *Teller) ConsumeCoin(coinName, from string, amount int64) (cost *contract.Cost, err error) {
	if amount <= 0 {
		return CommonErrorCost(1), ErrTransferNegValue
	}
	if h.Privilege(from) < 1 {
		return CommonErrorCost(1), ErrPermissionLost
	}
	bl := h.h.db.Coin(coinName, from)
	if bl < amount {
		return CommonErrorCost(2), ErrBalanceNotEnough
	}
	h.h.db.SetCoin(coinName, from, -1*amount)
	return TransferCost, nil
}

// GrantServi ...
func (h *Teller) GrantServi(to string, amount int64) (*contract.Cost, error) {
	if amount <= 0 {
		return CommonErrorCost(1), ErrTransferNegValue
	}
	cn := h.h.ctx.Value("contract_name").(string)
	if !strings.HasPrefix(cn, "iost.") {
		return CommonErrorCost(2), ErrPermissionLost
	}
	h.h.db.SetServi(to, amount)
	return TransferCost, nil
}

// ConsumeServi ...
func (h *Teller) ConsumeServi(from string, amount int64) (cost *contract.Cost, err error) {
	if amount <= 0 {
		return CommonErrorCost(1), ErrTransferNegValue
	}
	if h.Privilege(from) < 1 {
		return CommonErrorCost(1), ErrPermissionLost
	}
	bl := h.h.db.Servi(from)
	if bl < amount {
		return CommonErrorCost(2), ErrBalanceNotEnough
	}
	h.h.db.SetServi(from, -1*amount)
	return TransferCost, nil
}

// TotalServi ...
func (h *Teller) TotalServi() (ts int64, cost *contract.Cost) {
	ts = h.h.db.TotalServi()
	cost = GetCost
	return
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
	return h.Transfer(ContractAccountPrefix+c, to, amount)
}

// Deposit ...
func (h *Teller) Deposit(from string, amount int64) (*contract.Cost, error) {
	c := h.h.ctx.Value("contract_name").(string)
	return h.Transfer(from, ContractAccountPrefix+c, amount)

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
	if gasPrice < 0 {
		panic("gas_price error")
	}

	for k, c := range h.cost {
		fee := gasPrice * c.ToGas()
		if fee == 0 {
			continue
		}
		bfee := fee / 10
		if strings.HasPrefix(k, "IOST") {
			err := h.transfer(k, witness, fee-bfee)
			if err != nil {
				return err
			}
			// 10% of gas transfered to iost.bonus
			err = h.transfer(k, ContractAccountPrefix+"iost.bonus", bfee)
			if err != nil {
				return err
			}
		} else if strings.HasPrefix(k, ContractGasPrefix) {
			err := h.transfer(k, witness, fee-bfee)
			if err != nil {
				return err
			}
			// 10% of gas transfered to iost.bonus
			err = h.transfer(k, ContractAccountPrefix+"iost.bonus", bfee)
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
