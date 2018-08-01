package new_vm

import (
	"context"

	"github.com/iost-official/Go-IOS-Protocol/core/new_tx"
)

type Host struct {
	ctx context.Context
}

func (h *Host) LoadContext(ctx context.Context) *Host {
	return &Host{
		ctx: ctx,
	}
}
func (h *Host) Put(key, value string) {

}
func (h *Host) Get(key string) string {

}
func (h *Host) Del(key string) {

}
func (h *Host) MapPut(key, field, value string) {

}
func (h *Host) MapGet(key, field string) (value string) {

}
func (h *Host) MapKeys(key string) (fields []string) {

}
func (h *Host) MapDel(key, field string) {

}
func (h *Host) MapLen(key string) int {

}
func (h *Host) Typeof(key string) string {

}
func (h *Host) GlobalGet(contractName, key string) string {

}
func (h *Host) GlobalMapGet(contract, key, field string) (value string) {

}
func (h *Host) GlobalMapDel(contract, key, field string) {

}
func (h *Host) GlobalMapLen(contract, key string) int {

}
func (h *Host) RequireAuth(pubkey string) error {

}
func (h *Host) Receipt(s string) {

}
func (h *Host) Call(contract, api string, args ...string) ([]string, error) {

}
func (h *Host) CallWithReceipt(contract, api string, args ...string) (tx.Receipt, []string, error) {

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
