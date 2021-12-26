package host

import (
	"fmt"

	"github.com/iost-official/go-iost/v3/core/contract"
	"github.com/iost-official/go-iost/v3/vm/database"
)

// DBHandler is an application layer abstraction of our base basic_handler and map_handler.
// it offers interface which has an any type value and ramPayer semantic
// it also handles the Marshal and Unmarshal work and determine the cost of each operation
type DBHandler struct {
	h *Host
}

// NewDBHandler ...
func NewDBHandler(h *Host) DBHandler {
	return DBHandler{
		h: h,
	}
}

// IsValidKey return error if the key is invalid as state db key or field
func IsValidKey(key string) error {
	if len(key) == 0 || len(key) > 64 {
		return fmt.Errorf("key or field length invalid. expected [1, 64], actual %v", len(key))
	}
	for _, c := range key {
		if c < 32 || c > 126 || c == '-' || c == '@' {
			return fmt.Errorf("key or field has invalid char, %v", c)
		}
	}
	return nil
}

// Put put kv to db
func (h *DBHandler) Put(key string, value any, ramPayer ...string) (contract.Cost, error) {
	err := IsValidKey(key)
	if err != nil {
		return CommonErrorCost(1), err
	}

	mk := h.modifyKey(key)

	oldV := h.h.db.Get(mk)
	var payer string
	if len(ramPayer) > 0 {
		payer = ramPayer[0]
	} else {
		payer = h.parseValuePayer(oldV)
		if payer == "" {
			payer = h.h.ctx.Value("contract_name").(string)
		}
	}
	sv := h.modifyValue(value, payer)

	h.payRAM(mk, sv, oldV, payer)
	h.h.db.Put(mk, sv)

	cost := contract.NewCost(0, 0, int64(len(sv)/10))
	if cost.ToGas() < Costs["PutCost"].ToGas() {
		cost = Costs["PutCost"]
	}
	return cost, nil
}

// Get get value of key from db
func (h *DBHandler) Get(key string) (value any, cost contract.Cost) {
	mk := h.modifyKey(key)
	rtn := h.parseValue(h.h.db.Get(mk))
	return rtn, Costs["GetCost"]
}

// Del delete key
func (h *DBHandler) Del(key string) (contract.Cost, error) {
	err := IsValidKey(key)
	if err != nil {
		return CommonErrorCost(1), err
	}
	mk := h.modifyKey(key)
	h.releaseRAM(mk)
	h.h.db.Del(mk)
	return Costs["DelCost"], nil
}

// Has if db has key
func (h *DBHandler) Has(key string) (bool, contract.Cost) {
	mk := h.modifyKey(key)
	return h.h.db.Has(mk), Costs["GetCost"]
}

// MapPut put kfv to db
func (h *DBHandler) MapPut(key, field string, value any, ramPayer ...string) (contract.Cost, error) {
	err := IsValidKey(key)
	if err != nil {
		return CommonErrorCost(1), err
	}
	err = IsValidKey(field)
	if err != nil {
		return CommonErrorCost(1), err
	}

	mk := h.modifyKey(key)
	oldV := h.h.db.MGet(mk, field)
	var payer string
	if len(ramPayer) > 0 {
		payer = ramPayer[0]
	} else {
		payer = h.parseValuePayer(oldV)
		if payer == "" {
			payer = h.h.ctx.Value("contract_name").(string)
		}
	}
	sv := h.modifyValue(value, payer)

	h.payRAMForMap(mk, field, sv, oldV, payer)
	h.h.db.MPut(mk, field, sv)

	cost := contract.NewCost(0, 0, int64(len(sv)/10))
	if cost.ToGas() < Costs["PutCost"].ToGas() {
		cost = Costs["PutCost"]
	}
	return cost, nil
}

// MapGet get value by kf from db
func (h *DBHandler) MapGet(key, field string) (value any, cost contract.Cost) {
	mk := h.modifyKey(key)
	rtn := h.parseValue(h.h.db.MGet(mk, field))
	return rtn, Costs["GetCost"]
}

// MapKeys list keys
func (h *DBHandler) MapKeys(key string) (fields []string, cost contract.Cost) {
	mk := h.modifyKey(key)
	return h.h.db.MKeys(mk), Costs["KeysCost"]
}

// MapDel delete field
func (h *DBHandler) MapDel(key, field string) (contract.Cost, error) {
	err := IsValidKey(key)
	if err != nil {
		return CommonErrorCost(1), err
	}
	err = IsValidKey(field)
	if err != nil {
		return CommonErrorCost(1), err
	}
	mk := h.modifyKey(key)
	h.releaseRAMForMap(mk, field)
	h.h.db.MDel(mk, field)
	return Costs["DelCost"], nil
}

// MapHas if has field
func (h *DBHandler) MapHas(key, field string) (bool, contract.Cost) {
	mk := h.modifyKey(key)
	return h.h.db.MHas(mk, field), Costs["GetCost"]
}

