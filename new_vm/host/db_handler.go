package host

import (
	"github.com/iost-official/Go-IOS-Protocol/core/contract"
	"github.com/iost-official/Go-IOS-Protocol/new_vm/database"
)

type DBHandler struct {
	db  *database.Visitor
	ctx *Context
}

func NewDBHandler(db *database.Visitor, ctx *Context) DBHandler {
	return DBHandler{
		db:  db,
		ctx: ctx,
	}
}

func (h *DBHandler) Put(key string, value interface{}) *contract.Cost {
	cn := h.ctx.Value("contract_name").(string)

	v := database.MustMarshal(value)
	h.db.Put(cn+database.Separator+key, v)
	return contract.NewCost(1, 1, 1)
}

func (h *DBHandler) Get(key string) (value interface{}, cost *contract.Cost) {
	cn := h.ctx.Value("contract_name").(string)

	rtn := database.MustUnmarshal(h.db.Get(cn + database.Separator + key))
	return rtn, contract.NewCost(1, 1, 1)
}

func (h *DBHandler) Del(key string) *contract.Cost {
	cn := h.ctx.Value("contract_name").(string)

	h.db.Del(cn + database.Separator + key)

	return contract.NewCost(1, 1, 1)
}

func (h *DBHandler) MapPut(key, field string, value interface{}) *contract.Cost {
	cn := h.ctx.Value("contract_name").(string)

	v := database.MustMarshal(value)
	h.db.MPut(cn+database.Separator+key, field, v)

	return contract.NewCost(1, 1, 1)
}

func (h *DBHandler) MapGet(key, field string) (value interface{}, cost *contract.Cost) {
	cn := h.ctx.Value("contract_name").(string)

	ans := h.db.MGet(cn+database.Separator+key, field)
	rtn := database.MustUnmarshal(ans)
	return rtn, contract.NewCost(1, 1, 1)
}

func (h *DBHandler) MapKeys(key string) (fields []string, cost *contract.Cost) {
	cn := h.ctx.Value("contract_name").(string)

	return h.db.MKeys(cn + database.Separator + key), contract.NewCost(1, 1, 1)
}

func (h *DBHandler) MapDel(key, field string) *contract.Cost {
	cn := h.ctx.Value("contract_name").(string)

	h.db.Del(cn + database.Separator + key)
	return contract.NewCost(1, 1, 1)
}

func (h *DBHandler) MapLen(key string) (int, *contract.Cost) {
	cn := h.ctx.Value("contract_name").(string)

	return len(h.db.MKeys(cn + database.Separator + key)), contract.NewCost(1, 1, 1)
}

func (h *DBHandler) GlobalGet(con, key string) (value interface{}, cost *contract.Cost) {
	o := h.db.Get(con + database.Separator + key)
	return database.MustUnmarshal(o), contract.NewCost(1, 1, 1)
}

func (h *DBHandler) GlobalMapGet(con, key, field string) (value interface{}, cost *contract.Cost) {
	o := h.db.MGet(con+database.Separator+key, field)
	return database.MustUnmarshal(o), contract.NewCost(1, 1, 1)
}

func (h *DBHandler) GlobalMapKeys(con, key string) (keys []string, cost *contract.Cost) {
	return h.db.MKeys(con + database.Separator + key), contract.NewCost(1, 1, 1)
}

func (h *DBHandler) GlobalMapLen(con, key string) (length int, cost *contract.Cost) {
	k, cost := h.GlobalMapKeys(con, key)
	return len(k), contract.NewCost(1, 1, 1)
}
