package host

import (
	"encoding/json"
	"strings"

	"github.com/iost-official/go-iost/account"
	"github.com/iost-official/go-iost/core/contract"
	"github.com/iost-official/go-iost/vm/database"
)

// Authority module of ...
type Authority struct {
	h *Host
}

func (h *Authority) requireContractAuth(id, p string) (bool, contract.Cost) {
	if i, ok := h.h.ctx.Value("number").(int64); ok && i == 0 {
		return true, contract.Cost0()
	}
	cost := CommonOpCost(1)
	authContractList := h.h.ctx.Value("auth_contract_list").(map[string]int)
	if _, ok := authContractList[id]; ok || h.h.ctx.Value("contract_name").(string) == id {
		return true, cost
	}
	return false, cost
}

func (h *Authority) requireAuth(id, p string, isPublisher bool) (bool, contract.Cost) {
	if i, ok := h.h.ctx.Value("number").(int64); ok && i == 0 {
		return true, contract.Cost0()
	}
	if h.IsContract(id) {
		return h.requireContractAuth(id, p)
	}
	authList := h.h.ctx.Value("auth_list")
	authMap := authList.(map[string]int)
	reenterMap := make(map[string]int)

	if isPublisher {
		return AuthPublisher(h.h.db, id, p, authMap, reenterMap)
	}
	return Auth(h.h.db, id, p, authMap, reenterMap)
}

// RequireAuth check auth
func (h *Authority) RequireAuth(id, p string) (bool, contract.Cost) {
	return h.requireAuth(id, p, false)
}

// RequirePublisherAuth check publisher auth
func (h *Authority) RequirePublisherAuth(id string) (bool, contract.Cost) {
	return h.requireAuth(id, "active", true)
}

// IsContract to judge the id is contract format
func (h *Authority) IsContract(id string) bool {
	// todo tell apart contractid and accountid
	if strings.HasPrefix(id, "Contract") || strings.Contains(id, ".") {
		return true
	}
	return false
}

// ReadAuth read auth
func ReadAuth(vi *database.Visitor, id string) (*account.Account, contract.Cost) {
	sa := vi.MGet("auth.iost"+"-auth", id)
	acc := database.MustUnmarshal(sa)
	c := contract.NewCost(0, 0, int64(len(sa)))
	if acc == nil {
		return nil, c
	}
	var a account.Account
	err := json.Unmarshal([]byte(acc.(string)), &a)
	if err != nil {
		panic(err)
	}
	return &a, c
}

func auth(vi *database.Visitor, id, permission string, auth, reenter map[string]int, publisherOnly bool) (bool, contract.Cost) {
	if _, ok := reenter[id+"@"+permission]; ok {
		return false, CommonErrorCost(1)
	}
	reenter[id+"@"+permission] = 1

	a, c := ReadAuth(vi, id)

	if a == nil {
		return false, c
	}

	p, ok := a.Permissions[permission]
	if !ok {
		if permission == "owner" || permission == "active" {
			return false, c
		}
		return Auth(vi, id, "active", auth, reenter)
	}

	u := p.Items
	for _, g := range p.Groups {
		grp, ok := a.Groups[g]
		if !ok {
			continue
		}
		u = append(u, grp.Items...)
	}

	var authtype int
	if publisherOnly {
		authtype = 1
	} else {
		authtype = 0
	}

	var weight int
	for _, user := range u {
		if user.IsKeyPair {
			if i, ok := auth[user.ID]; ok && i > authtype {
				weight += user.Weight
				if weight >= p.Threshold {
					return true, c
				}
			}
		} else {
			ok, cost := Auth(vi, user.ID, user.Permission, auth, reenter)
			c.AddAssign(cost)
			if ok {
				weight += user.Weight
				if weight >= p.Threshold {
					return true, c
				}
			}
		}
	}

	if weight >= p.Threshold {
		return true, c
	}
	if permission == "active" {
		ok, c2 := Auth(vi, id, "owner", auth, reenter)
		c.AddAssign(c2)
		return ok, c
	} else if permission == "owner" {
		return false, c
	} else {
		ok, c2 := Auth(vi, id, "active", auth, reenter)
		c.AddAssign(c2)
		return ok, c
	}
}

// Auth check auth
func Auth(vi *database.Visitor, id, permission string, authMap, reenter map[string]int) (bool, contract.Cost) { // nolint
	return auth(vi, id, permission, authMap, reenter, false)
}

// AuthPublisher check publisher auth
func AuthPublisher(vi *database.Visitor, id, permission string, authMap, reenter map[string]int) (bool, contract.Cost) { // nolint
	return auth(vi, id, permission, authMap, reenter, true)
}
