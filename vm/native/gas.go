package native

import (
	"fmt"

	"github.com/iost-official/go-iost/common"
	"github.com/iost-official/go-iost/core/contract"
	"github.com/iost-official/go-iost/vm/host"
)

// IOSTRatio ...
const IOSTRatio int64 = 100000000

// GasMinPledgeInIOST Every user must pledge a minimum amount of IOST
var GasMinPledgeInIOST int64 = 10

// GasMinPledge Every user must pledge a minimum amount of IOST (including GAS and RAM)
var GasMinPledge = &common.Fixed{Value: GasMinPledgeInIOST * IOSTRatio, Decimal: 8}

// Each IOST you pledge, you will get `GasImmediateReward` gas immediately.
// Then gas will be generated at a rate of `GasIncreaseRate` gas per block.
// Then it takes `GasFulfillSeconds` time to reach the limit.
// Your gas production will stop when it reaches the limit.
// When you use some gas later, the total amount will be less than the limit,
// so gas production will resume again util the limit.

// GasImmediateReward immediate reward per IOST
var GasImmediateReward = &common.Fixed{Value: 10000 * 100, Decimal: 2}

// GasLimit gas limit per IOST
var GasLimit = &common.Fixed{Value: 30000 * 100, Decimal: 2}

// GasFulfillSeconds it takes 2 days to fulfill the gas buffer.
var GasFulfillSeconds int64 = 2 * 24 * 3600

// GasIncreaseRate gas increase per IOST per second
var GasIncreaseRate = GasLimit.Sub(GasImmediateReward).Div(GasFulfillSeconds)

//var GasIncreaseRate = &common.Fixed{Value: 1 * IOSTRatio, Decimal: 8}

// UnpledgeFreezeSeconds coins will be frozen for 3 days after being unpledged
var UnpledgeFreezeSeconds int64 = 3 * 24 * 3600

var gasABIs map[string]*abi
var gasInnerABIs map[string]*abi

func init() {
	gasABIs = make(map[string]*abi)
	register(gasABIs, constructor)
	register(gasABIs, pledgeGas)
	register(gasABIs, unpledgeGas)

	gasInnerABIs = make(map[string]*abi)
	register(gasInnerABIs, initFunc)
}

