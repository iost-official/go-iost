package native

import (
	"errors"

	"github.com/bitly/go-simplejson"
	"github.com/iost-official/go-iost/core/contract"
	"github.com/iost-official/go-iost/vm/host"
)

var systemABIs map[string]*abi

func init() {
	systemABIs = make(map[string]*abi)
	register(systemABIs, requireAuth)
	register(systemABIs, receipt)
	register(systemABIs, setCode)
	register(systemABIs, updateCode)
	register(systemABIs, destroyCode)
	register(systemABIs, initSetCode)
	register(systemABIs, cancelDelaytx)
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
			err = con.B64Decode(args[0].(string))
			if err != nil {
				return nil, host.CommonErrorCost(1), err
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

			cost2 = h.MapPut("contract_owner", actID, publisher)
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
			err = con.B64Decode(args[0].(string))
			if err != nil {
				return nil, host.CommonErrorCost(1), err
			}

			cost1, err := h.UpdateCode(con, []byte(args[1].(string)))
			cost.AddAssign(cost1)
			return []interface{}{}, cost, err
		},
	}
	// todo deprecated
	// destroyCode can only be invoked in native vm, avoid updating contract during running
	destroyCode = &abi{
		name: "DestroyCode",
		args: []string{"string"},
		do: func(h *host.Host, args ...interface{}) (rtn []interface{}, cost contract.Cost, err error) {

			cost, err = h.DestroyCode(args[0].(string))
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
			return []interface{}{actID}, cost, err
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
)
