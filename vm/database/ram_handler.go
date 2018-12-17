package database

import (
	"github.com/iost-official/go-iost/ilog"
	"strconv"
)

// RAMContractName name of basic token contract
const RAMContractName = "ram.iost"

// RAMHandler easy to get balance of token.iost
type RAMHandler struct {
	db database
}

const (
	ramFeeRate          = 0.01
	ramPriceCoefficient = 30 * 128 * 1024 * 1024
)

func (r *RAMHandler) getInt64(k string) int64 {
	finalKey := "b-" + RAMContractName + "-" + k
	val := r.db.Get(finalKey)
	resultRaw, ok := Unmarshal(val).(string)
	if !ok {
		ilog.Errorf("invalid key %v %v", finalKey, val)
		return 0
	}
	result, err := strconv.ParseInt(resultRaw, 10, 64)
	if err != nil {
		ilog.Errorf("invalid %v %v", k, resultRaw)
		return 0
	}
	return result
}

// UsedRAM ...
func (r *RAMHandler) UsedRAM() int64 {
	return r.getInt64("usedSpace")
}

// LeftRAM ...
func (r *RAMHandler) LeftRAM() int64 {
	return r.getInt64("leftSpace")
}

// TotalRAM ...
func (r *RAMHandler) TotalRAM() int64 {
	return r.UsedRAM() + r.LeftRAM()
}

func (r *RAMHandler) contractBalance() float64 {
	resultRaw := Unmarshal(r.db.Get("b-" + RAMContractName + "-" + "balance")).(string)
	result, err := strconv.ParseFloat(resultRaw, 64)
	if err != nil {
		ilog.Errorf("invalid balance %v", resultRaw)
		return 0
	}
	return result
}

// BuyPrice ...
func (r *RAMHandler) BuyPrice() float64 {
	return (1 + ramFeeRate) * ramPriceCoefficient / float64(r.LeftRAM())
}

// SellPrice ...
func (r *RAMHandler) SellPrice() float64 {
	return r.contractBalance() / float64(r.UsedRAM())
}