// Pledge Change all gas related storage here. If pledgeAmount > 0. pledge. If pledgeAmount < 0, unpledge.
func pledge(h *host.Host, pledger string, name string, pledgeAmountF *common.Fixed) (contract.Cost, error) {
	finalCost := contract.Cost0()
	if pledgeAmountF.IsZero() {
		return finalCost, fmt.Errorf("invalid pledge amount %v", pledgeAmountF.ToString())
	}
	pledged, cost := h.GasManager.GasPledge(name, pledger)
	finalCost.AddAssign(cost)
	if pledgeAmountF.IsNegative() {
		unpledgeAmount := pledgeAmountF.Neg()
		newPledge := pledged.Sub(unpledgeAmount)
		if pledger == name {
			if newPledge.LessThan(GasMinPledge) {
				return finalCost, fmt.Errorf("unpledge to much %v - %v < %v", pledged.ToString(), unpledgeAmount.ToString(), GasMinPledge.ToString())
			}
		}
		if newPledge.Value <= 0 {
			pledgeAmountF = pledged.Neg()
			cost = h.GasManager.DelGasPledge(name, pledger)
			finalCost.AddAssign(cost)
			//return fmt.Errorf("unpledge to much %v > %v", unpledgeAmount.ToString(), pledged.ToString()), finalCost
		}
	}

	limitDelta := pledgeAmountF.Multiply(GasLimit)
	rateDelta := pledgeAmountF.Multiply(GasIncreaseRate)
	gasDelta := pledgeAmountF.Multiply(GasImmediateReward)
	if pledgeAmountF.IsNegative() {
		// unpledge should not change current generated gas
		gasDelta.Value = 0
	}
	//fmt.Printf("limitd rated gasd %v %v %v\n", limitDelta, rateDelta, gasDelta)

	// pledge first time
	t, cost := h.GasManager.GasUpdateTime(name)
	finalCost.AddAssign(cost)
	if t == 0 {
		if pledgeAmountF.IsNegative() {
			return finalCost, fmt.Errorf("cannot unpledge! No pledge before")
		}
		cost = h.GasManager.SetGasPledge(name, pledger, pledgeAmountF)
		finalCost.AddAssign(cost)
		cost = h.GasManager.SetGasUpdateTime(name, h.Context().Value("time").(int64))
		finalCost.AddAssign(cost)
		cost = h.GasManager.SetGasRate(name, rateDelta)
		finalCost.AddAssign(cost)
		cost = h.GasManager.SetGasLimit(name, limitDelta)
		finalCost.AddAssign(cost)
		cost = h.GasManager.SetGasStock(name, gasDelta)
		finalCost.AddAssign(cost)
		return finalCost, nil
	}
	cost, _ = h.GasManager.RefreshGas(name)
	finalCost.AddAssign(cost)
	rateOld, cost := h.GasManager.GasRate(name)
	finalCost.AddAssign(cost)
	rateNew := rateOld.Add(rateDelta)
	if rateNew.Value <= 0 {
		return finalCost, fmt.Errorf("change gasRate failed! current: %v, delta %v", rateOld.ToString(), rateDelta.ToString())
	}
	limitOld, cost := h.GasManager.GasLimit(name)
	finalCost.AddAssign(cost)
	limitNew := limitOld.Add(limitDelta)
	if limitNew.Value <= 0 {
		return finalCost, fmt.Errorf("change gasLimit failed! current: %v, delta %v", limitOld.ToString(), limitDelta.ToString())
	}
	gasOld, cost := h.GasManager.GasStock(name)
	finalCost.AddAssign(cost)
	gasNew := gasOld.Add(gasDelta)
	if limitNew.LessThan(gasNew) {
		// clear the gas above the new limit.
		gasNew = limitNew
	}

	//fmt.Printf("Pledge %v", pledgeAmountF)
	newPledge := pledged.Add(pledgeAmountF)
	if !newPledge.IsZero() {
		cost = h.GasManager.SetGasPledge(name, pledger, newPledge)
		finalCost.AddAssign(cost)
	}
	cost = h.GasManager.SetGasRate(name, rateNew)
	finalCost.AddAssign(cost)
	cost = h.GasManager.SetGasLimit(name, limitNew)
	finalCost.AddAssign(cost)
	cost = h.GasManager.SetGasStock(name, gasNew)
	finalCost.AddAssign(cost)
	return finalCost, nil
}

