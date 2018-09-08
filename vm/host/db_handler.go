package host

import (
	"github.com/iost-official/Go-IOS-Protocol/core/contract"
	"github.com/iost-official/Go-IOS-Protocol/vm/database"
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

// Put ...
func (h *DBHandler) Put(key string, value interface{}) *contract.Cost {
	h.h.db.Put(
		h.modifyKey(key),
		database.MustMarshal(value),
	)
	return PutCost
}

// Get ...
func (h *DBHandler) Get(key string) (value interface{}, cost *contract.Cost) {

	rtn := database.MustUnmarshal(h.h.db.Get(h.modifyKey(key)))

	return rtn, GetCost
}

// Del ...
func (h *DBHandler) Del(key string) *contract.Cost {

	h.h.db.Del(h.modifyKey(key))

	return DelCost
}

// Has ...
func (h *DBHandler) Has(key string) bool {
	return h.h.db.Has(h.modifyKey(key))
}

// MapPut ...
func (h *DBHandler) MapPut(key, field string, value interface{}) *contract.Cost {
	v := database.MustMarshal(value)
	h.h.db.MPut(h.modifyKey(key), field, v)
	return PutCost
}

// MapGet ...
func (h *DBHandler) MapGet(key, field string) (value interface{}, cost *contract.Cost) {
	rtn := database.MustUnmarshal(h.h.db.MGet(h.modifyKey(key), field))
	return rtn, GetCost
}

// MapKeys ...
func (h *DBHandler) MapKeys(key string) (fields []string, cost *contract.Cost) {

	return h.h.db.MKeys(h.modifyKey(key)), KeysCost
}

// MapDel ...
func (h *DBHandler) MapDel(key, field string) *contract.Cost {
	h.h.db.MDel(h.modifyKey(key), field)
	return DelCost
}

// MapHas ...
func (h *DBHandler) MapHas(key, field string) bool {
	return h.h.db.MHas(h.modifyKey(key), field)
}

// MapLen ...
func (h *DBHandler) MapLen(key string) (int, *contract.Cost) {
	return len(h.h.db.MKeys(h.modifyKey(key))), KeysCost
}

// GlobalGet ...
func (h *DBHandler) GlobalGet(con, key string) (value interface{}, cost *contract.Cost) {
	o := h.h.db.Get(con + database.Separator + key)
	return database.MustUnmarshal(o), GetCost
}

// GlobalMapGet ...
func (h *DBHandler) GlobalMapGet(con, key, field string) (value interface{}, cost *contract.Cost) {
	o := h.h.db.MGet(con+database.Separator+key, field)
	return database.MustUnmarshal(o), GetCost
}

// GlobalMapKeys ...
func (h *DBHandler) GlobalMapKeys(con, key string) (keys []string, cost *contract.Cost) {
	return h.h.db.MKeys(con + database.Separator + key), GetCost
}

// GlobalMapLen ...
func (h *DBHandler) GlobalMapLen(con, key string) (length int, cost *contract.Cost) {
	k, cost := h.GlobalMapKeys(con, key)
	return len(k), cost
}

func (h *DBHandler) modifyKey(key string) string {
	return h.h.ctx.Value("contract_name").(string) + database.Separator + key
}
