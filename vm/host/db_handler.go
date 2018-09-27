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
func (h *DBHandler) Put(key string, value interface{}) *contract.Cost {
	h.h.db.Put(
		h.modifyKey(key),
		database.MustMarshal(value),
	)
	return PutCost
}

// Get get value of key from db
func (h *DBHandler) Get(key string) (value interface{}, cost *contract.Cost) {

	rtn := database.MustUnmarshal(h.h.db.Get(h.modifyKey(key)))

	return rtn, GetCost
}

// Del delete key
func (h *DBHandler) Del(key string) *contract.Cost {

	h.h.db.Del(h.modifyKey(key))

	return DelCost
}

// Has if db has key
func (h *DBHandler) Has(key string) (bool, *contract.Cost) {
	return h.h.db.Has(h.modifyKey(key)), GetCost
}

// MapPut put kfv to db
func (h *DBHandler) MapPut(key, field string, value interface{}) *contract.Cost {
	v := database.MustMarshal(value)
	h.h.db.MPut(h.modifyKey(key), field, v)
	return PutCost
}

// MapGet get value by kf from db
func (h *DBHandler) MapGet(key, field string) (value interface{}, cost *contract.Cost) {
	rtn := database.MustUnmarshal(h.h.db.MGet(h.modifyKey(key), field))
	return rtn, GetCost
}

// MapKeys list keys
func (h *DBHandler) MapKeys(key string) (fields []string, cost *contract.Cost) {

	return h.h.db.MKeys(h.modifyKey(key)), KeysCost
}

// MapDel delete field
func (h *DBHandler) MapDel(key, field string) *contract.Cost {
	h.h.db.MDel(h.modifyKey(key), field)
	return DelCost
}

// MapHas if has field
func (h *DBHandler) MapHas(key, field string) (bool, *contract.Cost) {
	return h.h.db.MHas(h.modifyKey(key), field), GetCost
}

// MapLen get length of map
func (h *DBHandler) MapLen(key string) (int, *contract.Cost) {
	return len(h.h.db.MKeys(h.modifyKey(key))), KeysCost
}

// GlobalGet get another contract's data
func (h *DBHandler) GlobalGet(con, key string) (value interface{}, cost *contract.Cost) {
	o := h.h.db.Get(con + database.Separator + key)
	return database.MustUnmarshal(o), GetCost
}

// GlobalMapGet get another contract's map data
func (h *DBHandler) GlobalMapGet(con, key, field string) (value interface{}, cost *contract.Cost) {
	o := h.h.db.MGet(con+database.Separator+key, field)
	return database.MustUnmarshal(o), GetCost
}

// GlobalMapKeys get another contract's map keys
func (h *DBHandler) GlobalMapKeys(con, key string) (keys []string, cost *contract.Cost) {
	return h.h.db.MKeys(con + database.Separator + key), GetCost
}

// GlobalMapLen get another contract's map length
func (h *DBHandler) GlobalMapLen(con, key string) (length int, cost *contract.Cost) {
	k, cost := h.GlobalMapKeys(con, key)
	return len(k), cost
}

func (h *DBHandler) modifyKey(key string) string {
	contractName, ok := h.h.ctx.Value("contract_name").(string)
	if !ok {
		return ""
	}

	return contractName + database.Separator + key
}
