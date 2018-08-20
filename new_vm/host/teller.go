package host

import (
	"strings"

	"github.com/iost-official/Go-IOS-Protocol/core/contract"
	"github.com/iost-official/Go-IOS-Protocol/ilog"
	"github.com/iost-official/Go-IOS-Protocol/new_vm/database"
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

func (h *Teller) Transfer(from, to string, amount int64) (*contract.Cost, error) {
	//ilog.Debug("amount : %v", amount)
	if amount <= 0 {
		return contract.NewCost(1, 1, 1), ErrTransferNegValue
	}

	bf := h.db.Balance(from)
	//ilog.Debug("%v's balance : %v", from, bf)
	if bf > amount {
		h.db.SetBalance(from, -1*amount)
		h.db.SetBalance(to, amount)
	} else {
		return contract.NewCost(1, 1, 1), ErrBalanceNotEnough
	}
	return contract.NewCost(1, 1, 1), nil
}

func (h *Teller) Withdraw(to string, amount int64) (*contract.Cost, error) {
	c := h.ctx.Value(ContractAccountPrefix + "contract_name").(string)
	return h.Transfer(c, to, amount)
}

func (h *Teller) Deposit(from string, amount int64) (*contract.Cost, error) {
	c := h.ctx.Value(ContractAccountPrefix + "contract_name").(string)
	return h.Transfer(from, c, amount)
}

func (h *Teller) TopUp(contract, from string, amount int64) (*contract.Cost, error) {
	return h.Transfer(from, ContractGasPrefix+contract, amount)
}

func (h *Teller) Countermand(contract, to string, amount int64) (*contract.Cost, error) {
	return h.Transfer(ContractGasPrefix+contract, to, amount)
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
			_, err := h.Transfer(k, witness, int64(fee))
			if err != nil {
				return err
			}
		} else if strings.HasPrefix(k, ContractGasPrefix) {
			_, err := h.Transfer(k, witness, int64(fee))
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
