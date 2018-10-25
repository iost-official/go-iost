package host

import (
	"strings"

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

// GetBalance return balance of an id
func (h *Teller) GetBalance(from string) (string, *contract.Cost, error) {
	var bl int64
	if strings.HasPrefix(from, "IOST") {
		bl = h.h.db.Balance(from)
	} else {
		bl = h.h.db.Balance(ContractAccountPrefix + from)
	}
	fpn := FixPointNumber{value: bl, decimal: 8}
	return fpn.ToString(), GetCost, nil
}

// GrantCoin issue coin
func (h *Teller) GrantCoin(coinName, to string, amountStr string) (*contract.Cost, error) {
	amount, _ := NewFixPointNumber(amountStr, 8)
	if amount.value <= 0 {
		return CommonErrorCost(1), ErrTransferNegValue
	}
	cn := h.h.ctx.Value("contract_name").(string)
	if !strings.HasPrefix(cn, "iost.") {
		return CommonErrorCost(2), ErrPermissionLost
	}
	h.h.db.SetCoin(coinName, to, amount.value)
	return TransferCost, nil
}

// ConsumeCoin consume coin from
func (h *Teller) ConsumeCoin(coinName, from string, amountStr string) (cost *contract.Cost, err error) {
	amount, _ := NewFixPointNumber(amountStr, 8)
	if amount.value <= 0 {
		return CommonErrorCost(1), ErrTransferNegValue
	}
	if h.Privilege(from) < 1 {
		return CommonErrorCost(1), ErrPermissionLost
	}
	bl := h.h.db.Coin(coinName, from)
	if bl < amount.value {
		return CommonErrorCost(2), ErrBalanceNotEnough
	}
	h.h.db.SetCoin(coinName, from, -1*amount.value)
	return TransferCost, nil
}

// GrantServi ...
func (h *Teller) GrantServi(to string, amountStr string) (*contract.Cost, error) {
	amount, _ := NewFixPointNumber(amountStr, 8)
	if amount.value <= 0 {
		return CommonErrorCost(1), ErrTransferNegValue
	}
	//cn := h.h.ctx.Value("contract_name").(string) todo privilege of system contracts
	//if !strings.HasPrefix(cn, "iost.") {
	//	return CommonErrorCost(2), ErrPermissionLost
	//}
	h.h.db.SetServi(to, amount.value)
	return TransferCost, nil
}

// ConsumeServi ...
func (h *Teller) ConsumeServi(from string, amountStr string) (cost *contract.Cost, err error) {
	amount, _ := NewFixPointNumber(amountStr, 8)
	if amount.value <= 0 {
		return CommonErrorCost(1), ErrTransferNegValue
	}
	if h.Privilege(from) < 1 {
		return CommonErrorCost(1), ErrPermissionLost
	}
	bl := h.h.db.Servi(from)
	if bl < amount.value {
		return CommonErrorCost(2), ErrBalanceNotEnough
	}
	h.h.db.SetServi(from, -1*amount.value)
	return TransferCost, nil
}

// TotalServi ...
func (h *Teller) TotalServi() (ts string, cost *contract.Cost) {
	fpn := FixPointNumber{value: h.h.db.TotalServi(), decimal: 8}
	ts = fpn.ToString()
	cost = GetCost
	return
}

// Transfer ...
func (h *Teller) Transfer(from, to string, amountStr string) (*contract.Cost, error) {
	//ilog.Debugf("amount : %v", amount)
	amount, _ := NewFixPointNumber(amountStr, 8)
	if amount.value <= 0 {
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

	err := h.transfer(from, to, amount.value)
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
	amount, _ := NewFixPointNumber(amountStr, 8)
	return TransferCost, h.transfer(ContractGasPrefix+c, to, amount.value)
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
			// 10% of gas transferred to iost.bonus
			err = h.transfer(k, ContractAccountPrefix+"iost.bonus", bfee)
			if err != nil {
				return err
			}
		} else if strings.HasPrefix(k, ContractGasPrefix) {
			err := h.transfer(k, witness, fee-bfee)
			if err != nil {
				return err
			}
			// 10% of gas transferred to iost.bonus
			err = h.transfer(k, ContractAccountPrefix+"iost.bonus", bfee)
			if err != nil {
				return err
			}
		} else {
			ilog.Errorf("key is: %v", k)
			panic("prefix error")
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

type FixPointNumber struct {
	value   int64
	decimal int
}

func NewFixPointNumber(amount string, decimal int) (*FixPointNumber, bool) {
	fpn := &FixPointNumber{value: 0, decimal: decimal}
	if len(amount) == 0 || amount[0] == '.' {
		return nil, false
	}
	i := 0
	for ; i < len(amount); i++ {
		if '0' <= amount[i] && amount[i] <= '9' {
			fpn.value = fpn.value*10 + int64(amount[i]-'0')
		} else if amount[i] == '.' {
			break
		} else {
			return nil, false
		}
	}
	for i = i + 1; i < len(amount) && decimal > 0; i++ {
		if '0' <= amount[i] && amount[i] <= '9' {
			fpn.value = fpn.value*10 + int64(amount[i]-'0')
			decimal = decimal - 1
		} else {
			return nil, false
		}
	}
	for decimal > 0 {
		fpn.value = fpn.value * 10
		decimal = decimal - 1
	}
	return fpn, true
}

func (fpn *FixPointNumber) ToString() string {
	val := fpn.value
	str := make([]byte, 0, 0)
	for val > 0 || len(str) <= fpn.decimal {
		str = append(str, byte('0'+val%10))
		val /= 10
	}
	rtn := make([]byte, 0, 0)
	for i := len(str) - 1; i >= 0; i-- {
		if i+1 == fpn.decimal {
			rtn = append(rtn, '.')
		}
		rtn = append(rtn, str[i])
	}
	for rtn[len(rtn)-1] == '0' {
		rtn = rtn[0 : len(rtn)-1]
	}
	return string(rtn)
}
