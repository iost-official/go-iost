package native

import (
	"encoding/json"
	"errors"

	"github.com/bitly/go-simplejson"
	"github.com/iost-official/go-iost/v3/core/contract"
	"github.com/iost-official/go-iost/v3/vm/host"
)

var systemABIs *abiSet

func init() {
	systemABIs = newAbiSet()
	systemABIs.Register(requireAuth)
	systemABIs.Register(receipt)
	systemABIs.Register(setCode)
	systemABIs.Register(updateCode)
	systemABIs.Register(initSetCode)
	systemABIs.Register(cancelDelaytx)
	systemABIs.Register(hostSettings)
	systemABIs.Register(updateNativeCode)
}

// var .
var (
	requireAuth = &abi{
		name: "requireAuth",
		args: []string{"string", "string"},
		do: func(h *host.Host, args ...any) (rtn []any, cost contract.Cost, err error) {
			var b bool
			b, cost = h.RequireAuth(args[0].(string), args[1].(string))
			rtn = []any{
				b,
			}
			return rtn, cost, nil
		},
	}
	receipt = &abi{
		name: "receipt",
		args: []string{"string"},
		do: func(h *host.Host, args ...any) (rtn []any, cost contract.Cost, err error) {
			cost = h.Receipt(args[0].(string))
			return []any{}, cost, nil
		},
	}
	// setcode can only be invoked in native vm, avoid updating contract during running
	setCode = &abi{
		name: "setCode",
		args: []string{"string"},
		do: func(h *host.Host, args ...any) (rtn []any, cost contract.Cost, err error) {

			cost = contract.Cost0()
			con := &contract.Contract{}
			codeRaw := args[0].(string)

			if codeRaw[0] == '{' {
				err = json.Unmarshal([]byte(codeRaw), con)
				if err != nil {
					return nil, host.CommonErrorCost(1), err
				}
			} else {
				err = con.B64Decode(codeRaw)
				if err != nil {
					return nil, host.CommonErrorCost(1), err
				}
			}

			info, cost1 := h.TxInfo()
			cost.AddAssign(cost1)
			var json *simplejson.Json
			json, err = simplejson.NewJson(info)
			if err != nil {
				return nil, cost, err
			}

			var id string
			id, err = json.Get("hash").String()
			if err != nil {
				return nil, cost, err
			}
			actID := "Contract" + id
			con.ID = actID

			publisher := h.Context().Value("publisher").(string)

			cost.AddAssign(host.SetCodeCost(len(con.Code)))
			if !CheckCost(h, cost) {
				return nil, cost, host.ErrOutOfGas
			}
			cost2, err := h.SetCode(con, publisher)
			cost.AddAssign(cost2)
			if err != nil {
				return nil, cost, err
			}

			cost2, err = h.MapPut("contract_owner", actID, publisher, publisher)
			cost.AddAssign(cost2)

			return []any{actID}, cost, err
		},
	}
	// updateCode can only be invoked in native vm, avoid updating contract during running
	updateCode = &abi{
		name: "updateCode",
		args: []string{"string", "string"},
		do: func(h *host.Host, args ...any) (rtn []any, cost contract.Cost, err error) {
			cost = contract.Cost0()
			con := &contract.Contract{}
			codeRaw := args[0].(string)

			cost.AddAssign(host.CommonOpCost(1))
			stackHeight := h.Context().Value("stack_height").(int)
			if stackHeight != 1 {
				return nil, cost, errors.New("can't call UpdateCode from other contract")
			}

			if codeRaw[0] == '{' {
				err = json.Unmarshal([]byte(codeRaw), con)
				if err != nil {
					return nil, host.CommonErrorCost(1), err
				}
			} else {
				err = con.B64Decode(codeRaw)
				if err != nil {
					return nil, host.CommonErrorCost(1), err
				}
			}

			cost.AddAssign(host.SetCodeCost(len(con.Code)))
			if !CheckCost(h, cost) {
				return nil, cost, host.ErrOutOfGas
			}

			cost1, err := h.UpdateCode(con, []byte(args[1].(string)))
			cost.AddAssign(cost1)
			return []any{}, cost, err
		},
	}

	// initSetCode can only be invoked in genesis block, use specific id for deploying contract
	initSetCode = &abi{
		name: "initSetCode",
		args: []string{"string", "string"},
		do: func(h *host.Host, args ...any) (rtn []any, cost contract.Cost, err error) {
			cost = contract.Cost0()

			if h.Context().Value("number").(int64) != 0 {
				return []any{}, cost, errors.New("initSetCode in normal block")
			}

			con := &contract.Contract{}
			err = con.B64Decode(args[1].(string))
			if err != nil {
				return nil, host.CommonErrorCost(1), err
			}

			actID := args[0].(string)
			con.ID = actID

			cost2, err := h.SetCode(con, AdminAccount)
			cost.AddAssign(cost2)
			if err != nil {
				return nil, cost, err
			}

			cost2, err = h.MapPut("contract_owner", actID, AdminAccount)
			cost.AddAssign(cost2)

			return []any{actID}, cost, err
		},
	}

	// updateNativeCode can only be invoked in native vm, avoid updating contract during running
	updateNativeCode = &abi{
		name: "updateNativeCode",
		args: []string{"string", "string", "string"},
		do: func(h *host.Host, args ...any) (rtn []any, cost contract.Cost, err error) {
			cost = contract.Cost0()
			con := &contract.Contract{}
			conID := args[0].(string)
			version := args[1].(string)
			codeRaw := args[2].(string)

			// check auth
			ok, cost0 := h.RequireAuth(AdminAccount, SystemPermission)
			cost.AddAssign(cost0)
			if !ok {
				return nil, cost, errors.New("update native code need admin@system permission")
			}

			cost.AddAssign(host.CommonOpCost(1))
			if version != "" {
				con = SystemContractABI(conID, version)
				if con == nil {
					return nil, cost, errors.New("invalid contractID or version")
				}
			} else {
				if codeRaw[0] == '{' {
					err = json.Unmarshal([]byte(codeRaw), con)
					if err != nil {
						return nil, host.CommonErrorCost(1), err
					}
				} else {
					err = con.B64Decode(codeRaw)
					if err != nil {
						return nil, host.CommonErrorCost(1), err
					}
				}
			}

			cost0, err = h.UpdateCode(con, []byte(""))
			cost.AddAssign(cost0)
			return []any{}, cost, err
		},
	}

	// cancelDelaytx cancels a delay transaction.
	cancelDelaytx = &abi{
		name: "cancelDelaytx",
		args: []string{"string"},
		do: func(h *host.Host, args ...any) (rtn []any, cost contract.Cost, err error) {
			return []any{}, host.Costs["GetCost"], errors.New("delaytx not exists")
		},
	}

	// hostSettings set host json
	hostSettings = &abi{
		name: "hostSettings",
		args: []string{"string"},
		do: func(h *host.Host, args ...any) (rtn []any, cost contract.Cost, err error) {
			// check auth
			ok, cost0 := h.RequireAuth(AdminAccount, SystemPermission)
			cost.AddAssign(cost0)
			if !ok {
				return nil, cost, errors.New("set host settings need admin@system permission")
			}

			cost0, _ = h.MapPut("settings", "host", args[0])
			cost.AddAssign(cost0)
			return nil, cost, nil
		},
	}
)
