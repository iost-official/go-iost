package host

import (
	"context"

	"strings"

	"strconv"

	"errors"

	"github.com/iost-official/Go-IOS-Protocol/core/contract"
	"github.com/iost-official/Go-IOS-Protocol/core/new_tx"
	"github.com/iost-official/Go-IOS-Protocol/new_vm/database"
)

var (
	ErrBalanceNotEnough = errors.New("balance not enough")
	ErrTransferNegValue = errors.New("trasfer amount less than zero")
	ErrReenter          = errors.New("re-entering")
)

type Caller interface {
	Call(host *Host, contractName, api string, args ...interface{}) (rtn []interface{}, cost *contract.Cost, err error)
}

type Host struct {
	Ctx     context.Context
	DB      *database.Visitor
	monitor Caller
	cost    *contract.Cost
}

func NewHost(ctx context.Context, db *database.Visitor, monitor Caller) *Host {
	return &Host{
		Ctx:     ctx,
		DB:      db,
		cost:    &contract.Cost{},
		monitor: monitor,
	}
}

//func (h *Host) LoadContext(Ctx context.Context) *Host {
//	return &Host{
//		Ctx:     Ctx,
//		DB:      h.DB,
//		monitor: h.monitor,
//		cost:    &contract.Cost{},
//	}
//}
func (h *Host) Cost() *contract.Cost {
	c := h.cost
	h.cost = &contract.Cost{}
	return c
}
func (h *Host) Context() context.Context {
	return h.Ctx
}

//
//func (h *Host) VerifyArgs(api string, args ...interface{}) error {
//	return nil
//}

