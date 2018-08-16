package host

import (
	"github.com/iost-official/Go-IOS-Protocol/core/contract"
	"github.com/iost-official/Go-IOS-Protocol/new_vm/database"
)

type Info struct {
	ctx *Context
}

func NewInfo(ctx *Context) Info {
	return Info{ctx: ctx}
}

func (h *Info) BlockInfo() (info database.SerializedJSON, cost *contract.Cost) { // todo 清理并且保证正确
	return h.ctx.Value("block_info").(database.SerializedJSON), contract.NewCost(1, 1, 1)
}
func (h *Info) TxInfo() (info database.SerializedJSON, cost *contract.Cost) {
	return h.ctx.Value("tx_info").(database.SerializedJSON), contract.NewCost(1, 1, 1)
}
func (h *Info) ABIConfig(key, value string) {
	switch key {
	case "payment":
		if value == "contract_pay" {
			h.ctx.GSet("abi_payment", 1)
		}
	}
}

func (h *Info) GasLimit() uint64 {
	return h.ctx.GValue("gas_limit").(uint64)
}
