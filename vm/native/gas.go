package native

import (
	"fmt"
	"github.com/iost-official/go-iost/vm/database"

	"github.com/iost-official/go-iost/common"
	"github.com/iost-official/go-iost/core/contract"
	"github.com/iost-official/go-iost/vm/host"
)

// IOSTRatio ...
const IOSTRatio int64 = 100000000

// GasMinPledgeOfUser Each user must pledge a minimum amount of IOST
var GasMinPledgeOfUser = &common.Fixed{Value: 10 * IOSTRatio, Decimal: 8}

// GasMinPledgePerAction One must (un)pledge more than 1 IOST
var GasMinPledgePerAction = &common.Fixed{Value: 1 * IOSTRatio, Decimal: 8}

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
const GasFulfillSeconds int64 = 2 * 24 * 3600

// GasIncreaseRate gas increase per IOST per second
var GasIncreaseRate = GasLimit.Sub(GasImmediateReward).Div(GasFulfillSeconds)

// UnpledgeFreezeSeconds coins will be frozen for 3 days after being unpledged
const UnpledgeFreezeSeconds int64 = 3 * 24 * 3600

var gasABIs *abiSet

// GasContractName the contract name
const GasContractName = "gas.iost"

func init() {
	gasABIs = newAbiSet()
	gasABIs.Register(initFunc, true)
	gasABIs.Register(constructor)
	gasABIs.Register(pledgeGas)
	gasABIs.Register(unpledgeGas)
	gasABIs.Register(rewardTGas)
	gasABIs.Register(transferTGas)
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
			if newPledge.LessThan(GasMinPledgeOfUser) {
				return finalCost, fmt.Errorf("unpledge to much %v - %v < %v", pledged.ToString(), unpledgeAmount.ToString(), GasMinPledgeOfUser.ToString())
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
	cost, _ = h.GasManager.RefreshPGas(name)
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
			if pledgeAmount.LessThan(GasMinPledgePerAction) {
				return nil, cost, fmt.Errorf("min pledge num is %d", GasMinPledgePerAction)
			}
			cost0, err = pledge(h, pledger, gasUser, pledgeAmount)
			cost.AddAssign(cost0)
			if err != nil {
				return nil, cost, err
			}
			contractName, cost0 := h.ContractName()
			cost.AddAssign(cost0)
			_, cost0, err = h.Call("token.iost", "transfer", fmt.Sprintf(`["iost", "%v", "%v", "%v", ""]`, pledger, contractName, pledgeAmountStr))
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
			if unpledgeAmount.LessThan(GasMinPledgePerAction) {
				return nil, cost, fmt.Errorf("min unpledge num is %d", GasMinPledgePerAction)
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
	rewardTGas = &abi{
		name: "reward",
		args: []string{"string", "string"},
		do: func(h *host.Host, args ...interface{}) (rtn []interface{}, cost contract.Cost, err error) {
			cost = contract.Cost0()
			//fmt.Println("context:" + h.Context().String())
			ok, cost0 := h.RequireAuth("auth.iost", "active")
			cost.AddAssign(cost0)
			if !ok {
				return nil, cost, fmt.Errorf("reward can only be called within auth.iost")
			}
			user := args[0].(string)
			if !h.IsValidAccount(user) {
				return nil, cost, fmt.Errorf("invalid user name %v", args[1])
			}
			f, err := common.NewFixed(args[1].(string), database.GasDecimal)
			if err != nil {
				return nil, cost, fmt.Errorf("invalid reward amount %v", err)
			}
			cost0 = h.ChangeTGas(user, f)
			cost.AddAssign(cost0)
			tgas, cost := h.TGas(user)
			cost.AddAssign(cost)
			return []interface{}{tgas.ToString()}, cost, nil
		},
	}
	transferTGas = &abi{
		name: "transfer",
		args: []string{"string", "string", "string"},
		do: func(h *host.Host, args ...interface{}) (rtn []interface{}, cost contract.Cost, err error) {
			cost = contract.Cost0()
			from := args[0].(string)
			if !h.IsValidAccount(from) {
				return nil, cost, fmt.Errorf("invalid user name %v", from)
			}
			to := args[1].(string)
			if !h.IsValidAccount(to) {
				return nil, cost, fmt.Errorf("invalid user name %v", to)
			}
			auth, cost0 := h.RequireAuth(from, "transfer")
			cost.AddAssign(cost0)
			if !auth {
				return nil, cost, host.ErrPermissionLost
			}
			f, err := common.NewFixed(args[2].(string), database.GasDecimal)
			if err != nil {
				return nil, cost, fmt.Errorf("invalid gas amount %v", err)
			}
			tGas, cost0 := h.TGas(from)
			cost.AddAssign(cost0)
			if tGas.LessThan(f) {
				return nil, cost, fmt.Errorf("transferable gas not enough %v < %v", tGas.ToString(), f.ToString())
			}
			cost0 = h.ChangeTGas(from, f.Neg())
			cost.AddAssign(cost0)
			cost0 = h.ChangeTGas(to, f)
			cost.AddAssign(cost0)
			return []interface{}{}, cost, nil
		},
	}
)
