package new_vm

import (
	"context"

	"strings"

	"strconv"

	"github.com/iost-official/Go-IOS-Protocol/core/contract"
	"github.com/iost-official/Go-IOS-Protocol/core/new_tx"
	"github.com/iost-official/Go-IOS-Protocol/new_vm/database"
	"github.com/pkg/errors"
)

var (
	ErrBalanceNotEnough = errors.New("balance not enough")
	ErrTransferNegValue = errors.New("trasfer amount less than zero")
	ErrReenter          = errors.New("re-entering")
)

type Host struct {
	ctx context.Context
	db  *database.Visitor
	//monitor *Monitor
	cost *contract.Cost
}

func NewHost(ctx context.Context, db *database.Visitor) *Host {
	return &Host{
		ctx:  ctx,
		db:   db,
		cost: &contract.Cost{},
	}
}

//func (h *Host) LoadContext(ctx context.Context) *Host {
//	return &Host{
//		ctx:     ctx,
//		db:      h.db,
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
	return h.ctx
}

func (h *Host) VerifyArgs(api string, args ...string) error {
	return nil
}

func (h *Host) Put(key string, value interface{}) {
	c := h.ctx.Value("contract_name").(string)
	v := database.MustMarshall(value)
	h.db.Put(c+database.Separator+key, v)
}
func (h *Host) Get(key string) interface{} {
	c := h.ctx.Value("contract_name").(string)
	rtn := database.MustUnmarshall(h.db.Get(c + database.Separator + key))

	return rtn
}
func (h *Host) Del(key string) {
	c := h.ctx.Value("contract_name").(string)
	h.db.Del(c + database.Separator + key)
}
func (h *Host) MapPut(key, field string, value interface{}) {
	c := h.ctx.Value("contract_name").(string)
	v := database.MustMarshall(value)
	h.db.MPut(c+database.Separator+key, field, v)
}
func (h *Host) MapGet(key, field string) (value interface{}) {
	c := h.ctx.Value("contract_name").(string)
	ans := h.db.MGet(c+database.Separator+key, field)
	rtn := database.MustUnmarshall(ans)
	return rtn
}
func (h *Host) MapKeys(key string) (fields []string) {
	c := h.ctx.Value("contract_name").(string)
	return h.db.MKeys(c + database.Separator + key)
}
func (h *Host) MapDel(key, field string) {
	c := h.ctx.Value("contract_name").(string)
	h.db.Del(c + database.Separator + key)
}
func (h *Host) MapLen(key string) int {
	c := h.ctx.Value("contract_name").(string)
	return len(h.db.MKeys(c + database.Separator + key))
}
func (h *Host) GlobalGet(contract, key string) interface{} {
	o := h.db.Get(contract + database.Separator + key)
	return database.MustUnmarshall(o)
}
func (h *Host) GlobalMapGet(contract, key, field string) (value interface{}) {
	o := h.db.MGet(contract+database.Separator+key, field)
	return database.MustUnmarshall(o)
}
func (h *Host) GlobalMapKeys(contract, key string) []string {
	return h.db.MKeys(contract + database.Separator + key)
}
func (h *Host) GlobalMapLen(contract, key string) int {
	return len(h.GlobalMapKeys(contract, key))
}
func (h *Host) RequireAuth(pubkey string) bool {
	authList := h.ctx.Value("auth_list")
	i, ok := authList.(map[string]int)[pubkey]
	return ok && i > 0
}
func (h *Host) Receipt(s string) {
	rec := tx.Receipt{
		Type:    tx.UserDefined,
		Content: s,
	}
	trec := h.ctx.Value("tx_receipt").(*tx.TxReceipt)
	(*trec).Receipts = append(trec.Receipts, rec)
}
func (h *Host) Call(contract, api string, args ...interface{}) ([]string, *contract.Cost, error) {

	// save stack
	record := contract + "-" + api

	height := h.ctx.Value("stack_height").(int)

	for i := 0; i < height; i++ {
		key := "stack" + strconv.Itoa(i)
		if h.ctx.Value(key).(string) == record {
			return nil, nil, ErrReenter
		}
	}

	key := "stack" + strconv.Itoa(height)
	ctx := h.ctx
	h.ctx = context.WithValue(h.ctx, "stack_height", height+1)
	h.ctx = context.WithValue(h.ctx, key, record)

	// check args and

	rtn, cost, err := staticMonitor.Call(h, contract, api, args...)
	h.ctx = ctx
	return rtn, cost, err
}
func (h *Host) CallWithReceipt(contract, api string, args ...interface{}) ([]string, *contract.Cost, error) {
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

	trec := h.ctx.Value("tx_receipt").(*tx.TxReceipt)
	(*trec).Receipts = append(trec.Receipts, receipt)

	return rtn, cost, err

}
func (h *Host) Transfer(from, to string, amount int64) error {
	if amount <= 0 {
		return ErrTransferNegValue
	}

	bf := h.db.Balance(from)
	if bf > amount {
		h.db.SetBalance(from, -1*amount)
		h.db.SetBalance(to, amount)
	} else {
		return ErrBalanceNotEnough
	}
	return nil
}
func (h *Host) Withdraw(to string, amount int64) error {
	c := h.ctx.Value("contract_name").(string)
	return h.Transfer(c, to, amount)
}
func (h *Host) Deposit(from string, amount int64) error {
	c := h.ctx.Value("contract_name").(string)
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
	h.db.SetContract(&c)
}
func (h *Host) BlockInfo() string {
	return h.ctx.Value("block_info").(string)
}
func (h *Host) TxInfo() string {
	return h.ctx.Value("tx_info").(string)
}
func (h *Host) ABIConfig(key, value string) {
	ps := h.ctx.Value("abi_config").(map[string]*string)[key]
	*ps = value
}
func (h *Host) PayCost(c *contract.Cost, who string, gasPrice uint64) {
	witness := h.ctx.Value("witness").(string)
	fee := gasPrice * c.ToGas()
	if strings.HasPrefix(who, "IOST") {
		h.Transfer(who, witness, int64(fee))
	} else {
		h.Transfer("g-"+who, witness, int64(fee))
	}
}
