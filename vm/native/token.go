package native

import (
	"github.com/iost-official/go-iost/vm/host"
	"github.com/iost-official/go-iost/core/contract"
	"errors"
	"strconv"
)

var tokenABIs map[string]*abi

// const prefix
const (
	TokenInfoMapPrefix = "TI"
	TokenBalanceMapPrefix = "TB"
	IssuerMapField = "issuer"
	SupplyMapField = "supply"
	TotalSupplyMapField = "totalSupply"
	CanTransferMapField = "canTransfer"
	DefaultRateMapField = "defaultRate"
	DecimalMapField = "decimal"
)

func init() {
	tokenABIs = make(map[string]*abi)
	register(&tokenABIs, createTokenABI)
	register(&tokenABIs, issueTokenABI)
	register(&tokenABIs, transferTokenABI)
	register(&tokenABIs, balanceOfTokenABI)
	register(&tokenABIs, getSupplyTokenABI)
	register(&tokenABIs, getTotalSupplyTokenABI)
}

func checkTokenExists(h *host.Host, tokenName string) (ok bool, cost *contract.Cost) {
	exists, cost0 := h.MapHas(TokenInfoMapPrefix + tokenName, IssuerMapField)
	return exists, cost0
}

func getBalance(h *host.Host, tokenName string, from string) (balance int64, cost *contract.Cost) {
	balance = int64(0)
	cost = contract.Cost0()
	ok, cost0 := h.MapHas(TokenBalanceMapPrefix + from, tokenName)
	cost.AddAssign(cost0)
	if ok {
		tmp, cost0 := h.MapGet(TokenBalanceMapPrefix + from, tokenName)
		cost.AddAssign(cost0)
		balance = tmp.(int64)
	}
	return balance, cost
}

func setBalance(h *host.Host, tokenName string, from string, balance int64) (cost *contract.Cost) {
	cost = h.MapPut(TokenBalanceMapPrefix + from, tokenName, balance)
	return cost
}

