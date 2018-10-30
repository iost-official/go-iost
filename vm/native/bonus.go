package native

import (
	"github.com/iost-official/go-iost/core/contract"
	"github.com/iost-official/go-iost/vm/host"
)

var bonusABIs map[string]*abi

func init() {
	bonusABIs = make(map[string]*abi)
	register(&bonusABIs, claimBonus)
	register(&bonusABIs, constructor)
	register(&bonusABIs, initFunc)
}

var (
	constructor = &abi{
		name: "constructor",
		args: []string{},
		do: func(h *host.Host, args ...interface{}) (rtn []interface{}, cost *contract.Cost, err error) {
			return []interface{}{}, host.CommonErrorCost(1), nil
		},
	}
	initFunc = &abi{
		name: "init",
		args: []string{},
		do: func(h *host.Host, args ...interface{}) (rtn []interface{}, cost *contract.Cost, err error) {
			return []interface{}{}, host.CommonErrorCost(1), nil
		},
	}
	claimBonus = &abi{
		name: "ClaimBonus",
		args: []string{"string", "string"},
		do: func(h *host.Host, args ...interface{}) (rtn []interface{}, cost *contract.Cost, err error) {
			cost = contract.Cost0()
			acc := args[0].(string)
			amount := args[1].(string)

			ok, cost0 := h.RequireAuth(acc, "active")
			cost.AddAssign(cost0)
			if !ok {
				return nil, cost, host.ErrPermissionLost
			}

			_, cost0 = h.TotalServi()
			cost.AddAssign(cost0)

			cost0, err = h.ConsumeServi(acc, amount)
			cost.AddAssign(cost0)
			if err != nil {
				return nil, cost, err
			}

			_, cost0, err = h.GetBalance("iost.bonus")
			cost.AddAssign(cost0)
			if err != nil {
				return nil, cost, err
			}
			/*
				token := amount * 1.0 / totalServi * bl
				if token > bl {
					token = bl
				}
				if token <= 0 {
					return []interface{}{}, cost, nil
				}

				cost0, err = h.Withdraw(acc, token)
				cost.AddAssign(cost0)

			*/
			return []interface{}{}, cost, err
		},
	}
)
