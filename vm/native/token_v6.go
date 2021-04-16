package native

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/iost-official/go-iost/v3/core/contract"
	"github.com/iost-official/go-iost/v3/vm/host"
)

var tokenABIsV6 *abiSet

func init() {
	tokenABIsV6 = newAbiSet()
	tokenABIsV6.Register(initTokenABI, true)
	tokenABIsV6.Register(balanceOfTokenABI)
	tokenABIsV6.Register(createTokenABI)
	tokenABIsV6.Register(supplyTokenABI)
	tokenABIsV6.Register(totalSupplyTokenABI)
	tokenABIsV6.Register(issueTokenABIV2)
	tokenABIsV6.Register(transferFreezeTokenABIV2)
	tokenABIsV6.Register(destroyTokenABIV2)
	tokenABIsV6.Register(transferTokenABIV3)
	tokenABIsV6.Register(recycleTokenABI)

	initReserveList()
}

type ReserveToken struct {
	sym     string
	sudoAcc string
	until   time.Time
}

var ReserveList []ReserveToken

func initReserveList() {
	ReserveList = make([]ReserveToken, 0)
	husdTime, err := time.Parse(time.RFC3339, "2021-01-10T00:00:00Z")
	if err != nil {
		panic(err)
	}
	ReserveList = append(ReserveList, ReserveToken{"husd", "admin", husdTime})
	lolTime, err := time.Parse(time.RFC3339, "2021-05-30T00:00:00Z")
	if err != nil {
		panic(err)
	}
	ReserveList = append(ReserveList, ReserveToken{"lol", "emogi", lolTime})
}

func checkReserveToken(h *host.Host, tokenSym string) bool {
	valid := false
	blockTime := h.Context().Value("time").(int64)
	publisher := h.Context().Value("publisher").(string)
	for _, item := range ReserveList {
		fmt.Println("checkReserveToken", item.sym, item.until.UnixNano(), blockTime)
		if item.sym == tokenSym && blockTime < item.until.UnixNano() && (publisher == "admin" || publisher == item.sudoAcc) {
			valid = true
			break
		}
	}
	if valid {
		fmt.Println("checkReserveToken bypass", tokenSym, publisher)
	}
	return valid
}

var (
	recycleTokenABI = &abi{
		name: "recycle",
		args: []string{"string"},
		do: func(h *host.Host, args ...interface{}) (rtn []interface{}, cost contract.Cost, err error) {
			cost = contract.Cost0()
			cost.AddAssign(host.CommonOpCost(1))
			tokenSym := args[0].(string)
			publisher := h.Context().Value("publisher").(string)
			if publisher != "admin" {
				return nil, cost, fmt.Errorf("permission denied")
			}
			valid := checkReserveToken(h, tokenSym)
			if !valid {
				return nil, cost, fmt.Errorf("invalid token to recycle")
			}
			cost0, err := h.MapDel(TokenInfoMapPrefix+tokenSym, IssuerMapField)
			cost.AddAssign(cost0)
			if err != nil {
				return nil, cost, err
			}
			exist, cost0 := checkTokenExists(h, tokenSym)
			cost.AddAssign(cost0)
			if exist {
				return nil, cost, fmt.Errorf("recycle token failed")
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
