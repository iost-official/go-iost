package host

import (
	"encoding/json"

	"github.com/iost-official/go-iost/account"
	"github.com/iost-official/go-iost/core/contract"
	"strings"
)

// Authority module of ...
type Authority struct {
	h *Host
}

// RequireAuth check auth
func (h *Authority) RequireAuth(id, p string) (bool, *contract.Cost) {
	authList := h.h.ctx.Value("auth_list")
	authMap := authList.(map[string]int)
	// add current contract auth
	authMap[h.h.ctx.Value("contract_name").(string)] = 1
	reenterMap := make(map[string]int)

	return h.requireAuth(id, p, authMap, reenterMap)
}

// ReadAuth read auth
func (h *Authority) isContract(id string) bool {
	// todo tell apart contractid and accountid
	if strings.HasPrefix(id, "Contract") || strings.Contains(id, ".") {
		return true
	}
	return false
}

// ReadAuth read auth
func (h *Authority) ReadAuth(id string) (*account.Account, *contract.Cost) {
	if h.isContract(id) {
		return account.NewInitAccount(id, id, id), CommonOpCost(1)
	}
	acc, cost := h.h.GlobalMapGet("iost.auth", "account", id)
	if acc == nil {
		return nil, cost
	}
	var a account.Account
	err := json.Unmarshal([]byte(acc.(string)), &a)
	if err != nil {
		panic(err)
	}
	return &a, cost
}

func (h *Authority) requireAuth(id, permission string, auth, reenter map[string]int) (bool, *contract.Cost) {
	if _, ok := reenter[id+"@"+permission]; ok {
		return false, CommonErrorCost(1)
	}
	reenter[id+"@"+permission] = 1

	a, c := h.ReadAuth(id)
	if a == nil {
		return false, c
	}

	p, ok := a.Permissions[permission]
	if !ok {
		p = a.Permissions["active"]
	}

	u := p.Users
	for _, g := range p.Groups {
		grp, ok := a.Groups[g]
		if !ok {
			continue
		}
		u = append(u, grp.Users...)
	}

	var weight int
	for _, user := range u {
		if user.IsKeyPair {
			if _, ok := auth[user.ID]; ok {
				weight += user.Weight
				if weight >= p.Threshold {
					return true, c
				}
			}
		} else {
			ok, cost := h.requireAuth(user.ID, user.Permission, auth, reenter)
			c.AddAssign(cost)
			if ok {
				weight += user.Weight
				if weight >= p.Threshold {
					return true, c
				}
			}
		}
	}

	return weight >= p.Threshold, c
}
