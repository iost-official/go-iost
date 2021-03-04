package host

import (
	"encoding/json"

	"github.com/iost-official/go-iost/v3/core/contract"
	"github.com/iost-official/go-iost/v3/vm/database"
)

// Info current info handler of this isolate
type Info struct {
	h *Host
}

// NewInfo new info
func NewInfo(h *Host) Info {
	return Info{h: h}
}

// BlockInfo get block info, in json
func (h *Info) BlockInfo() (info database.SerializedJSON, cost contract.Cost) {
	blkInfo := make(map[string]interface{})

	blkInfo["parent_hash"] = h.h.ctx.Value("parent_hash")
	blkInfo["number"] = h.h.ctx.Value("number")
	blkInfo["witness"] = h.h.ctx.Value("witness")
	blkInfo["time"] = h.h.ctx.Value("time")

	bij, err := json.Marshal(blkInfo)
	if err != nil {
		panic(err)
	}

	return database.SerializedJSON(bij), Costs["ContextCost"]
}

// BlockTime get block time, in int64
func (h *Info) BlockTime() (ntime int64, cost contract.Cost) {
	ntime = h.h.ctx.Value("time").(int64)
	return ntime, Costs["ContextCost"]
}

// ContractName get block time, in int64
func (h *Info) ContractName() (name string, cost contract.Cost) {
	name = h.h.ctx.Value("contract_name").(string)
	return name, Costs["ContextCost"]
}

// ContextInfo get context info
func (h *Info) ContextInfo() (info database.SerializedJSON, cost contract.Cost) {
	ctxInfo := make(map[string]interface{})

	ctxInfo["contract_name"] = h.h.ctx.Value("contract_name")
	ctxInfo["abi_name"] = h.h.ctx.Value("abi_name")
	ctxInfo["publisher"] = h.h.ctx.Value("publisher")

	if h.h.IsFork3_3_0 {
		ctxInfo["caller"] = h.h.ctx.Value("caller")
	}

	cij, err := json.Marshal(ctxInfo)
	if err != nil {
		panic(err)
	}

	return database.SerializedJSON(cij), Costs["ContextCost"]
}

// TxInfo get tx info
func (h *Info) TxInfo() (info database.SerializedJSON, cost contract.Cost) {
	txInfo := make(map[string]interface{})
	txInfo["time"] = h.h.ctx.Value("tx_time")
	txInfo["hash"] = h.h.ctx.Value("tx_hash")
	txInfo["expiration"] = h.h.ctx.Value("expiration")
	txInfo["gas_limit"] = h.h.ctx.GValue("gas_limit")
	txInfo["gas_ratio"] = h.h.ctx.Value("gas_ratio")
	txInfo["auth_list"] = h.h.ctx.Value("auth_list")
	txInfo["publisher"] = h.h.ctx.Value("publisher")
	if h.h.IsFork3_3_0 {
		txInfo["amount_limit"] = h.h.ctx.Value("amount_limit")
		txInfo["actions"] = h.h.ctx.Value("actions")
	}

	tij, err := json.Marshal(txInfo)
	if err != nil {
		panic(err)
	}

	return database.SerializedJSON(tij), Costs["ContextCost"]
}

// ABIConfig set this abi config
func (h *Info) ABIConfig(key, value string) {
	switch key {
	case "payment":
		if value == "contract_pay" {
			h.h.ctx.GSet("abi_payment", 1)
		}
	}
}

// GasLimitValue get gas limit
func (h *Info) GasLimitValue() int64 {
	return h.h.ctx.GValue("gas_limit").(int64)
}