var (
	constructor = &abi{
		name: "constructor",
		args: []string{},
		do: func(h *host.Host, args ...interface{}) (rtn []interface{}, cost contract.Cost, err error) {
			return []interface{}{}, host.CommonErrorCost(1), nil
		},
	}
	initFunc = &abi{
		name: "init",
		args: []string{},
		do: func(h *host.Host, args ...interface{}) (rtn []interface{}, cost contract.Cost, err error) {
			return []interface{}{}, host.CommonErrorCost(1), nil
		},
	}
	pledgeGas = &abi{
		name: "pledge",
		args: []string{"string", "string", "string"},
		do: func(h *host.Host, args ...interface{}) (rtn []interface{}, cost contract.Cost, err error) {
			cost = contract.Cost0()
			pledger, ok := args[0].(string)
			cost.AddAssign(host.CommonErrorCost(1))
			if !ok || !h.IsValidAccount(pledger) {
				return nil, cost, fmt.Errorf("invalid user name %s", args[0])
			}
			gasUser, ok := args[1].(string)
			cost.AddAssign(host.CommonErrorCost(1))
			if !ok || !h.IsValidAccount(gasUser) {
				return nil, cost, fmt.Errorf("invalid user name %s", args[1])
			}
			auth, cost0 := h.RequireAuth(pledger, "transfer")
			cost.AddAssign(cost0)
			if !auth {
				return nil, cost, host.ErrPermissionLost
			}
			pledgeAmountStr, ok := args[2].(string)
			if !ok {
				return nil, cost, fmt.Errorf("invalid amount %s", args[2])
			}
			pledgeAmount, err := common.NewFixed(pledgeAmountStr, 8)
			cost.AddAssign(host.CommonErrorCost(1))
			if err != nil || pledgeAmount.Value <= 0 {
				return nil, cost, fmt.Errorf("invalid amount %s", args[2])
			}
			var minPledgeAmount int64 = 1 * IOSTRatio
			if pledgeAmount.Value < minPledgeAmount {
				return nil, cost, fmt.Errorf("min pledge num is %d", minPledgeAmount)
			}
			contractName, cost0 := h.ContractName()
			cost.AddAssign(cost0)
			_, cost0, err = h.Call("token.iost", "transfer", fmt.Sprintf(`["iost", "%v", "%v", "%v", ""]`, pledger, contractName, pledgeAmountStr))
			cost.AddAssign(cost0)
			if err != nil {
				return nil, cost, err
			}
			cost0, err = pledge(h, pledger, gasUser, pledgeAmount)
			cost.AddAssign(cost0)
			if err != nil {
				return nil, cost, err
			}
			return []interface{}{}, cost, nil
		},
	}
	unpledgeGas = &abi{
		name: "unpledge",
		args: []string{"string", "string", "string"},
		do: func(h *host.Host, args ...interface{}) (rtn []interface{}, cost contract.Cost, err error) {
			cost = contract.Cost0()
			pledger, ok := args[0].(string)
			cost.AddAssign(host.CommonErrorCost(1))
			if !ok || !h.IsValidAccount(pledger) {
				return nil, cost, fmt.Errorf("invalid user name %s", args[1])
			}
			gasUser, ok := args[1].(string)
			cost.AddAssign(host.CommonErrorCost(1))
			if !ok || !h.IsValidAccount(gasUser) {
				return nil, cost, fmt.Errorf("invalid user name %s", args[0])
			}
			auth, cost0 := h.RequireAuth(pledger, "transfer")
			cost.AddAssign(cost0)
			if !auth {
				return nil, cost, host.ErrPermissionLost
			}
			unpledgeAmountStr, ok := args[2].(string)
			if !ok {
				return nil, cost, fmt.Errorf("invalid amount %s", args[2])
			}
			unpledgeAmount, err := common.NewFixed(unpledgeAmountStr, 8)
			cost.AddAssign(host.CommonErrorCost(1))
			if err != nil || unpledgeAmount.Value <= 0 {
				return nil, cost, fmt.Errorf("invalid amount %s", args[2])
			}
			var minUnpledgeAmount int64 = 1 * IOSTRatio
			if unpledgeAmount.Value < minUnpledgeAmount {
				return nil, cost, fmt.Errorf("min unpledge num is %d", minUnpledgeAmount)
			}
			pledged, cost := h.GasManager.GasPledge(gasUser, pledger)
			if pledged.IsZero() {
				return nil, cost, fmt.Errorf("%v did not pledge for %v", pledger, gasUser)
			}
			if pledged.LessThan(unpledgeAmount) {
				unpledgeAmount = pledged.Neg()
			}
			cost0, err = pledge(h, pledger, gasUser, unpledgeAmount.Neg())
			cost.AddAssign(cost0)
			if err != nil {
				return nil, cost, err
			}
			contractName, cost0 := h.ContractName()
			cost.AddAssign(cost0)
			freezeTime := h.Context().Value("time").(int64) + UnpledgeFreezeSeconds*1e9
			_, cost0, err = h.CallWithAuth("token.iost", "transferFreeze",
				fmt.Sprintf(`["iost", "%v", "%v", "%v", %v, ""]`, contractName, pledger, unpledgeAmount.ToString(), freezeTime))
			cost.AddAssign(cost0)
			if err != nil {
				return nil, cost, err
			}
			return []interface{}{}, cost, nil
		},
	}
)
