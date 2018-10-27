package native

import (
	"errors"
	"fmt"
	"github.com/iost-official/go-iost/core/contract"
	"github.com/iost-official/go-iost/vm/host"
	"strings"
)

const (
	// Every user must pledge a mininum amount of IOST (including GAS and RAM)
	gasMinPledgeAmount = 100
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
	// priv: 2yquS3ySrGWPEKywCPzX4RTJugqRh7kJSo5aehsLYPEWkUxBWA39oMrZ7ZxuM4fgyXYs2cPwh5n8aNNpH5x2VyK1
	gasAccount = "IOST2mCzj85xkSvMf1eoGtrexQcwE6gK8z5xr6Kc48DwxXPCqQJva4"
	//gasAccount = account.NewAccount(common.Base58Decode("5C9JWxSk6w8qpeow1tKK6owvQzxjBoVaSWTfcxmHqpnEcRGDX26T9px1ScXUKhsghUNwTvoxMxxcQoLdoZhSswkx"), crypto.Ed25519)
)

var gasABIs map[string]*abi

func init() {
	gasABIs = make(map[string]*abi)
	register(&gasABIs, constructor)
	register(&gasABIs, initFunc)
	register(&gasABIs, pledgeGas)
	register(&gasABIs, unpledgeGas)
}

var (
	pledgeGas = &abi{
		name: "PledgeGas",
		args: []string{"string", "number"},
		do: func(h *host.Host, args ...interface{}) (rtn []interface{}, cost *contract.Cost, err error) {
			cost = contract.Cost0()
			userName, ok := args[0].(string)
			if !ok {
				return nil, host.CommonErrorCost(1), fmt.Errorf("invalid user name %s", args[0])
			}
			if !strings.HasPrefix(userName, "IOST") {
				return nil, host.CommonErrorCost(1), errors.New("userName should start with IOST")
			}
			auth, cost0 := h.RequireAuth(userName)
			cost.AddAssign(cost0)
			if !auth {
				return nil, host.CommonErrorCost(1), errors.New("wtf") //host.ErrPermissionLost
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
			cost0, err = h.Teller.Transfer(userName, gasAccount, pledgeAmount)
			cost.AddAssign(cost0)
			if err != nil {
				return nil, cost, err
			}
			h.DB().SetGasPledge(userName, h.DB().GetGasPledge(userName)+pledgeAmount)
			err = h.GasManager.ChangeGas(userName, pledgeAmount*gasImmediateRatio, pledgeAmount*gasRateRatio, pledgeAmount*gasLimitRatio)
			cost.AddAssign(host.PledgeGasCost)
			if err != nil {
				return nil, cost, err
			}
			return []interface{}{}, cost, nil
		},
	}
	unpledgeGas = &abi{
		name: "UnpledgeGas",
		args: []string{"string", "number"},
		do: func(h *host.Host, args ...interface{}) (rtn []interface{}, cost *contract.Cost, err error) {
			cost = contract.Cost0()
			userName, ok := args[0].(string)
			if !ok {
				return nil, host.CommonErrorCost(1), fmt.Errorf("invalid user name %s", args[0])
			}
			if !strings.HasPrefix(userName, "IOST") {
				return nil, host.CommonErrorCost(1), errors.New("userName should start with IOST")
			}
			auth, cost0 := h.RequireAuth(userName)
			cost.AddAssign(cost0)
			if !auth {
				return nil, host.CommonErrorCost(1), errors.New("haha") //host.ErrPermissionLost
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
			if pledged < unpledgeAmount+gasMinPledgeAmount {
				return nil, host.CommonErrorCost(1), fmt.Errorf("you can unpledge %d most", (pledged - minPledgeAmount))
			}
			err = h.GasManager.ChangeGas(userName, 0, -unpledgeAmount*gasRateRatio, -unpledgeAmount*gasLimitRatio)
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
			return []interface{}{}, cost, nil
		},
	}
)
