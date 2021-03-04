package host

import (
	"encoding/json"
	"strings"

	"github.com/iost-official/go-iost/v3/account"
	"github.com/iost-official/go-iost/v3/core/contract"
	"github.com/iost-official/go-iost/v3/vm/database"
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
	if h.h.IsFork3_1_0 {
		if _, ok := authContractList[id]; ok {
			return true, cost
		}
	} else {
		if _, ok := authContractList[id]; ok || h.h.ctx.Value("contract_name").(string) == id {
			return true, cost
		}
	}
	return false, cost
}

func (h *Authority) requireAuth(id, p string, authType int) (bool, contract.Cost) {
	if i, ok := h.h.ctx.Value("number").(int64); ok && i == 0 {
		return true, contract.Cost0()
	}
	if h.IsContract(id) {
		return h.requireContractAuth(id, p)
	}
	authList := h.h.ctx.Value("auth_list")
	authMap := authList.(map[string]int)

	signerList := h.h.ctx.Value("signer_list")
	signerMap := signerList.(map[string]bool)

	reenterMap := make(map[string]int)

	return auth(h.h.db, id, p, authMap, reenterMap, signerMap, authType)
}

// RequireAuth check auth
func (h *Authority) RequireAuth(id, p string) (bool, contract.Cost) {
	if h.h.IsFork3_3_0 {
		return h.requireAuth(id, p, authNormal)
	}
	return h.requireAuth(id, p, authSigner)
}

// RequireSignerAuth check signer auth
func (h *Authority) RequireSignerAuth(id, p string) (bool, contract.Cost) {
	return h.requireAuth(id, p, authSigner)
}

// RequirePublisherAuth check publisher auth
func (h *Authority) RequirePublisherAuth(id string) (bool, contract.Cost) {
	return h.requireAuth(id, "active", authPublisher)
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

const (
	authNormal int = iota
	authSigner
	authPublisher
)

func auth(vi *database.Visitor, id, permission string, authMap, reenter map[string]int, signerMap map[string]bool, authType int) (bool, contract.Cost) { // nolint
	if _, ok := reenter[id+"@"+permission]; ok {
		return false, CommonErrorCost(1)
	}
	reenter[id+"@"+permission] = 1

	if authType == authNormal && signerMap[id+"@"+permission] {
		return true, CommonOpCost(1)
	}

	a, c := ReadAuth(vi, id)

	if a == nil {
		return false, c
	}

	p, ok := a.Permissions[permission]
	if !ok {
		if permission == "owner" || permission == "active" {
			return false, c
		}
		return auth(vi, id, "active", authMap, reenter, signerMap, authType)
	}

	u := p.Items
	for _, g := range p.Groups {
		grp, ok := a.Groups[g]
		if !ok {
			continue
		}
		u = append(u, grp.Items...)
	}

	var atype int
	if authType == authPublisher {
		atype = 1
	} else {
		atype = 0
	}

	var weight int
	for _, user := range u {
		if user.IsKeyPair {
			if authType == authNormal {
				continue
			}
			if i, ok := authMap[user.ID]; ok && i > atype {
				weight += user.Weight
				if weight >= p.Threshold {
					return true, c
				}
			}
		} else {
			ok, cost := auth(vi, user.ID, user.Permission, authMap, reenter, signerMap, authType)
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
		ok, c2 := auth(vi, id, "owner", authMap, reenter, signerMap, authType)
		c.AddAssign(c2)
		return ok, c
	} else if permission == "owner" {
		return false, c
	} else {
		ok, c2 := auth(vi, id, "active", authMap, reenter, signerMap, authType)
		c.AddAssign(c2)
		return ok, c
	}
}
