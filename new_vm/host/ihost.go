package host

import "github.com/iost-official/Go-IOS-Protocol/core/contract"

type IHost interface {
	RequireAuth(pubkey string) bool
	Cost() *contract.Cost
	Receipt(s string)
	Transfer(from, to string, amount int64) error
	TopUp(contract, from string, amount int64) error
	Countermand(contract, to string, amount int64) error
	Call(contract, api string, args ...interface{}) ([]interface{}, *contract.Cost, error)
	CallWithReceipt(contract, api string, args ...interface{}) ([]interface{}, *contract.Cost, error)
	SetCode(ct string)
	DestroyCode(contractName string)
}
