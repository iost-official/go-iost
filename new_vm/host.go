package new_vm

import (
	"context"

	"github.com/iost-official/Go-IOS-Protocol/core/new_tx"
	"github.com/iost-official/Go-IOS-Protocol/new_vm/database"
)

type Host struct {
	ctx     context.Context
	db      *database.Visitor
	monitor *Monitor
}

func (h *Host) LoadContext(ctx context.Context) *Host {
	return &Host{
		ctx:     ctx,
		db:      h.db,
		monitor: h.monitor,
	}
}
func (h *Host) Put(key, value string) {
	h.db.Checkout(h.ctx.Value("commit").(string))
	contract := h.ctx.Value("contract_name").(string)
	h.db.Put(contract+database.Separator+key, value)
}
func (h *Host) Get(key string) string {
	h.db.Checkout(h.ctx.Value("commit").(string))
	contract := h.ctx.Value("contract_name").(string)
	return h.db.Get(contract + database.Separator + key)
}
func (h *Host) Del(key string) {
	h.db.Checkout(h.ctx.Value("commit").(string))
	contract := h.ctx.Value("contract_name").(string)
	h.db.Del(contract + database.Separator + key)
}
func (h *Host) MapPut(key, field, value string) {
	h.db.Checkout(h.ctx.Value("commit").(string))
	contract := h.ctx.Value("contract_name").(string)
	h.db.MPut(contract+database.Separator+key, field, value)
}
func (h *Host) MapGet(key, field string) (value string) {
	h.db.Checkout(h.ctx.Value("commit").(string))
	contract := h.ctx.Value("contract_name").(string)
	return h.db.MGet(contract+database.Separator+key, field)
}
func (h *Host) MapKeys(key string) (fields []string) {
	h.db.Checkout(h.ctx.Value("commit").(string))
	contract := h.ctx.Value("contract_name").(string)
	return h.db.MKeys(contract + database.Separator + key)
}
func (h *Host) MapDel(key, field string) {
	h.db.Checkout(h.ctx.Value("commit").(string))
	contract := h.ctx.Value("contract_name").(string)
	h.db.Del(contract + database.Separator + key)
}
func (h *Host) MapLen(key string) int {
	h.db.Checkout(h.ctx.Value("commit").(string))
	contract := h.ctx.Value("contract_name").(string)
	return len(h.db.MKeys(contract + database.Separator + key))
}
func (h *Host) GlobalGet(contract, key string) string {
	h.db.Checkout(h.ctx.Value("commit").(string))
	return h.db.Get(contract + database.Separator + key)
}
func (h *Host) GlobalMapGet(contract, key, field string) (value string) {
	h.db.Checkout(h.ctx.Value("commit").(string))
	return h.db.MGet(contract+database.Separator+key, field)
}
func (h *Host) GlobalMapKeys(contract, key string) []string {
	h.db.Checkout(h.ctx.Value("commit").(string))
	return h.db.MKeys(contract + database.Separator + key)
}
func (h *Host) GlobalMapLen(contract, key string) int {
	h.db.Checkout(h.ctx.Value("commit").(string))
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
	trec.Receipts = append(trec.Receipts, rec)
}
func (h *Host) Call(contract, api string, args ...string) ([]string, error) {
	rtn, _, err := h.CallWithReceipt(contract, api, args...)
	return rtn, err
}
func (h *Host) CallWithReceipt(contract, api string, args ...string) ([]string, tx.Receipt, error) {
	return h.monitor.Call(h.ctx, contract, api, args...)
}
func (h *Host) Transfer(from, to string, amount int64) error {

}
func (h *Host) Withdraw(to string, amount int64) error {

}
func (h *Host) Deposit(from string, amount int64) error {

}
func (h *Host) TopUp(contract, from string, amount int64) error {

}
func (h *Host) Countermand(contract, to string, amount int64) error {

}
func (h *Host) BlockInfo() string {

}
func (h *Host) TxInfo() string {

}
