package host

import (
	"strings"

	"github.com/iost-official/Go-IOS-Protocol/core/contract"
	"github.com/iost-official/Go-IOS-Protocol/ilog"
	"github.com/iost-official/Go-IOS-Protocol/vm/database"
)

const (
	ContractAccountPrefix = "CA"
	ContractGasPrefix     = "CG"
)

type Teller struct {
	db   *database.Visitor
	ctx  *Context
	cost map[string]*contract.Cost
}

func NewTeller(db *database.Visitor, ctx *Context) Teller {
	return Teller{
		db:   db,
		ctx:  ctx,
		cost: make(map[string]*contract.Cost),
	}
}

func (h *Teller) transfer(from, to string, amount int64) error {
	bf := h.db.Balance(from)
	//ilog.Debug("%v's balance : %v", from, bf)
	if bf > amount {
		h.db.SetBalance(from, -1*amount)
		h.db.SetBalance(to, amount)
		return nil
	}
	return ErrBalanceNotEnough
}

func (h *Teller) Transfer(from, to string, amount int64) (*contract.Cost, error) {
	//ilog.Debug("amount : %v", amount)
	if amount <= 0 {
		return contract.NewCost(1, 1, 1), ErrTransferNegValue
	}

	if strings.HasPrefix(from, ContractAccountPrefix) {
		if from != ContractAccountPrefix+h.ctx.Value("contract_name").(string) {
			return contract.NewCost(1, 1, 1), ErrPermissionLost
		}
	} else {
		if h.Privilege(from) < 1 {
			return contract.NewCost(1, 1, 1), ErrPermissionLost
		}
	}

	err := h.transfer(from, to, amount)
	return contract.NewCost(1, 1, 1), err
}

func (h *Teller) Withdraw(to string, amount int64) (*contract.Cost, error) {
	c := h.ctx.Value("contract_name").(string)
	return contract.NewCost(1, 1, 1), h.transfer(ContractAccountPrefix+c, to, amount)
}

func (h *Teller) Deposit(from string, amount int64) (*contract.Cost, error) {
	c := h.ctx.Value("contract_name").(string)
	return contract.NewCost(1, 1, 1), h.transfer(from, ContractAccountPrefix+c, amount)
}

func (h *Teller) TopUp(c, from string, amount int64) (*contract.Cost, error) {
	return contract.NewCost(1, 1, 1), h.transfer(from, ContractGasPrefix+c, amount)

}

func (h *Teller) Countermand(c, to string, amount int64) (*contract.Cost, error) {
	return contract.NewCost(1, 1, 1), h.transfer(ContractGasPrefix+c, to, amount)
}

func (h *Teller) PayCost(c *contract.Cost, who string) {
	h.cost[who] = c
}

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
			ilog.Error("key is:", k)
			panic("prefix error")
		}
	}

	return nil
}

func (h *Teller) Privilege(id string) int {
	am := h.ctx.Value("auth_list").(map[string]int)
	i, ok := am[id]
	if !ok {
		i = 0
	}
	return i
}