var (
	createTokenABI = &abi{
		name: "create",
		args: []string{"string", "string", "number", "json"},
		do: func(h *host.Host, args ...interface{}) (rtn []interface{}, cost *contract.Cost, err error) {
			cost = contract.Cost0()
			cost.AddAssign(host.CommonOpCost(1))
			tokenName := args[0].(string)
			issuer := args[1].(string)
			totalSupply := args[2].(int64)
			// todo config

			// check auth
			ok, cost0 := h.RequireAuth(issuer)
			cost.AddAssign(cost0)
			if !ok {
				return nil, cost, host.ErrPermissionLost
			}

			// check exists
			ok, cost0 = checkTokenExists(h, tokenName)
			cost.AddAssign(cost0)
			if ok {
				return nil, cost, host.ErrTokenExists
			}

			// check valid
			decimal := 8
			if decimal >= 19 {
				return nil, cost, errors.New("invalid decimal")
			}

			// put table
			cost0 = h.MapPut(TokenInfoMapPrefix + tokenName, IssuerMapField, issuer)
			cost.AddAssign(cost0)
			cost0 = h.MapPut(TokenInfoMapPrefix + tokenName, TotalSupplyMapField, totalSupply)
			cost.AddAssign(cost0)
			cost0 = h.MapPut(TokenInfoMapPrefix + tokenName, SupplyMapField, 0)
			cost.AddAssign(cost0)
			cost0 = h.MapPut(TokenInfoMapPrefix + tokenName, CanTransferMapField, true)
			cost.AddAssign(cost0)
			cost0 = h.MapPut(TokenInfoMapPrefix + tokenName, DefaultRateMapField, 1.0)
			cost.AddAssign(cost0)
			cost0 = h.MapPut(TokenInfoMapPrefix + tokenName, DecimalMapField, decimal)
			cost.AddAssign(cost0)

			return []interface{}{}, cost, nil
		},
	}

	issueTokenABI = &abi{
		name: "issue",
		args: []string{"string", "string", "string"},
		do: func(h *host.Host, args ...interface{}) (rtn []interface{}, cost *contract.Cost, err error) {
			cost = contract.Cost0()
			cost.AddAssign(host.CommonOpCost(1))
			tokenName := args[0].(string)
			to := args[1].(string)
			amountStr := args[2].(string)

			// get token info
			ok, cost0 := checkTokenExists(h, tokenName)
			cost.AddAssign(cost0)
			if !ok {
				return nil, cost, host.ErrTokenNotExists
			}
			issuer, cost0 := h.MapGet(TokenInfoMapPrefix + tokenName, IssuerMapField)
			cost.AddAssign(cost0)
			supply, cost0 := h.MapGet(TokenInfoMapPrefix + tokenName, SupplyMapField)
			cost.AddAssign(cost0)
			totalSupply, cost0 := h.MapGet(TokenInfoMapPrefix + tokenName, TotalSupplyMapField)
			cost.AddAssign(cost0)

			// check auth
			ok, cost0 = h.RequireAuth(issuer.(string))
			cost.AddAssign(cost0)
			if !ok {
				return nil, cost, host.ErrPermissionLost
			}

			// todo get amount
			amount, err := strconv.ParseInt(amountStr, 10, 64)
			cost.AddAssign(host.CommonOpCost(1))
			if err != nil {
				return nil, cost, err
			}
			if amount <= 0 {
				return nil, cost, host.ErrInvalidAmount
			}

			// check supply
			if totalSupply.(int64) - supply.(int64) < amount {
				return nil, cost, errors.New("supply to much")
			}

			// put table
			cost0 = h.MapPut(TokenInfoMapPrefix + tokenName, SupplyMapField, supply.(int64) + amount)
			cost.AddAssign(cost0)

			balance, cost0 := getBalance(h, tokenName, to)
			cost.AddAssign(cost0)

			balance += amount
			cost0 = setBalance(h, tokenName, to, balance)
			cost.AddAssign(cost0)

			return []interface{}{}, cost, nil
		},
	}

	transferTokenABI = &abi{
		name: "transfer",
		args: []string{"string", "string", "string", "string"},
		do: func(h *host.Host, args ...interface{}) (rtn []interface{}, cost *contract.Cost, err error) {
			cost = contract.Cost0()
			cost.AddAssign(host.CommonOpCost(1))
			tokenName := args[0].(string)
			from := args[1].(string)
			to := args[1].(string)
			amountStr := args[1].(string)

			// get token info
			ok, cost0 := checkTokenExists(h, tokenName)
			cost.AddAssign(cost0)
			if !ok {
				return nil, cost, host.ErrTokenNotExists
			}

			// check auth
			ok, cost0 = h.RequireAuth(from)
			cost.AddAssign(cost0)
			if !ok {
				return nil, cost, host.ErrPermissionLost
			}

			// todo get amount
			amount, err := strconv.ParseInt(amountStr, 10, 64)
			cost.AddAssign(host.CommonOpCost(1))
			if err != nil {
				return nil, cost, err
			}
			if amount <= 0 {
				return nil, cost, host.ErrInvalidAmount
			}

			// put table
			fbalance, cost0 := getBalance(h, tokenName, from)
			cost.AddAssign(cost0)
			tbalance, cost0 := getBalance(h, tokenName, to)
			cost.AddAssign(cost0)
			if fbalance < amount {
				return nil, cost, host.ErrBalanceNotEnough
			}

			fbalance -= amount
			tbalance += amount

			cost0 = setBalance(h, tokenName, from, fbalance)
			cost.AddAssign(cost0)
			cost0 = setBalance(h, tokenName, to, tbalance)
			cost.AddAssign(cost0)

			return []interface{}{}, cost, nil
		},
	}

	balanceOfTokenABI = &abi{
		name: "balanceOf",
		args: []string{"string", "string"},
		do: func(h *host.Host, args ...interface{}) (rtn []interface{}, cost *contract.Cost, err error) {
			cost = contract.Cost0()
			cost.AddAssign(host.CommonOpCost(1))
			tokenName := args[0].(string)
			to := args[1].(string)

			// check token info
			ok, cost0 := checkTokenExists(h, tokenName)
			cost.AddAssign(cost0)
			if !ok {
				return nil, cost, host.ErrTokenNotExists
			}

			balance, cost0 := getBalance(h, tokenName, to)
			cost.AddAssign(cost0)

			return []interface{}{balance}, cost, nil
		},
	}

	getSupplyTokenABI = &abi{
		name: "getSupply",
		args: []string{"string"},
		do: func(h *host.Host, args ...interface{}) (rtn []interface{}, cost *contract.Cost, err error) {
			cost = contract.Cost0()
			cost.AddAssign(host.CommonOpCost(1))
			tokenName := args[0].(string)

			// check token info
			ok, cost0 := checkTokenExists(h, tokenName)
			cost.AddAssign(cost0)
			if !ok {
				return nil, cost, host.ErrTokenNotExists
			}

			supply, cost0 := h.MapGet(TokenInfoMapPrefix + tokenName, SupplyMapField)
			cost.AddAssign(cost0)

			return []interface{}{supply.(int64)}, cost, nil
		},
	}

	getTotalSupplyTokenABI = &abi{
		name: "getTotalSupply",
		args: []string{"string"},
		do: func(h *host.Host, args ...interface{}) (rtn []interface{}, cost *contract.Cost, err error) {
			cost = contract.Cost0()
			cost.AddAssign(host.CommonOpCost(1))
			tokenName := args[0].(string)

			// check token info
			ok, cost0 := checkTokenExists(h, tokenName)
			cost.AddAssign(cost0)
			if !ok {
				return nil, cost, host.ErrTokenNotExists
			}

			totalSupply, cost0 := h.MapGet(TokenInfoMapPrefix + tokenName, TotalSupplyMapField)
			cost.AddAssign(cost0)

			return []interface{}{totalSupply.(int64)}, cost, nil
		},
	}
)
