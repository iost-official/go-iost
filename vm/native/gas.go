package native

import (
	"errors"
	"fmt"
	"strings"

	"github.com/iost-official/go-iost/common"
	"github.com/iost-official/go-iost/core/contract"
	"github.com/iost-official/go-iost/vm/host"
)

// IOSTRatio ...
const IOSTRatio int64 = 100000000

var gasMinPledgeInIOST int64 = 10

// GasMinPledge Every user must pledge a minimum amount of IOST (including GAS and RAM)
var GasMinPledge = &common.Fixed{Value: gasMinPledgeInIOST * IOSTRatio, Decimal: 8}

// Each IOST you pledge, you will get `GasImmediateReward` gas immediately.
// Then gas will be generated at a rate of `GasIncreaseRate` gas per block.
// Your gas production will stop when it reaches the limit.
// When you use some gas later, the total amount will be less than the limit,
// so gas production will continue again util the limit.

// GasImmediateReward immediate reward per IOST
var GasImmediateReward = &common.Fixed{Value: 10 * IOSTRatio, Decimal: 8}

// GasLimit gas limit per IOST
var GasLimit = &common.Fixed{Value: 30 * IOSTRatio, Decimal: 8}

// GasIncreaseRate gas increase per IOST per block
var GasIncreaseRate = &common.Fixed{Value: 1 * IOSTRatio, Decimal: 8}

var gasABIs map[string]*abi

func init() {
	gasABIs = make(map[string]*abi)
	register(gasABIs, constructor)
	register(gasABIs, initFunc)
	register(gasABIs, pledgeGas)
	register(gasABIs, unpledgeGas)
}

// Pledge Change all gas related storage here. If pledgeAmount > 0. pledge. If pledgeAmount < 0, unpledge.
func pledge(h *host.Host, name string, pledgeAmountF *common.Fixed) error {
	pledgeAmount := pledgeAmountF.Value
	if pledgeAmount == 0 {
		return fmt.Errorf("invalid pledge amount %v", pledgeAmount)
	}
	if pledgeAmount < 0 {
		unpledgeAmount := pledgeAmountF.Neg()
		pledged := h.DB().GasHandler.GasPledge(name)
		// how to deal with overflow here?
		if pledged.Sub(unpledgeAmount).LessThan(GasMinPledge) {
			return fmt.Errorf("%v should be pledged at least ", GasMinPledge)
		}
	}

	limitDelta := pledgeAmountF.Multiply(GasLimit)
	rateDelta := pledgeAmountF.Multiply(GasIncreaseRate)
	gasDelta := pledgeAmountF.Multiply(GasImmediateReward)
	if pledgeAmount < 0 {
		// unpledge should not change current generated gas
		gasDelta.Value = 0
	}
	fmt.Printf("limitd rated gasd %v %v %v\n", limitDelta, rateDelta, gasDelta)

	// pledge first time
	if h.DB().GasHandler.GasUpdateTime(name) == 0 {
		if pledgeAmount < 0 {
			return fmt.Errorf("cannot unpledge! No pledge before")
		}
		h.DB().GasHandler.SetGasPledge(name, pledgeAmountF)
		h.DB().GasHandler.SetGasUpdateTime(name, h.Context().Value("number").(int64))
		h.DB().GasHandler.SetGasRate(name, rateDelta)
		h.DB().GasHandler.SetGasLimit(name, limitDelta)
		h.DB().GasHandler.SetGasStock(name, gasDelta)
		return nil
	}
	h.GasManager.RefreshGas(name)
	rateOld := h.DB().GasHandler.GetGasRate(name)
	rateNew := rateOld.Add(rateDelta)
	if rateNew.Value <= 0 {
		return fmt.Errorf("change gasRate failed! current: %v, delta %v", rateOld, rateDelta)
	}
	limitOld := h.DB().GasHandler.GasLimit(name)
	limitNew := limitOld.Add(limitDelta)
	if limitNew.Value <= 0 {
		return fmt.Errorf("change gasLimit failed! current: %v, delta %v", limitOld, limitDelta)
	}
	gasOld := h.DB().GasHandler.GasStock(name)
	gasNew := gasOld.Add(gasDelta)
	if limitNew.LessThan(gasNew) {
		// clear the gas above the new limit.
		gasNew = limitNew
	}

	fmt.Printf("Pledge %v", pledgeAmountF)
	h.DB().GasHandler.SetGasPledge(name, h.DB().GasHandler.GasPledge(name).Add(pledgeAmountF))
	h.DB().GasHandler.SetGasRate(name, rateNew)
	h.DB().GasHandler.SetGasLimit(name, limitNew)
	h.DB().GasHandler.SetGasStock(name, gasNew)
	return nil
}

