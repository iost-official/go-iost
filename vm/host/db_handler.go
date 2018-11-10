package host

import (
	"github.com/iost-official/go-iost/core/contract"
	"github.com/iost-official/go-iost/vm/database"
)

// DBHandler struct {
type DBHandler struct {
	h *Host
}

// NewDBHandler ...
func NewDBHandler(h *Host) DBHandler {
	return DBHandler{
		h: h,
	}
}

// Put put kv to db
func (h *DBHandler) Put(key string, value interface{}, owner ...string) contract.Cost {
	sv := database.MustMarshal(value)
	var mk = h.modifyKey(key, owner...)
	h.payRAM(mk, sv, owner...)
	h.h.db.Put(mk, sv)
	return PutCost
}

// Get get value of key from db
func (h *DBHandler) Get(key string, owner ...string) (value interface{}, cost contract.Cost) {
	var mk = h.modifyKey(key, owner...)
	rtn := database.MustUnmarshal(h.h.db.Get(mk))
	return rtn, GetCost
}

// Del delete key
func (h *DBHandler) Del(key string, owner ...string) contract.Cost {
	var mk = h.modifyKey(key, owner...)
	h.releaseRAM(mk, owner...)
	h.h.db.Del(h.modifyKey(key))
	return DelCost
}

// Has if db has key
func (h *DBHandler) Has(key string, owner ...string) (bool, contract.Cost) {
	return h.h.db.Has(h.modifyKey(key, owner...)), GetCost
}

// MapPut put kfv to db
func (h *DBHandler) MapPut(key, field string, value interface{}, owner ...string) contract.Cost {
	sv := database.MustMarshal(value)
	mk := h.modifyKey(key, owner...)

	h.payRAMForMap(mk, field, sv, owner...)
	h.h.db.MPut(mk, field, sv)
	return PutCost
}

// MapGet get value by kf from db
func (h *DBHandler) MapGet(key, field string, owner ...string) (value interface{}, cost contract.Cost) {
	var mk = h.modifyKey(key, owner...)
	rtn := database.MustUnmarshal(h.h.db.MGet(mk, field))
	return rtn, GetCost
}

// MapKeys list keys
func (h *DBHandler) MapKeys(key string, owner ...string) (fields []string, cost contract.Cost) {
	var mk = h.modifyKey(key, owner...)
	return h.h.db.MKeys(mk), KeysCost
}

// MapDel delete field
func (h *DBHandler) MapDel(key, field string, owner ...string) contract.Cost {
	var mk = h.modifyKey(key, owner...)
	h.releaseRAMForMap(mk, field)

	h.h.db.MDel(h.modifyKey(key), field)
	return DelCost
}

// MapHas if has field
func (h *DBHandler) MapHas(key, field string, owner ...string) (bool, contract.Cost) {
	return h.h.db.MHas(h.modifyKey(key, owner...), field), GetCost
}

// MapLen get length of map
func (h *DBHandler) MapLen(key string, owner ...string) (int, contract.Cost) {
	return len(h.h.db.MKeys(h.modifyKey(key, owner...))), KeysCost
}

// GlobalHas if another contract's db has key
func (h *DBHandler) GlobalHas(con, key string, owner ...string) (bool, contract.Cost) {
	return h.h.db.Has(h.modifyGlobalKey(con, key, owner...)), GetCost
}

// GlobalGet get another contract's data
func (h *DBHandler) GlobalGet(con, key string, owner ...string) (value interface{}, cost contract.Cost) {
	o := h.h.db.Get(h.modifyGlobalKey(con, key, owner...))
	return database.MustUnmarshal(o), GetCost
}

// GlobalMapHas if another contract's map has field
func (h *DBHandler) GlobalMapHas(con, key, field string, owner ...string) (bool, contract.Cost) {
	return h.h.db.MHas(h.modifyGlobalKey(con, key, owner...), field), GetCost
}

// GlobalMapGet get another contract's map data
func (h *DBHandler) GlobalMapGet(con, key, field string, owner ...string) (value interface{}, cost contract.Cost) {
	o := h.h.db.MGet(h.modifyGlobalKey(con, key, owner...), field)
	return database.MustUnmarshal(o), GetCost
}

// GlobalMapKeys get another contract's map keys
func (h *DBHandler) GlobalMapKeys(con, key string, owner ...string) (keys []string, cost contract.Cost) {
	return h.h.db.MKeys(h.modifyGlobalKey(con, key, owner...)), GetCost
}

// GlobalMapLen get another contract's map length
func (h *DBHandler) GlobalMapLen(con, key string, owner ...string) (length int, cost contract.Cost) {
	k, cost := h.GlobalMapKeys(con, key)
	return len(k), cost
}

func (h *DBHandler) modifyKey(key string, owner ...string) string {
	contractName, ok := h.h.ctx.Value("contract_name").(string)
	if !ok {
		return ""
	}
	return h.modifyGlobalKey(contractName, key, owner...)
}

func (h *DBHandler) modifyGlobalKey(contractName, key string, owner ...string) string {
	if len(owner) > 0 {
		return contractName + "@" + owner[0] + database.Separator + key
	}
	return contractName + database.Separator + key
}

func (h *DBHandler) payRAM(k, v string, who ...string) {
	var a string
	if len(who) > 0 {
		a = who[0]
	} else {
		a, _ = h.h.ctx.Value("contract_name").(string)
	}
	oldV := h.h.db.Get(k)
	if oldV == "n" {
		h.h.PayCost(contract.Cost{
			Data: int64(len(k) + len(v)),
		}, a)
	} else {
		h.h.PayCost(contract.Cost{
			Data: int64(len(v) - len(oldV)),
		}, a)
	}
}

func (h *DBHandler) payRAMForMap(k, f, v string, who ...string) {
	var a string
	if len(who) > 0 {
		a = who[0]
	} else {
		a, _ = h.h.ctx.Value("contract_name").(string)
	}
	oldV := h.h.db.MGet(k, f)
	if oldV == "n" {
		h.h.PayCost(contract.Cost{
			Data: int64(len(k) + 2*len(f) + len(v)),
		}, a)
	} else {
		h.h.PayCost(contract.Cost{
			Data: int64(len(v) - len(oldV)),
		}, a)
	}

}

func (h *DBHandler) releaseRAM(k string, who ...string) {
	var a string
	if len(who) > 0 {
		a = who[0]
	} else {
		a, _ = h.h.ctx.Value("contract_name").(string)
	}
	v := h.h.db.Get(k)
	h.h.PayCost(contract.Cost{
		Data: -int64(len(k) + len(v)),
	}, a)
}

func (h *DBHandler) releaseRAMForMap(k, f string, who ...string) {
	var a string
	if len(who) > 0 {
		a = who[0]
	} else {
		a, _ = h.h.ctx.Value("contract_name").(string)
	}
	v := h.h.db.Get(k)
	h.h.PayCost(contract.Cost{
		Data: -int64(len(k) + 2*len(f) + len(v)),
	}, a)
}
