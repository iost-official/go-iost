package native

import (
	"errors"

	"fmt"

	"github.com/bitly/go-simplejson"
	"github.com/iost-official/go-iost/core/contract"
	"github.com/iost-official/go-iost/vm/host"
	"strings"
)

// DomainABIs list of domain abi
var domainABIs *abiSet
var domain0ABIs *abiSet

func init() {
	domainABIs = newAbiSet()
	domainABIs.Register(initDomainABI, true)
	domainABIs.Register(linkDomainABI)
	domainABIs.Register(transferDomainABI)
	domainABIs.Register(canUpdateDomainABI)
	domain0ABIs = newAbiSet()
	domain0ABIs.Register(initDomainABI, true)
	domain0ABIs.Register(canUpdateDomainABI)
}

func checkURLValid(name string) error {
	if len(name) < 5 || len(name) > 16 {
		return fmt.Errorf("url invalid. url length should be between 5,16 got %v", name)
	}
	if !strings.Contains(name, ".") {
		return fmt.Errorf("url invalid. url must contain '.'")
	}
	for _, ch := range name {
		if !(ch >= 'a' && ch <= 'z' || ch >= '0' && ch <= '9' || ch == '_' || ch == '.') {
			return fmt.Errorf("url invalid. url contains invalid character %v", ch)
		}
	}
	return nil
}

var (
	canUpdateDomainABI = &abi{
		name: "can_update",
		args: []string{"string"},
		do: func(h *host.Host, args ...interface{}) (rtn []interface{}, cost contract.Cost, err error) {
			ok, cost := h.RequireAuth("admin", "system.iost")
			return []interface{}{ok}, cost, nil
		},
	}
	initDomainABI = &abi{
		name: "init",
		args: []string{},
		do: func(h *host.Host, args ...interface{}) (rtn []interface{}, cost contract.Cost, err error) {
			return []interface{}{}, host.CommonErrorCost(1), nil
		},
	}

	linkDomainABI = &abi{
		name: "Link",
		args: []string{"string", "string"},
		do: func(h *host.Host, args ...interface{}) (rtn []interface{}, cost contract.Cost, err error) {
			cost = contract.Cost0()
			url := args[0].(string)
			cid := args[1].(string)

			cost.AddAssign(host.CommonOpCost(1))
			err = checkURLValid(url)
			if err != nil {
				return nil, cost, err
			}
			if strings.HasSuffix(url, ".iost") {
				ok, c := h.RequireAuth("admin", "domain.iost")
				cost.AddAssign(c)
				if !ok {
					return nil, cost, errors.New("only admin has permission to claim url .iost")
				}
			}

			txInfo, c := h.TxInfo()
			cost.AddAssign(c)
			tij, err := simplejson.NewJson(txInfo)
			if err != nil {
				panic(err)
			}

			applicant := tij.Get("publisher").MustString()

			owner := h.DNS.URLOwner(url)

			if owner != "" && owner != applicant {
				cost.AddAssign(host.CommonErrorCost(1))
				return nil, cost, errors.New("no privilege of claimed url")
			}

			ok, c := h.RequireAuth(applicant, "domain.iost")
			cost.AddAssign(c)

			if !ok {
				return nil, cost, errors.New("no permission of claimed url")
			}

			h.WriteLink(url, cid, applicant)
			cost.AddAssign(host.Costs["PutCost"])
			cost.AddAssign(host.Costs["PutCost"])
			cost.AddAssign(host.Costs["PutCost"])

			return nil, cost, nil
		},
	}
	transferDomainABI = &abi{
		name: "Transfer",
		args: []string{"string", "string"},
		do: func(h *host.Host, args ...interface{}) (rtn []interface{}, cost contract.Cost, err error) {
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

			owner := h.DNS.URLOwner(url)

			if owner == "" {
				cost.AddAssign(host.CommonErrorCost(1))
				return nil, cost, errors.New("url doesn't have owner. Link directly")
			}

			if owner != applicant {
				cost.AddAssign(host.CommonErrorCost(1))
				return nil, cost, errors.New("no privilege of claimed url")
			}

			ok, c := h.RequireAuth(applicant, "domain.iost")
			cost.AddAssign(c)

			if !ok {
				return nil, cost, errors.New("no permission of claimed url")
			}

			h.URLTransfer(url, to)
			cost.AddAssign(host.Costs["PutCost"])

			return nil, cost, nil

		},
	}
)
