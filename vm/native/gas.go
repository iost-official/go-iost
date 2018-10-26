package native

import (
	"errors"
	"fmt"
	"github.com/iost-official/go-iost/core/contract"
	"github.com/iost-official/go-iost/ilog"
	"github.com/iost-official/go-iost/vm/host"
	"strings"
)

const (
	// Every user must pledge a mininum amount of IOST (including GAS and RAM)
	minPledgeAmount = 100
	// Each IOST you pledge, you will get `gasImmediateRatio` gas immediately.
	// Then gas will be generated at a rate of `gasRateRatio` gas per block.
	// So after `pergasFillBlockNum` blocks (3 days here), you will reach `gasLimitRatio` gas.
	// Your gas production will stop because it reaches the limit.
	// When you use some gas later, the total amount will be less than the limit,
	// gas production will continue again util the limit.
	gasImmediateRatio = 10
	gasLimitRatio     = 30
	// gasFillBlockNum = 3 * 24 * 3600 / common.SlotLength
	gasRateRatio = 1 //(gasLimitRatio - gasImmediateRatio) / float64(gasFillBlockNum)
	// priv: 57nzU2rPGWNkA1EiShe3dmyPaocTmy2tfVtBK4oHh1MWrRWdQH3L688HsK2XUUiJpQsLJ3Dwc7uFQzmJBLJ7DXsZ
	gasAccount = "IOST2Jc2wHbHvQ2NGVdvJafoxo17GU2vdHqsJViAPuAzrV9v8zXsrS"
	//gasAccount = account.NewAccount(common.Base58Decode("5C9JWxSk6w8qpeow1tKK6owvQzxjBoVaSWTfcxmHqpnEcRGDX26T9px1ScXUKhsghUNwTvoxMxxcQoLdoZhSswkx"), crypto.Ed25519)
)

var gasABIs map[string]*abi

func init() {
	gasABIs = make(map[string]*abi)
	register(&gasABIs, createCoin)
	register(&gasABIs, issueCoin)
	register(&gasABIs, setCoinRate)
}

var (
	pledgeGas = &abi{
		name: "PledgeGas",
		args: []string{"string", "string"},
		do: func(h *host.Host, args ...interface{}) (rtn []interface{}, cost *contract.Cost, err error) {
			cost = contract.Cost0()
			userName, ok := args[0].(string)
			if !ok {
				return nil, host.CommonErrorCost(1), fmt.Errorf("invalid user name %s", args[0])
			}
			if !strings.HasPrefix(userName, "IOST") {
				return nil, host.CommonErrorCost(1), errors.New("userName should start with IOST")
			}
			pledgeAmount, ok := args[1].(int64)
			if !ok {
				return nil, host.CommonErrorCost(1), fmt.Errorf("invalid amount %s", args[1])
			}
			var minPledgeAmount int64 = 1
			if pledgeAmount < minPledgeAmount {
				return nil, host.CommonErrorCost(1), fmt.Errorf("min pledge num is %d", minPledgeAmount)
			}
			balance := h.DB().Balance(userName)
			if balance < pledgeAmount {
				return nil, host.CommonErrorCost(1), fmt.Errorf("balance not enough %d < %d", balance, pledgeAmount)
			}
			cost0, err := h.Teller.Transfer(userName, gasAccount, pledgeAmount)
			cost.AddAssign(cost0)
			if err != nil {
				return nil, cost, err
			}
			h.DB().SetGas(userName, h.DB().GetGas(userName)+pledgeAmount*gasImmediateRatio)
			h.DB().SetGasPledge(userName, h.DB().GetGasPledge(userName)+pledgeAmount)
			err = h.GasManager.ChangeGasRateAndLimit(userName, pledgeAmount*gasRateRatio, pledgeAmount*gasLimitRatio)
			cost.AddAssign(host.PledgeGasCost)
			if err != nil {
				return nil, cost, err
			}
			if ilog.GetLevel() < ilog.LevelDebug {
				ilog.Debugf("gas pledge: %s\n gas rate %d, gas limit %d, pledge %d",
					userName, h.DB().GetGasRate(userName), h.DB().GetGasLimit(userName), h.DB().GetGasPledge(userName))
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
			if !ok {
				return nil, host.CommonErrorCost(1), fmt.Errorf("invalid user name %s", args[0])
			}
			if !strings.HasPrefix(userName, "IOST") {
				return nil, host.CommonErrorCost(1), errors.New("userName should start with IOST")
			}
			unpledgeAmount, ok := args[1].(int64)
			if !ok {
				return nil, host.CommonErrorCost(1), fmt.Errorf("invalid amount %s", args[1])
			}
			var minPledgeAmount int64 = 1
			if unpledgeAmount < minPledgeAmount {
				return nil, host.CommonErrorCost(1), fmt.Errorf("min pledge num is %d", minPledgeAmount)
			}
			pledged := h.DB().GetGasPledge(userName)
			if pledged < unpledgeAmount+minPledgeAmount {
				return nil, host.CommonErrorCost(1), fmt.Errorf("you can unpledge %d most", (pledged - minPledgeAmount))
			}
			err = h.GasManager.ChangeGasRateAndLimit(userName, -unpledgeAmount*gasRateRatio, -unpledgeAmount*gasLimitRatio)
			cost.AddAssign(host.PledgeGasCost)
			if err != nil {
				return nil, cost, err
			}
			h.DB().SetGasPledge(userName, h.DB().GetGasPledge(userName)-unpledgeAmount)
			// TODO fix the account here
			err = h.Teller.TransferWithoutCheckPermission(gasAccount, userName, unpledgeAmount)
			if err != nil {
				return nil, cost, err
			}
			if ilog.GetLevel() < ilog.LevelDebug {
				ilog.Debugf("gas unpledge %s\n gas rate %d, gas limit %d, pledge %d",
					userName, h.DB().GetGasRate(userName), h.DB().GetGasLimit(userName), h.DB().GetGasPledge(userName))
			}
			return []interface{}{}, cost, nil
		},
	}
)
