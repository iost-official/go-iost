package native

import (
	"errors"
	"fmt"
	"strings"

	"github.com/iost-official/go-iost/core/contract"
	"github.com/iost-official/go-iost/vm/host"
)

const (
	// priv: 2yquS3ySrGWPEKywCPzX4RTJugqRh7kJSo5aehsLYPEWkUxBWA39oMrZ7ZxuM4fgyXYs2cPwh5n8aNNpH5x2VyK1
	gasAccount = "IOST2mCzj85xkSvMf1eoGtrexQcwE6gK8z5xr6Kc48DwxXPCqQJva4"
	//gasAccount = account.NewAccount(common.Base58Decode("5C9JWxSk6w8qpeow1tKK6owvQzxjBoVaSWTfcxmHqpnEcRGDX26T9px1ScXUKhsghUNwTvoxMxxcQoLdoZhSswkx"), crypto.Ed25519)
)

// IOSTRatio ...
const IOSTRatio int64 = 100000000

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
			auth, cost0 := h.RequireAuth(userName, "transfer")
			cost.AddAssign(cost0)
			if !auth {
				return nil, host.CommonErrorCost(1), host.ErrPermissionLost
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
			// TODO fix the account here
			err = h.Teller.TransferRaw(userName, gasAccount, pledgeAmount*IOSTRatio)
			cost.AddAssign(host.TransferCost)
			if err != nil {
				return nil, cost, err
			}
			err = h.GasManager.Pledge(userName, pledgeAmount)
			cost.AddAssign(host.PledgeForGasCost)
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
			auth, cost0 := h.RequireAuth(userName, "transfer")
			cost.AddAssign(cost0)
			if !auth {
				return nil, host.CommonErrorCost(1), host.ErrPermissionLost
			}
			unpledgeAmount, ok := args[1].(int64)
			if !ok {
				return nil, host.CommonErrorCost(1), fmt.Errorf("invalid amount %s", args[1])
			}
			var minPledgeAmount int64 = 1
			if unpledgeAmount < minPledgeAmount {
				return nil, host.CommonErrorCost(1), fmt.Errorf("min pledge num is %d", minPledgeAmount)
			}
			err = h.GasManager.Pledge(userName, -unpledgeAmount)
			cost.AddAssign(host.PledgeForGasCost)
			if err != nil {
				return nil, cost, err
			}
			// TODO fix the account here
			err = h.Teller.TransferRaw(gasAccount, userName, unpledgeAmount*IOSTRatio)
			cost.AddAssign(host.PledgeForGasCost)
			if err != nil {
				return nil, cost, err
			}
			return []interface{}{}, cost, nil
		},
	}
)
