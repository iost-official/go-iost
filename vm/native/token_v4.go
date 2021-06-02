package native

import (
	"errors"
	"fmt"
	"math"

	"github.com/iost-official/go-iost/v3/core/contract"
	"github.com/iost-official/go-iost/v3/vm/host"
)

var tokenABIsV4 *abiSet

func init() {
	tokenABIsV4 = newAbiSet()
	tokenABIsV4.Register(initTokenABI, true)
	tokenABIsV4.Register(createTokenABI)
	tokenABIsV4.Register(balanceOfTokenABI)
	tokenABIsV4.Register(supplyTokenABI)
	tokenABIsV4.Register(totalSupplyTokenABI)

	// modified methods for V2
	tokenABIsV4.Register(issueTokenABIV2)
	tokenABIsV4.Register(transferFreezeTokenABIV2)
	tokenABIsV4.Register(destroyTokenABIV2)

	// modified methods for V3
	tokenABIsV4.Register(transferTokenABIV3)

	// modified methods for V4
	tokenABIsV4.Register(updateTokenTotalSupplyABI)
}

var (
	updateTokenTotalSupplyABI = &abi{
		name: "updateTotalSupply",
		args: []string{"string", "number"},
		do: func(h *host.Host, args ...interface{}) (rtn []interface{}, cost contract.Cost, err error) {
			cost = contract.Cost0()
			cost.AddAssign(host.CommonOpCost(1))
			tokenSym := args[0].(string)
			newTotalSupply := args[1].(int64)

			if tokenSym == "iost" {
				return nil, cost, errors.New("native token cannot be changed")
			}
			// get token info
			ok, cost0 := checkTokenExists(h, tokenSym)
			cost.AddAssign(cost0)
			if !ok {
				return nil, cost, host.ErrTokenNotExists
			}
			issuer, cost0 := h.MapGet(TokenInfoMapPrefix+tokenSym, IssuerMapField)
			cost.AddAssign(cost0)
			supply, cost0 := h.MapGet(TokenInfoMapPrefix+tokenSym, SupplyMapField)
			cost.AddAssign(cost0)
			if !CheckCost(h, cost) {
				return nil, cost, host.ErrOutOfGas
			}

			// check auth
			ok, cost0 = h.RequireAuth(issuer.(string), TokenPermission)
			cost.AddAssign(cost0)
			if !ok {
				return nil, cost, host.ErrPermissionLost
			}
			if !CheckCost(h, cost) {
				return nil, cost, host.ErrOutOfGas
			}

			// get amount by fixed point number
			if newTotalSupply <= 0 {
				return nil, cost, host.ErrInvalidAmount
			}
			decimal, cost := h.MapGet(TokenInfoMapPrefix+tokenSym, DecimalMapField)
			cost.AddAssign(cost)
			decimalInt := int(decimal.(int64))
			if newTotalSupply > math.MaxInt64/int64(math.Pow10(decimalInt)) {
				return nil, cost, fmt.Errorf("invalid totalSupply, must be less than %v", math.MaxInt64/int64(math.Pow10(decimalInt)))
			}
			newTotalSupply *= int64(math.Pow10(decimalInt))
			if newTotalSupply < supply.(int64) {
				return nil, cost, errors.New("invalid totalSupply, must be more than current supply")
			}
			if !CheckCost(h, cost) {
				return nil, cost, host.ErrOutOfGas
			}
			publisher := h.Context().Value("publisher").(string)
			cost0, _ = h.MapPut(TokenInfoMapPrefix+tokenSym, TotalSupplyMapField, newTotalSupply, publisher)
			cost.AddAssign(cost0)
			return []interface{}{}, cost, nil
		},
	}
)
