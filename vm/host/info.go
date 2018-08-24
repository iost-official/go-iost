package host

import (
	"encoding/json"

	"github.com/iost-official/Go-IOS-Protocol/core/contract"
	"github.com/iost-official/Go-IOS-Protocol/vm/database"
)

// Info ...
type Info struct {
	h *Host
}

// NewInfo ...
func NewInfo(h *Host) Info {
	return Info{h: h}
}

// BlockInfo ...
func (h *Info) BlockInfo() (info database.SerializedJSON, cost *contract.Cost) { // todo 清理并且保证正确

	blkInfo := make(map[string]interface{})

	blkInfo["parent_hash"] = h.h.ctx.Value("parent_hash")
	blkInfo["number"] = h.h.ctx.Value("number")
	blkInfo["witness"] = h.h.ctx.Value("witness")
	blkInfo["time"] = h.h.ctx.Value("time")

	bij, err := json.Marshal(blkInfo)
	if err != nil {
		panic(err)
	}

	return database.SerializedJSON(bij), BlockInfoCost
}

// TxInfo ...
func (h *Info) TxInfo() (info database.SerializedJSON, cost *contract.Cost) {

	txInfo := make(map[string]interface{})
	txInfo["time"] = h.h.ctx.Value("time")
	txInfo["hash"] = h.h.ctx.Value("tx_hash")
	txInfo["expiration"] = h.h.ctx.Value("expiration")
	txInfo["gas_limit"] = h.h.ctx.GValue("gas_limit")
	txInfo["gas_price"] = h.h.ctx.Value("gas_price")
	txInfo["auth_list"] = h.h.ctx.Value("auth_list")

	tij, err := json.Marshal(txInfo)
	if err != nil {
		panic(err)
	}

	return database.SerializedJSON(tij), TxInfoCost
}

// ABIConfig ...
func (h *Info) ABIConfig(key, value string) {
	switch key {
	case "payment":
		if value == "contract_pay" {
			h.h.ctx.GSet("abi_payment", 1)
		}
	}
}

// GasLimit ...
func (h *Info) GasLimit() int64 {
	return h.h.ctx.GValue("gas_limit").(int64)
}