func (h *Host) Put(key string, value interface{}) {
	c := h.Ctx.Value("contract_name").(string)
	v := database.MustMarshal(value)
	h.DB.Put(c+database.Separator+key, v)
}
func (h *Host) Get(key string) interface{} {
	c := h.Ctx.Value("contract_name").(string)
	rtn := database.MustUnmarshal(h.DB.Get(c + database.Separator + key))

	return rtn
}
func (h *Host) Del(key string) {
	c := h.Ctx.Value("contract_name").(string)
	h.DB.Del(c + database.Separator + key)
}
func (h *Host) MapPut(key, field string, value interface{}) {
	c := h.Ctx.Value("contract_name").(string)
	v := database.MustMarshal(value)
	h.DB.MPut(c+database.Separator+key, field, v)
}
func (h *Host) MapGet(key, field string) (value interface{}) {
	c := h.Ctx.Value("contract_name").(string)
	ans := h.DB.MGet(c+database.Separator+key, field)
	rtn := database.MustUnmarshal(ans)
	return rtn
}
func (h *Host) MapKeys(key string) (fields []string) {
	c := h.Ctx.Value("contract_name").(string)
	return h.DB.MKeys(c + database.Separator + key)
}
func (h *Host) MapDel(key, field string) {
	c := h.Ctx.Value("contract_name").(string)
	h.DB.Del(c + database.Separator + key)
}
func (h *Host) MapLen(key string) int {
	c := h.Ctx.Value("contract_name").(string)
	return len(h.DB.MKeys(c + database.Separator + key))
}
func (h *Host) GlobalGet(contract, key string) interface{} {
	o := h.DB.Get(contract + database.Separator + key)
	return database.MustUnmarshal(o)
}
func (h *Host) GlobalMapGet(contract, key, field string) (value interface{}) {
	o := h.DB.MGet(contract+database.Separator+key, field)
	return database.MustUnmarshal(o)
}
func (h *Host) GlobalMapKeys(contract, key string) []string {
	return h.DB.MKeys(contract + database.Separator + key)
}
func (h *Host) GlobalMapLen(contract, key string) int {
	return len(h.GlobalMapKeys(contract, key))
}
func (h *Host) RequireAuth(pubkey string) bool {
	authList := h.Ctx.Value("auth_list")
	i, ok := authList.(map[string]int)[pubkey]
	return ok && i > 0
}
func (h *Host) Receipt(s string) {
	rec := tx.Receipt{
		Type:    tx.UserDefined,
		Content: s,
	}
	trec := h.Ctx.Value("tx_receipt").(*tx.TxReceipt)
	(*trec).Receipts = append(trec.Receipts, rec)
}
func (h *Host) Call(contract, api string, args ...interface{}) ([]interface{}, *contract.Cost, error) {

	// save stack
	record := contract + "-" + api

	height := h.Ctx.Value("stack_height").(int)

	for i := 0; i < height; i++ {
		key := "stack" + strconv.Itoa(i)
		if h.Ctx.Value(key).(string) == record {
			return nil, nil, ErrReenter
		}
	}

	key := "stack" + strconv.Itoa(height)
	ctx := h.Ctx
	h.Ctx = context.WithValue(h.Ctx, "stack_height", height+1)
	h.Ctx = context.WithValue(h.Ctx, key, record)

	// check args and

	rtn, cost, err := h.monitor.Call(h, contract, api, args...)
	h.Ctx = ctx
	return rtn, cost, err
}
func (h *Host) CallWithReceipt(contract, api string, args ...interface{}) ([]interface{}, *contract.Cost, error) {
	rtn, cost, err := h.Call(contract, api, args...)

	var receipt tx.Receipt
	if err != nil {
		receipt = tx.Receipt{
			Type:    tx.SystemDefined,
			Content: err.Error(),
		}
	}
	receipt = tx.Receipt{
		Type:    tx.SystemDefined,
		Content: "success",
	}

	trec := h.Ctx.Value("tx_receipt").(*tx.TxReceipt)
	(*trec).Receipts = append(trec.Receipts, receipt)

	return rtn, cost, err

}
func (h *Host) Transfer(from, to string, amount int64) error {
	if amount <= 0 {
		return ErrTransferNegValue
	}

	bf := h.DB.Balance(from)
	if bf > amount {
		h.DB.SetBalance(from, -1*amount)
		h.DB.SetBalance(to, amount)
	} else {
		return ErrBalanceNotEnough
	}
	return nil
}
func (h *Host) Withdraw(to string, amount int64) error {
	c := h.Ctx.Value("contract_name").(string)
	return h.Transfer(c, to, amount)
}
func (h *Host) Deposit(from string, amount int64) error {
	c := h.Ctx.Value("contract_name").(string)
	return h.Transfer(from, c, amount)
}
func (h *Host) TopUp(contract, from string, amount int64) error {
	return h.Transfer(from, "g-"+contract, amount)
}
func (h *Host) Countermand(contract, to string, amount int64) error {
	return h.Transfer("g-"+contract, to, amount)
}
func (h *Host) SetCode(ct string) { // 不在这里做编译
	c := contract.Contract{}
	c.Decode(ct)
	h.DB.SetContract(&c)
}
func (h *Host) DestroyCode(contractName string) {
	// todo 释放kv

	h.DB.DelContract(contractName)
}
func (h *Host) BlockInfo() database.SerializedJSON {
	return h.Ctx.Value("block_info").(database.SerializedJSON)
}
func (h *Host) TxInfo() database.SerializedJSON {
	return h.Ctx.Value("tx_info").(database.SerializedJSON)
}
func (h *Host) ABIConfig(key, value string) {
	abi := h.Ctx.Value("abi_config").(*contract.ABI)

	switch key {
	case "payment":
		if value == "contract_pay" {
			(*abi).Payment = 1
		}
	}
}
func (h *Host) PayCost(c *contract.Cost, who string, gasPrice int64) {
	if gasPrice <= 0 {
		panic("gas_price error")
	}
	witness := h.Ctx.Value("witness").(string)
	fee := gasPrice * c.ToGas()
	if strings.HasPrefix(who, "IOST") {
		h.Transfer(who, witness, int64(fee))
	} else {
		err := h.Transfer("g-"+who, witness, int64(fee))
		if err != nil {
			panic(err)
		}
	}
}
