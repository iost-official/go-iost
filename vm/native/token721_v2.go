package native

import (
	"encoding/json"

	"fmt"

	"github.com/iost-official/go-iost/core/contract"
	"github.com/iost-official/go-iost/vm/host"
)

var token721ABIsV2 *abiSet

func init() {
	token721ABIsV2 = newAbiSet()
	token721ABIsV2.Register(initToken721ABI, true)
	token721ABIsV2.Register(createToken721ABI)
	token721ABIsV2.Register(issueToken721ABI)
	token721ABIsV2.Register(balanceOfToken721ABI)
	token721ABIsV2.Register(ownerOfToken721ABI)
	token721ABIsV2.Register(tokenOfOwnerByIndexToken721ABI)
	token721ABIsV2.Register(tokenMetadataToken721ABI)

	// modified methods for V2
	token721ABIsV2.Register(transferToken721ABIV2)
	token721ABIsV2.Register(transferWithMemoToken721ABIV2)
	token721ABIsV2.Register(approveToken721ABIV2)
}

var (
	transferToken721ABIV2 = &abi{
		name: "transfer",
		args: []string{"string", "string", "string", "string"},
		do: func(h *host.Host, args ...interface{}) (rtn []interface{}, cost contract.Cost, err error) {
			cost = contract.Cost0()
			cost.AddAssign(host.CommonOpCost(1))
			tokenSym := args[0].(string)
			from := args[1].(string)
			to := args[2].(string)
			tokenID := args[3].(string)

			// get token info
			ok, cost0 := checkToken721Exists(h, tokenSym)
			cost.AddAssign(cost0)
			if !ok {
				return nil, cost, host.ErrTokenNotExists
			}

			// check auth
			caller := h.Caller()
			if caller.Name == from {
				ok, cost0 = h.RequireAuth(from, TransferPermission)
				cost.AddAssign(cost0)
				if !ok {
					return nil, cost, host.ErrPermissionLost
				}
			} else {
				val, cost0 := h.MapGet(Token721ApproveTokenPrefix+tokenSym, tokenID)
				cost.AddAssign(cost0)
				if val == nil {
					return nil, cost, host.ErrPermissionLost
				}
				ok, cost0 = h.RequireAuth(val.(string), TransferPermission)
				cost.AddAssign(cost0)
				if !ok {
					return nil, cost, host.ErrPermissionLost
				}
			}
			cost0, err = h.MapDel(Token721ApproveTokenPrefix+tokenSym, tokenID)
			cost.AddAssign(cost0)
			if err != nil {
				return nil, cost, fmt.Errorf("error can not remove approval. %v %v", tokenSym, caller.Name)
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

			// change 'from' balance
			fbalance, cost0, err := getToken721Balance(h, tokenSym, from)
			cost.AddAssign(cost0)
			if err != nil {
				return nil, cost, err
			}
			fbalance--
			cost0 = setToken721Balance(h, tokenSym, from, fbalance, publisher)
			cost.AddAssign(cost0)

			// change 'to' balance
			tbalance, cost0, err := getToken721Balance(h, tokenSym, to)
			cost.AddAssign(cost0)
			if err != nil {
				return nil, cost, err
			}
			tbalance++
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
			if err != nil {
				return nil, cost, err
			}

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

	transferWithMemoToken721ABIV2 = &abi{
		name: "transferWithMemo",
		args: []string{"string", "string", "string", "string", "string"},
		do: func(h *host.Host, args ...interface{}) (rtn []interface{}, cost contract.Cost, err error) {
			memo := args[4].(string)
			if len(memo) > 512 {
				return nil, cost, host.ErrMemoTooLarge
			}
			return transferToken721ABIV2.do(h, args...)
		},
	}

	approveToken721ABIV2 = &abi{
		name: "approve",
		args: []string{"string", "string", "string", "string"},
		do: func(h *host.Host, args ...interface{}) (rtn []interface{}, cost contract.Cost, err error) {
			cost = contract.Cost0()
			cost.AddAssign(host.CommonOpCost(1))
			tokenSym := args[0].(string)
			from := args[1].(string)
			to := args[2].(string)
			tokenID := args[3].(string)

			// get token info
			ok, cost0 := checkToken721Exists(h, tokenSym)
			cost.AddAssign(cost0)
			if !ok {
				return nil, cost, host.ErrTokenNotExists
			}

			// check direct call
			caller := h.Caller()
			if caller.Name != from {
				return nil, cost, host.ErrPermissionLost
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
			cost0, err = h.MapPut(Token721ApproveTokenPrefix+tokenSym, tokenID, to, publisher)
			cost.AddAssign(cost0)
			if err != nil {
				return nil, cost, err
			}

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
)
