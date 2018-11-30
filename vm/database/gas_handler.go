package database

import (
	"github.com/iost-official/go-iost/common"
	"github.com/iost-official/go-iost/ilog"
)

const (
	// GasRateKey : how much gas is generated per IOST per second
	GasRateKey = "gr"
	// GasLimitKey : how much gas can be generated max per IOST
	GasLimitKey = "gl"
	// GasUpdateTimeKey : when the gas state is refreshed last time, for internal use
	GasUpdateTimeKey = "gt"
	// GasStockKey : how much gas is there when last time refreshed
	GasStockKey = "gs"
	// GasPledgeKey : who pledge how much coins for me
	GasPledgeKey = "gp"
)

// decimals of gas
const (
	GasDecimal = 2
)
const gasMaxIncreaseSeconds = 3 * 24 * 3600

// PledgerInfo ...
type PledgerInfo struct {
	Pledger string
	Amount  *common.Fixed
}

// GasContractName name of basic token contract
const GasContractName = "gas.iost"

// GasHandler easy to get balance of token.iost
type GasHandler struct {
	BasicHandler
	MapHandler
}

// If no key exists, return 0
func (g *GasHandler) getFixed(owner string, key string) *common.Fixed {
	result := MustUnmarshal(g.BasicHandler.Get(GasContractName + Separator + key + owner))
	if result == nil {
		//ilog.Errorf("GasHandler failed %v %v", owner, key)
		return nil
	}
	value, ok := result.(*common.Fixed)
	if !ok {
		ilog.Errorf("GasHandler failed %v %v %v", owner, key, result)
		return nil
	}
	return value
}

// GasRate ...
func (g *GasHandler) GasRate(name string) *common.Fixed {
	f := g.getFixed(name, GasRateKey)
	if f == nil {
		return &common.Fixed{
			Value:   0,
			Decimal: GasDecimal,
		}
	}
	return f
}

// GasLimit ...
func (g *GasHandler) GasLimit(name string) *common.Fixed {
	f := g.getFixed(name, GasLimitKey)
	if f == nil {
		return &common.Fixed{
			Value:   0,
			Decimal: GasDecimal,
		}
	}
	return f
}

// GasUpdateTime ...
func (g *GasHandler) GasUpdateTime(name string) int64 {
	value := MustUnmarshal(g.BasicHandler.Get(GasContractName + Separator + GasUpdateTimeKey + name))
	if value == nil {
		return 0
	}
	return value.(int64)
}

// GasStock `gasStock` means the gas amount at last update time.
func (g *GasHandler) GasStock(name string) *common.Fixed {
	f := g.getFixed(name, GasStockKey)
	if f == nil {
		return &common.Fixed{
			Value:   0,
			Decimal: GasDecimal,
		}
	}
	return f
}

// GasPledge ...
func (g *GasHandler) GasPledge(name string, pledger string) *common.Fixed {
	ok := g.MapHandler.MHas(GasContractName+Separator+GasPledgeKey+name, pledger)
	if !ok {
		return &common.Fixed{
			Value:   0,
			Decimal: 8,
		}
	}
	result := MustUnmarshal(g.MapHandler.MGet(GasContractName+Separator+GasPledgeKey+name, pledger))
	value, ok := result.(*common.Fixed)
	if !ok {
		return nil
	}
	return value
}

// PledgerInfo get who pledged how much coins for me
func (g *GasHandler) PledgerInfo(name string) []PledgerInfo {
	pledgers := g.MapHandler.MKeys(GasContractName + Separator + GasPledgeKey + name)
	ilog.Errorf("pledge keys %v %v", pledgers, name)
	result := make([]PledgerInfo, 0)
	for _, pledger := range pledgers {
		s := g.MapHandler.MGet(GasContractName+Separator+GasPledgeKey+name, pledger)
		v := MustUnmarshal(s)
		pledge, ok := v.(*common.Fixed)
		if !ok {
			return make([]PledgerInfo, 0)
		}
		result = append(result, PledgerInfo{pledger, pledge})
	}
	return result
}

// CurrentTotalGas return current total gas. It is min(limit, last_updated_gas + time_since_last_updated * increase_speed)
func (g *GasHandler) CurrentTotalGas(name string, now int64) (result *common.Fixed) {
	if now <= 0 {
		ilog.Fatalf("CurrentTotalGas failed. invalid now time %v", now)
	}
	result = g.GasStock(name)
	gasUpdateTime := g.GasUpdateTime(name)
	var durationSeconds int64
	if gasUpdateTime > 0 {
		durationSeconds = (now - gasUpdateTime) / 1e9
		if durationSeconds > gasMaxIncreaseSeconds {
			durationSeconds = gasMaxIncreaseSeconds
		}
	}
	if durationSeconds < 0 {
		ilog.Fatalf("CurrentTotalGas durationSeconds invalid %v = %v - %v", durationSeconds, now, gasUpdateTime)
	}
	rate := g.GasRate(name)
	limit := g.GasLimit(name)
	//fmt.Printf("CurrentTotalGas user %v stock %v rate %v limit %v\n", name, result, rate, limit)
	delta := rate.Times(durationSeconds)
	if delta == nil {
		ilog.Errorf("CurrentTotalGas may overflow rate %v durationSeconds %v", rate, durationSeconds)
		return
	}
	result = result.Add(delta)
	if limit.LessThan(result) {
		result = limit
	}
	return
}
