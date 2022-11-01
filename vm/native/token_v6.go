package native

import (
	"fmt"

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
}

var (
	recycleTokenABI = &abi{
		name: "recycle",
		args: []string{"string"},
		do: func(h *host.Host, args ...any) (rtn []any, cost contract.Cost, err error) {
			cost = contract.Cost0()
			cost.AddAssign(host.CommonOpCost(1))
			return nil, cost, fmt.Errorf("permission denied")
		},
	}
)