var (
	pledgeGas = &abi{
		name: "PledgeGas",
		args: []string{"string", "string"},
		do: func(h *host.Host, args ...interface{}) (rtn []interface{}, cost *contract.Cost, err error) {
			cost = contract.Cost0()
			userName, ok := args[0].(string)
			cost.AddAssign(host.CommonErrorCost(1))
			if !ok {
				return nil, cost, fmt.Errorf("invalid user name %s", args[0])
			}
			// todo check is account name
			// if !strings.HasPrefix(userName, "IOST") {
			// 	return nil, cost, errors.New("userName should start with IOST")
			// }
			auth, cost0 := h.RequireAuth(userName, "transfer")
			cost.AddAssign(cost0)
			if !auth {
				return nil, cost, host.ErrPermissionLost
			}
			pledgeAmountStr, ok := args[1].(string)
			if !ok {
				return nil, cost, fmt.Errorf("invalid amount %s", args[1])
			}
			pledgeAmount, ok := common.NewFixed(pledgeAmountStr, 8)
			cost.AddAssign(host.CommonErrorCost(1))
			if !ok || pledgeAmount.Value <= 0 {
				return nil, cost, fmt.Errorf("invalid amount %s", args[1])
			}
			var minPledgeAmount int64 = 1 * IOSTRatio
			if pledgeAmount.Value < minPledgeAmount {
				return nil, cost, fmt.Errorf("min pledge num is %d", minPledgeAmount)
			}
			contractName, cost0 := h.ContractName()
			cost.AddAssign(cost0)
			_, cost0, err = h.Call("iost.token", "transfer", fmt.Sprintf(`["iost", "%v", "%v", "%v"]`, userName, contractName, pledgeAmountStr))
			cost.AddAssign(cost0)
			if err != nil {
				return nil, cost, err
			}
			err = pledge(h, userName, pledgeAmount)
			cost.AddAssign(host.PledgeForGasCost)
			if err != nil {
				return nil, cost, err
			}
			return []interface{}{}, cost, nil
		},
	}
	unpledgeGas = &abi{
		name: "UnpledgeGas",
		args: []string{"string", "string"},
		do: func(h *host.Host, args ...interface{}) (rtn []interface{}, cost *contract.Cost, err error) {
			cost = contract.Cost0()
			userName, ok := args[0].(string)
			cost.AddAssign(host.CommonErrorCost(1))
			if !ok {
				return nil, cost, fmt.Errorf("invalid user name %s", args[0])
			}
			if !strings.HasPrefix(userName, "IOST") {
				return nil, cost, errors.New("userName should start with IOST")
			}
			auth, cost0 := h.RequireAuth(userName, "transfer")
			cost.AddAssign(cost0)
			if !auth {
				return nil, cost, host.ErrPermissionLost
			}
			unpledgeAmountStr, ok := args[1].(string)
			if !ok {
				return nil, cost, fmt.Errorf("invalid amount %s", args[1])
			}
			unpledgeAmount, ok := common.NewFixed(unpledgeAmountStr, 8)
			cost.AddAssign(host.CommonErrorCost(1))
			if !ok || unpledgeAmount.Value <= 0 {
				return nil, cost, fmt.Errorf("invalid amount %s", args[1])
			}
			var minUnpledgeAmount int64 = 1 * IOSTRatio
			if unpledgeAmount.Value < minUnpledgeAmount {
				return nil, cost, fmt.Errorf("min unpledge num is %d", minUnpledgeAmount)
			}
			err = pledge(h, userName, unpledgeAmount.Neg())
			cost.AddAssign(host.PledgeForGasCost)
			if err != nil {
				return nil, cost, err
			}
			contractName, cost0 := h.ContractName()
			cost.AddAssign(cost0)
			cost0, err = h.Withdraw(userName, unpledgeAmountStr)
			_, cost0, err = h.Call("iost.token", "transfer", fmt.Sprintf(`["iost", "%v", "%v", "%v"]`,  contractName, userName, unpledgeAmountStr))
			cost.AddAssign(cost0)
			if err != nil {
				return nil, cost, err
			}
			return []interface{}{}, cost, nil
		},
	}
)
