package native

import (
	"fmt"

	"github.com/iost-official/go-iost/core/contract"
	"github.com/iost-official/go-iost/vm/host"
)

var coinABIs map[string]*abi

// const prifix
const (
	CoinContractPrefix = "OC"
	CoinRatePrefix     = "OR"
)

func init() {
	coinABIs = make(map[string]*abi)
	register(coinABIs, createCoin)
	register(coinABIs, issueCoin)
	register(coinABIs, setCoinRate)
}

var (
	createCoin = &abi{
		name: "CreateCoin",
		args: []string{"string", "string"},
		do: func(h *host.Host, args ...interface{}) (rtn []interface{}, cost *contract.Cost, err error) {
			cost = contract.Cost0()
			coinName := args[0].(string)
			coinContract := args[1].(string)
			val, cost0 := h.Get(CoinContractPrefix + coinName)
			cost.AddAssign(cost0)
			if val != nil {
				return nil, cost, host.ErrCoinExists
			}
			cost0 = h.Put(CoinContractPrefix+coinName, coinContract)
			cost.AddAssign(cost0)
			cost0 = h.Put(CoinRatePrefix+coinName, 1)
			cost.AddAssign(cost0)
			return []interface{}{}, cost, nil
		},
	}

	issueCoin = &abi{
		name: "IssueCoin",
		args: []string{"string", "string", "string"},
		do: func(h *host.Host, args ...interface{}) (rtn []interface{}, cost *contract.Cost, err error) {
			cost = contract.Cost0()
			coinName := args[0].(string)
			account := args[1].(string)
			amount := args[2].(string)

			// get coin contract
			coinContract, cost0 := h.Get(CoinContractPrefix + coinName)
			cost.AddAssign(cost0)
			if coinContract == nil {
				return nil, cost, host.ErrCoinNotExists
			}

			// check can_issue
			rtn, cost0, err = h.Call(coinContract.(string), "can_issue", fmt.Sprintf(`["%v","%s"]`, account, amount))
			cost.AddAssign(cost0)
			if err != nil {
				return nil, cost, err
			}
			if len(rtn) < 1 || rtn[0].(bool) != true {
				return nil, cost, host.ErrCoinIssueRefused
			}

			cost0, err = h.GrantCoin(coinName, account, amount)
			cost.AddAssign(cost0)

			return []interface{}{}, cost, err
		},
	}

	setCoinRate = &abi{
		name: "SetCoinRate",
		args: []string{"string", "number"},
		do: func(h *host.Host, args ...interface{}) (rtn []interface{}, cost *contract.Cost, err error) {
			cost = contract.Cost0()
			coinName := args[0].(string)
			// 1 iost = rate * coin, rate should be a fixed-point number
			rate := args[1].(int64)

			// get coin contract
			coinContract, cost0 := h.Get(CoinContractPrefix + coinName)
			cost.AddAssign(cost0)

			// check can_issue
			rtn, cost0, err = h.Call(coinContract.(string), "can_setrate", fmt.Sprintf(`[%v]`, rate))
			cost.AddAssign(cost0)
			if err != nil {
				return nil, cost, err
			}
			if len(rtn) < 1 || rtn[0].(bool) != true {
				return nil, cost, host.ErrCoinSetRateRefused
			}

			cost0 = h.Put(CoinRatePrefix+coinName, rate)
			cost.AddAssign(cost0)

			return []interface{}{}, cost, nil
		},
	}
)
