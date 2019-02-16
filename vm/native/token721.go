package native

import (
	"encoding/json"
	"errors"
	"math"
	"strconv"

	"fmt"

	"github.com/iost-official/go-iost/core/contract"
	"github.com/iost-official/go-iost/vm/host"
)

var token721ABIs *abiSet

// const prefix
const (
	Token721InfoMapPrefix        = "T721I"
	Token721BalanceMapPrefix     = "T721B"
	Token721IssuerMapField       = "T721issuer"
	Token721MetadataMapPrefix    = "T721M"
	Token721MetadataKeySeparator = "#"
)

func init() {
	token721ABIs = newAbiSet()
	token721ABIs.Register(initToken721ABI, true)
	token721ABIs.Register(createToken721ABI)
	token721ABIs.Register(issueToken721ABI)
	token721ABIs.Register(transferToken721ABI)
	token721ABIs.Register(balanceOfToken721ABI)
	token721ABIs.Register(ownerOfToken721ABI)
	token721ABIs.Register(tokenOfOwnerByIndexToken721ABI)
	token721ABIs.Register(tokenMetadataToken721ABI)
}

func checkToken721Exists(h *host.Host, tokenSym string) (ok bool, cost contract.Cost) {
	exists, cost0 := h.MapHas(Token721InfoMapPrefix+tokenSym, Token721IssuerMapField)
	return exists, cost0
}

func getToken721Balance(h *host.Host, tokenSym string, from string) (balance int64, cost contract.Cost, err error) {
	balance = int64(0)
	cost = contract.Cost0()
	ok, cost0 := h.MapHas(Token721BalanceMapPrefix+from, tokenSym)
	cost.AddAssign(cost0)
	if ok {
		tmp, cost0 := h.MapGet(Token721BalanceMapPrefix+from, tokenSym)
		cost.AddAssign(cost0)
		balance = tmp.(int64)
	}
	return balance, cost, nil
}

func setToken721Balance(h *host.Host, tokenSym string, from string, balance int64, ramPayer string) (cost contract.Cost) {
	cost, _ = h.MapPut(Token721BalanceMapPrefix+from, tokenSym, balance, ramPayer)
	return cost

}

