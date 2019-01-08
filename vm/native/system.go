package native

import (
	"errors"

	"encoding/json"

	"github.com/bitly/go-simplejson"
	"github.com/iost-official/go-iost/core/contract"
	"github.com/iost-official/go-iost/vm/host"
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
}

// var .
var (
	requireAuth = &abi{
		name: "RequireAuth",
		args: []string{"string", "string"},
		do: func(h *host.Host, args ...interface{}) (rtn []interface{}, cost contract.Cost, err error) {
			var b bool
			b, cost = h.RequireAuth(args[0].(string), args[1].(string))
			rtn = []interface{}{
				b,
			}
			return rtn, cost, nil
		},
	}
	receipt = &abi{
		name: "Receipt",
		args: []string{"string"},
		do: func(h *host.Host, args ...interface{}) (rtn []interface{}, cost contract.Cost, err error) {
			cost = h.Receipt(args[0].(string))
			return []interface{}{}, cost, nil
		},
	}
	// setcode can only be invoked in native vm, avoid updating contract during running
	setCode = &abi{
		name: "SetCode",
		args: []string{"string"},
		do: func(h *host.Host, args ...interface{}) (rtn []interface{}, cost contract.Cost, err error) {

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
			cost2, err := h.SetCode(con, publisher)
			cost.AddAssign(cost2)
			if err != nil {
				return nil, cost, err
			}

			cost2, err = h.MapPut("contract_owner", actID, publisher)
			cost.AddAssign(cost2)

			return []interface{}{actID}, cost, err
		},
	}
	// updateCode can only be invoked in native vm, avoid updating contract during running
	updateCode = &abi{
		name: "UpdateCode",
		args: []string{"string", "string"},
		do: func(h *host.Host, args ...interface{}) (rtn []interface{}, cost contract.Cost, err error) {
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

			cost1, err := h.UpdateCode(con, []byte(args[1].(string)))
			cost.AddAssign(cost1)
			return []interface{}{}, cost, err
		},
	}

	// initSetCode can only be invoked in genesis block, use specific id for deploying contract
	initSetCode = &abi{
		name: "InitSetCode",
		args: []string{"string", "string"},
		do: func(h *host.Host, args ...interface{}) (rtn []interface{}, cost contract.Cost, err error) {
			cost = contract.Cost0()

			if h.Context().Value("number").(int64) != 0 {
				return []interface{}{}, cost, errors.New("InitSetCode in normal block")
			}

			con := &contract.Contract{}
			err = con.B64Decode(args[1].(string))
			if err != nil {
				return nil, host.CommonErrorCost(1), err
			}

			actID := args[0].(string)
			con.ID = actID

			cost2, err := h.SetCode(con, "")
			cost.AddAssign(cost2)

			cost2, err = h.MapPut("contract_owner", actID, "admin")
			cost.AddAssign(cost2)

			return []interface{}{actID}, cost, err
		},
	}

	// updateNativeCode can only be invoked in native vm, avoid updating contract during running
	updateNativeCode = &abi{
		name: "UpdateNativeCode",
		args: []string{"string", "string", "string"},
		do: func(h *host.Host, args ...interface{}) (rtn []interface{}, cost contract.Cost, err error) {
			cost = contract.Cost0()
			con := &contract.Contract{}
			conID := args[0].(string)
			version := args[1].(string)
			codeRaw := args[2].(string)

			// check auth
			ok, cost0 := h.RequireAuth("admin", systemPermission)
			cost.AddAssign(cost0)
			if !ok {
				return nil, cost, errors.New("set host settings need admin@system permission")
			}

			cost.AddAssign(host.CommonOpCost(1))
			if version != "" {
				con = SystemContractABI(conID, version)
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
			return []interface{}{}, cost, err
		},
	}

	// cancelDelaytx cancels a delay transaction.
	cancelDelaytx = &abi{
		name: "CancelDelaytx",
		args: []string{"string"},
		do: func(h *host.Host, args ...interface{}) (rtn []interface{}, cost contract.Cost, err error) {

			cost, err = h.CancelDelaytx(args[0].(string))
			return []interface{}{}, cost, err
		},
	}

	// hostSettings set host json
	hostSettings = &abi{
		name: "hostSettings",
		args: []string{"string"},
		do: func(h *host.Host, args ...interface{}) (rtn []interface{}, cost contract.Cost, err error) {
			// check auth
			ok, cost0 := h.RequireAuth("admin", systemPermission)
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