// MapLen get length of map
func (h *DBHandler) MapLen(key string) (int, contract.Cost) {
	keys, cost := h.MapKeys(key)
	return len(keys), cost
}

// GlobalHas if another contract's db has key
func (h *DBHandler) GlobalHas(con, key string) (bool, contract.Cost) {
	mk := h.modifyGlobalKey(con, key)
	return h.h.db.Has(mk), Costs["GetCost"]
}

// GlobalGet get another contract's data
func (h *DBHandler) GlobalGet(con, key string) (value any, cost contract.Cost) {
	mk := h.modifyGlobalKey(con, key)
	rtn := h.parseValue(h.h.db.Get(mk))
	return rtn, Costs["GetCost"]
}

// GlobalMapHas if another contract's map has field
func (h *DBHandler) GlobalMapHas(con, key, field string) (bool, contract.Cost) {
	mk := h.modifyGlobalKey(con, key)
	return h.h.db.MHas(mk, field), Costs["GetCost"]
}

// GlobalMapGet get another contract's map data
func (h *DBHandler) GlobalMapGet(con, key, field string) (value any, cost contract.Cost) {
	mk := h.modifyGlobalKey(con, key)
	rtn := h.parseValue(h.h.db.MGet(mk, field))
	return rtn, Costs["GetCost"]
}

// GlobalMapKeys get another contract's map keys
func (h *DBHandler) GlobalMapKeys(con, key string) (keys []string, cost contract.Cost) {
	mk := h.modifyGlobalKey(con, key)
	return h.h.db.MKeys(mk), Costs["GetCost"]
}

// GlobalMapLen get another contract's map length
func (h *DBHandler) GlobalMapLen(con, key string) (length int, cost contract.Cost) {
	k, cost := h.GlobalMapKeys(con, key)
	return len(k), cost
}

func (h *DBHandler) modifyKey(key string) string {
	contractName, ok := h.h.ctx.Value("contract_name").(string)
	if !ok {
		return ""
	}
	return h.modifyGlobalKey(contractName, key)
}

func (h *DBHandler) modifyGlobalKey(contractName, key string) string {
	return contractName + database.Separator + key
}

func (h *DBHandler) modifyValue(value any, ramPayer ...string) string {
	payer := ""
	if len(ramPayer) > 0 {
		payer = ramPayer[0]
	}
	return database.MustMarshal(value, payer)
}

func (h *DBHandler) parseValue(value string) any {
	return database.MustUnmarshal(value)
}

func (h *DBHandler) parseValuePayer(value string) string {
	_, extra := database.MustUnmarshalWithExtra(value)
	return extra
}

func (h *DBHandler) payRAM(k, v, oldV string, who string) {
	oLen := int64(len(oldV) + len(k))
	nLen := int64(len(v) + len(k))
	h.payRAMInner(oldV, oLen, nLen, who)
}

func (h *DBHandler) payRAMForMap(k, f, v, oldV string, who string) {
	oLen := int64(len(oldV) + len(k) + 2*len(f))
	nLen := int64(len(v) + len(k) + 2*len(f))
	h.payRAMInner(oldV, oLen, nLen, who)
}

func (h *DBHandler) payRAMInner(oldV string, oLen int64, nLen int64, payer string) {
	var data int64
	dataList := make([]contract.DataItem, 0)
	if oldV == "n" {
		dataList = append(dataList, contract.DataItem{Payer: payer, Val: nLen})
		data = nLen
	} else {
		oldPayer := h.parseValuePayer(oldV)
		if oldPayer == "" {
			oldPayer = h.h.ctx.Value("contract_name").(string)
		}
		if oldPayer == payer {
			dataList = append(dataList, contract.DataItem{Payer: payer, Val: nLen - oLen})
			data = nLen - oLen
		} else {
			dataList = append(dataList, contract.DataItem{Payer: oldPayer, Val: -oLen})
			dataList = append(dataList, contract.DataItem{Payer: payer, Val: nLen})
			data = nLen - oLen
		}
	}
	h.h.AddCacheCost(contract.Cost{Data: data, DataList: dataList})
}

func (h *DBHandler) releaseRAMInner(oldV string, oLen int64) {
	data := int64(0)
	dataList := make([]contract.DataItem, 0)
	if oldV == "n" {
		return
	}
	oldPayer := h.parseValuePayer(oldV)
	if oldPayer != "" {
		dataList = append(dataList, contract.DataItem{Payer: oldPayer, Val: -oLen})
	}
	h.h.AddCacheCost(contract.Cost{Data: data, DataList: dataList})
}

func (h *DBHandler) releaseRAM(k string) {
	v := h.h.db.Get(k)
	oLen := int64(len(k) + len(v))
	h.releaseRAMInner(v, oLen)
}

func (h *DBHandler) releaseRAMForMap(k, f string, who ...string) {
	v := h.h.db.MGet(k, f)
	oLen := int64(len(k) + 2*len(f) + len(v))
	h.releaseRAMInner(v, oLen)
}
