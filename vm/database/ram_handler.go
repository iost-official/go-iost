package database

import (
	"strconv"

	"github.com/iost-official/go-iost/ilog"
)

// RAMContractName name of basic token contract
const RAMContractName = "ram.iost"

// RAMHandler easy to get balance of token.iost
type RAMHandler struct {
	BasicHandler
}

const (
	ramFeeRate = 0.02
	ramF       = 1.0
)

// AccountRAMInfo ...
type AccountRAMInfo struct {
	// Used ...
	Used int64
	// Available ...
	Available int64
	// Total ...
	Total int64
}

func (r *RAMHandler) getInt64(k string) int64 {
	finalKey := RAMContractName + "-" + k
	val := r.BasicHandler.Get(finalKey)
	resultRaw, ok := Unmarshal(val).(string)
	if !ok {
		if val != NilPrefix {
			ilog.Errorf("invalid key %v %v", finalKey, val)
		}
		return 0
	}
	result, err := strconv.ParseInt(resultRaw, 10, 64)
	if err != nil {
		ilog.Warnf("invalid value string, key:%v value:%v", k, resultRaw)
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
	resultRaw := Unmarshal(r.BasicHandler.Get(RAMContractName + "-" + "balance")).(string)
	result, err := strconv.ParseFloat(resultRaw, 64)
	if err != nil {
		ilog.Errorf("invalid balance %v", resultRaw)
		return 0
	}
	return result
}

// BuyPrice ...
func (r *RAMHandler) BuyPrice() float64 {
	return (1 + ramFeeRate) * ramF * r.contractBalance() / float64(r.LeftRAM())
}

// SellPrice ...
func (r *RAMHandler) SellPrice() float64 {
	return ramF * r.contractBalance() / float64(r.LeftRAM())
}

// GetAccountRAMInfo ...
func (r *RAMHandler) GetAccountRAMInfo(acc string) *AccountRAMInfo {
	return &AccountRAMInfo{
		Used:      r.getInt64("UR" + acc),
		Total:     r.getInt64("TR" + acc),
		Available: (&TokenHandler{r.BasicHandler.db}).TokenBalance("ram", acc),
	}
}

// ChangeUsedRAMInfo ...
func (r *RAMHandler) ChangeUsedRAMInfo(acc string, delta int64) {
	value, _ := Marshal(strconv.FormatInt(r.getInt64("UR"+acc)+delta, 10))
	r.BasicHandler.Put(RAMContractName+"-"+"UR"+acc, value)
}
