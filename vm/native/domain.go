package native

import (
	"errors"

	"github.com/bitly/go-simplejson"
	"github.com/iost-official/go-iost/core/contract"
	"github.com/iost-official/go-iost/vm/host"
)

var domainABIs map[string]*abi

func init() {
	domainABIs = make(map[string]*abi)
	register(&domainABIs, link)
	register(&domainABIs, transferURL)

}

var (
	link = &abi{
		name: "Link",
		args: []string{"string", "string"},
		do: func(h *host.Host, args ...interface{}) (rtn []interface{}, cost *contract.Cost, err error) {
			cost = contract.Cost0()
			url := args[0].(string)
			cid := args[1].(string)

			txInfo, c := h.TxInfo()
			cost.AddAssign(c)
			tij, err := simplejson.NewJson(txInfo)
			if err != nil {
				panic(err)
			}

			applicant := tij.Get("publisher").MustString()

			owner := h.DHCP.URLOwner(url)

			if owner != "" && owner != applicant {
				cost.AddAssign(host.CommonErrorCost(1))
				return nil, cost, errors.New("no privilege of claimed url")
			}

			ok, c := h.RequireAuth(applicant)
			cost.AddAssign(c)

			if !ok {
				return nil, cost, errors.New("no privilege of claimed url")
			}

			h.WriteLink(url, cid, applicant)
			cost.AddAssign(host.PutCost)
			cost.AddAssign(host.PutCost)
			cost.AddAssign(host.PutCost)

			return nil, cost, nil
		},
	}
	transferURL = &abi{
		name: "Transfer",
		args: []string{"string", "string"},
		do: func(h *host.Host, args ...interface{}) (rtn []interface{}, cost *contract.Cost, err error) {
			cost = contract.Cost0()
			url := args[0].(string)
			to := args[1].(string)

			txInfo, c := h.TxInfo()
			cost.AddAssign(c)
			tij, err := simplejson.NewJson(txInfo)
			if err != nil {
				panic(err)
			}

			applicant := tij.Get("publisher").MustString()

			owner := h.DHCP.URLOwner(url)

			if owner != "" && owner != applicant {
				cost.AddAssign(host.CommonErrorCost(1))
				return nil, cost, errors.New("no privilege of claimed url")
			}

			ok, c := h.RequireAuth(applicant)
			cost.AddAssign(c)

			if !ok {
				return nil, cost, errors.New("no privilege of claimed url")
			}

			h.URLTransfer(url, to)
			cost.AddAssign(host.PutCost)

			return nil, cost, nil

		},
	}
)