var (
	initToken721ABI = &abi{
		name: "init",
		args: []string{},
		do: func(h *host.Host, args ...interface{}) (rtn []interface{}, cost contract.Cost, err error) {
			return []interface{}{}, host.CommonErrorCost(1), nil
		},
	}
	createToken721ABI = &abi{
		name: "create",
		args: []string{"string", "string", "number"},
		do: func(h *host.Host, args ...interface{}) (rtn []interface{}, cost contract.Cost, err error) {
			cost = contract.Cost0()
			cost.AddAssign(host.CommonOpCost(1))
			tokenSym := args[0].(string)
			issuer := args[1].(string)
			totalSupply := args[2].(int64)

			cost.AddAssign(host.CommonOpCost(1))
			err = checkTokenSymValid(tokenSym)
			if err != nil {
				return nil, cost, err
			}
			// check auth
			ok, cost0 := h.RequireAuth(issuer, TokenPermission)
			cost.AddAssign(cost0)
			if !ok {
				return nil, cost, host.ErrPermissionLost
			}
			if !CheckCost(h, cost) {
				return nil, cost, host.ErrOutOfGas
			}

			// check exists
			ok, cost0 = checkToken721Exists(h, tokenSym)
			cost.AddAssign(cost0)
			if ok {
				return nil, cost, host.ErrTokenExists
			}

			// check valid
			if totalSupply > math.MaxInt64 {
				return nil, cost, errors.New("invalid total supply")
			}

			publisher := h.Context().Value("publisher").(string)
			// put info
			cost0, _ = h.MapPut(Token721InfoMapPrefix+tokenSym, Token721IssuerMapField, issuer, publisher)
			cost.AddAssign(cost0)
			cost0, _ = h.MapPut(Token721InfoMapPrefix+tokenSym, TotalSupplyMapField, totalSupply, publisher)
			cost.AddAssign(cost0)
			cost0, _ = h.MapPut(Token721InfoMapPrefix+tokenSym, SupplyMapField, int64(0), publisher)
			cost.AddAssign(cost0)

			// generate receipt
			message, err := json.Marshal(args)
			cost.AddAssign(host.CommonOpCost(1))
			if err != nil {
				return nil, cost, err
			}
			cost0 = h.Receipt(string(message))
			cost.AddAssign(cost0)
			return []interface{}{}, cost, nil
		},
	}

	issueToken721ABI = &abi{
		name: "issue",
		args: []string{"string", "string", "string"},
		do: func(h *host.Host, args ...interface{}) (rtn []interface{}, cost contract.Cost, err error) {
			cost = contract.Cost0()
			cost.AddAssign(host.CommonOpCost(1))
			tokenSym := args[0].(string)
			to := args[1].(string)
			metaDataJSON := args[2].(string)

			// get token info
			ok, cost0 := checkToken721Exists(h, tokenSym)
			cost.AddAssign(cost0)
			if !ok {
				return nil, cost, host.ErrTokenNotExists
			}
			issuer, cost0 := h.MapGet(Token721InfoMapPrefix+tokenSym, Token721IssuerMapField)
			cost.AddAssign(cost0)
			supply, cost0 := h.MapGet(Token721InfoMapPrefix+tokenSym, SupplyMapField)
			cost.AddAssign(cost0)
			totalSupply, cost0 := h.MapGet(Token721InfoMapPrefix+tokenSym, TotalSupplyMapField)
			cost.AddAssign(cost0)
			if !CheckCost(h, cost) {
				return nil, cost, host.ErrOutOfGas
			}

			// check supply
			if totalSupply.(int64)-supply.(int64) <= 0 {
				return nil, cost, errors.New("supply too much")
			}

			tokenID := strconv.FormatInt(supply.(int64), 10)
			// check auth
			ok, cost0 = h.RequireAuth(issuer.(string), TokenPermission)
			cost.AddAssign(cost0)
			if !ok {
				return nil, cost, host.ErrPermissionLost
			}
			if !CheckCost(h, cost) {
				return nil, cost, host.ErrOutOfGas
			}

			publisher := h.Context().Value("publisher").(string)
			// set supply, set balance
			cost0, err = h.MapPut(Token721InfoMapPrefix+tokenSym, SupplyMapField, supply.(int64)+1, publisher)
			cost.AddAssign(cost0)
			if err != nil {
				return nil, cost, err
			}

			cost0, err = h.MapPut(Token721InfoMapPrefix+tokenSym, tokenID, to, publisher)
			cost.AddAssign(cost0)
			if err != nil {
				return nil, cost, err
			}

			tbalance, cost0, err := getToken721Balance(h, tokenSym, to)
			cost.AddAssign(cost0)
			if err != nil {
				return nil, cost, err
			}
			tbalance++
			cost0 = setToken721Balance(h, tokenSym, to, tbalance, publisher)
			cost.AddAssign(cost0)

			cost0, err = h.MapPut(Token721MetadataMapPrefix+tokenSym+Token721MetadataKeySeparator+to, tokenID, metaDataJSON, publisher)
			cost.AddAssign(cost0)

			// generate receipt
			message, err := json.Marshal(args)
			cost.AddAssign(host.CommonOpCost(1))
			if err != nil {
				return nil, cost, err
			}
			cost0 = h.Receipt(string(message))
			cost.AddAssign(cost0)

			return []interface{}{tokenID}, cost, nil
		},
	}

	transferToken721ABI = &abi{
		name: "transfer",
		args: []string{"string", "string", "string", "string"},
		do: func(h *host.Host, args ...interface{}) (rtn []interface{}, cost contract.Cost, err error) {
			cost = contract.Cost0()
			cost.AddAssign(host.CommonOpCost(1))
			tokenSym := args[0].(string)
			from := args[1].(string)
			to := args[2].(string)
			tokenID := args[3].(string)

			if from == to {
				return []interface{}{}, cost, nil
			}

			// get token info
			ok, cost0 := checkToken721Exists(h, tokenSym)
			cost.AddAssign(cost0)
			if !ok {
				return nil, cost, host.ErrTokenNotExists
			}

			// check auth
			ok, cost0 = h.RequireAuth(from, TransferPermission)
			cost.AddAssign(cost0)
			if !ok {
				return nil, cost, host.ErrPermissionLost
			}
			if !CheckCost(h, cost) {
				return nil, cost, host.ErrOutOfGas
			}

			tmp, cost0 := h.MapGet(Token721InfoMapPrefix+tokenSym, tokenID)
			cost.AddAssign(cost0)
			if tmp == nil {
				return nil, cost, fmt.Errorf("error tokenID not exists. %v %v", tokenSym, tokenID)
			}
			owner := tmp.(string)
			if owner != from {
				return nil, cost, fmt.Errorf("error token owner isn't from. owner: %v, from: %v", owner, from)
			}

			publisher := h.Context().Value("publisher").(string)
			cost0, err = h.MapPut(Token721InfoMapPrefix+tokenSym, tokenID, to, publisher)
			cost.AddAssign(cost0)
			if err != nil {
				return nil, cost, err
			}

			fbalance, cost0, err := getToken721Balance(h, tokenSym, from)
			cost.AddAssign(cost0)
			if err != nil {
				return nil, cost, err
			}
			tbalance, cost0, err := getToken721Balance(h, tokenSym, to)
			cost.AddAssign(cost0)
			if err != nil {
				return nil, cost, err
			}

			fbalance--
			tbalance++

			cost0 = setToken721Balance(h, tokenSym, from, fbalance, publisher)
			cost.AddAssign(cost0)
			cost0 = setToken721Balance(h, tokenSym, to, tbalance, publisher)
			cost.AddAssign(cost0)

			metaDataJSON, cost0 := h.MapGet(Token721MetadataMapPrefix+tokenSym+Token721MetadataKeySeparator+from, tokenID)
			cost.AddAssign(cost0)
			cost0, err = h.MapDel(Token721MetadataMapPrefix+tokenSym+Token721MetadataKeySeparator+from, tokenID)
			cost.AddAssign(cost0)
			if err != nil {
				return nil, cost, err
			}
			cost0, err = h.MapPut(Token721MetadataMapPrefix+tokenSym+Token721MetadataKeySeparator+to, tokenID, metaDataJSON, publisher)
			cost.AddAssign(cost0)

			// generate receipt
			message, err := json.Marshal(args)
			cost.AddAssign(host.CommonOpCost(1))
			if err != nil {
				return nil, cost, err
			}
			cost0 = h.Receipt(string(message))
			cost.AddAssign(cost0)
			return []interface{}{}, cost, nil
		},
	}

	balanceOfToken721ABI = &abi{
		name: "balanceOf",
		args: []string{"string", "string"},
		do: func(h *host.Host, args ...interface{}) (rtn []interface{}, cost contract.Cost, err error) {
			cost = contract.Cost0()
			cost.AddAssign(host.CommonOpCost(1))
			tokenSym := args[0].(string)
			to := args[1].(string)

			// check token info
			ok, cost0 := checkToken721Exists(h, tokenSym)
			cost.AddAssign(cost0)
			if !ok {
				return nil, cost, host.ErrTokenNotExists
			}

			tbalance, cost0, err := getToken721Balance(h, tokenSym, to)
			cost.AddAssign(cost0)
			if err != nil {
				return nil, cost, err
			}

			return []interface{}{tbalance}, cost, nil
		},
	}

	ownerOfToken721ABI = &abi{
		name: "ownerOf",
		args: []string{"string", "string"},
		do: func(h *host.Host, args ...interface{}) (rtn []interface{}, cost contract.Cost, err error) {
			cost = contract.Cost0()
			cost.AddAssign(host.CommonOpCost(1))
			tokenSym := args[0].(string)
			tokenID := args[1].(string)

			// check token info
			ok, cost0 := checkToken721Exists(h, tokenSym)
			cost.AddAssign(cost0)
			if !ok {
				return nil, cost, host.ErrTokenNotExists
			}

			ok, cost0 = h.MapHas(Token721InfoMapPrefix+tokenSym, tokenID)
			cost.AddAssign(cost0)
			if !ok {
				return nil, cost, host.ErrTokenExists
			}
			tmp, cost0 := h.MapGet(Token721InfoMapPrefix+tokenSym, tokenID)
			cost.AddAssign(cost0)
			owner := tmp.(string)

			return []interface{}{owner}, cost, nil
		},
	}

	tokenOfOwnerByIndexToken721ABI = &abi{
		name: "tokenOfOwnerByIndex",
		args: []string{"string", "string", "number"},
		do: func(h *host.Host, args ...interface{}) (rtn []interface{}, cost contract.Cost, err error) {
			cost = contract.Cost0()
			cost.AddAssign(host.CommonOpCost(1))
			tokenSym := args[0].(string)
			owner := args[1].(string)
			index := args[2].(int64)
			ok, cost0 := checkToken721Exists(h, tokenSym)
			cost.AddAssign(cost0)
			if !ok {
				return nil, cost, host.ErrTokenNotExists
			}
			tokens, cost0 := h.MapKeys(Token721MetadataMapPrefix + tokenSym + Token721MetadataKeySeparator + owner)
			cost.AddAssign(cost0)
			if int(index) >= len(tokens) {
				return nil, cost, errors.New("out of range")
			}

			return []interface{}{tokens[index]}, cost, nil
		},
	}

	tokenMetadataToken721ABI = &abi{
		name: "tokenMetadata",
		args: []string{"string", "string"},
		do: func(h *host.Host, args ...interface{}) (rtn []interface{}, cost contract.Cost, err error) {
			cost = contract.Cost0()
			cost.AddAssign(host.CommonOpCost(1))
			tokenSym := args[0].(string)
			tokenID := args[1].(string)
			ok, cost0 := checkToken721Exists(h, tokenSym)
			cost.AddAssign(cost0)
			if !ok {
				return nil, cost, host.ErrTokenNotExists
			}
			ok, cost0 = h.MapHas(Token721InfoMapPrefix+tokenSym, tokenID)
			cost.AddAssign(cost0)
			if !ok {
				return nil, cost, host.ErrTokenExists
			}
			tmp, cost0 := h.MapGet(Token721InfoMapPrefix+tokenSym, tokenID)
			cost.AddAssign(cost0)
			owner := tmp.(string)

			ok, cost0 = h.MapHas(Token721MetadataMapPrefix+tokenSym+Token721MetadataKeySeparator+owner, tokenID)
			cost.AddAssign(cost0)
			if !ok {
				return nil, cost, host.ErrTokenExists
			}

			metaDataJSON, cost0 := h.MapGet(Token721MetadataMapPrefix+tokenSym+Token721MetadataKeySeparator+owner, tokenID)
			cost.AddAssign(cost0)
			return []interface{}{metaDataJSON.(string)}, cost, nil
		},
	}
)
