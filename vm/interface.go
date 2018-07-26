/*
Package vm  define vm of smart contract. Use verifier/ to verify txs and blocks
*/
package vm

import (
	"github.com/iost-official/prototype/account"
	"github.com/iost-official/prototype/common"
	"github.com/iost-official/prototype/core/state"
)

type Privilege int

const (
	Private Privilege = iota
	Protected
	Public
)

type IOSTAccount string

//go:generate gencode go -schema=structs.schema -package=vm
//go:generate mockgen -destination mocks/mock_contract.go -package vm_mock github.com/iost-official/prototype/vm contract

// Code type, can be compile to contract
type Code string

type VM interface {
	Prepare(monitor Monitor) error
	Start(contract Contract) error
	Restart(contract Contract) error
	Stop()
	Call(ctx *Context, pool state.Pool, methodName string, args ...state.Value) ([]state.Value, state.Pool, error)
	PC() uint64
	Contract() Contract
}

type Method interface {
	Name() string
	InputCount() int
	OutputCount() int
	Privilege() Privilege
}

type Contract interface {
	Info() ContractInfo
	SetPrefix(prefix string)
	SetSender(sender IOSTAccount)
	AddSigner(signer IOSTAccount)
	API(apiName string) (Method, error)
	Code() string
	common.Serializable
}

type Monitor interface {
	StartVM(contract Contract) (VM, error)
	StopVM(contract Contract)
	Stop()
	GetMethod(contractPrefix, methodName string) (Method, *ContractInfo, error)
	Call(ctx *Context, pool state.Pool, contractPrefix, methodName string, args ...state.Value) ([]state.Value, state.Pool, uint64, error)
}

func PubkeyToIOSTAccount(pubkey []byte) IOSTAccount {
	return IOSTAccount(account.GetIdByPubkey(pubkey))
}

func IOSTAccountToPubkey(id IOSTAccount) []byte {
	return account.GetPubkeyByID(string(id))
}

func HashToPrefix(hash []byte) string {
	return common.Base58Encode(hash)
}

func PrefixToHash(prefix string) []byte {
	return common.Base58Decode(prefix)
}

func CheckPrivilege(ctx *Context, info ContractInfo, name string) int {
	for {
		if ctx != nil {
			if IOSTAccount(name) == ctx.Publisher {
				return 2
			}
			for _, signer := range ctx.Signers {
				if IOSTAccount(name) == signer {
					return 1
				}
			}
			ctx = ctx.Base
		} else {
			break
		}
	}

	if IOSTAccount(name) == info.Publisher {
		return 2
	}
	for _, signer := range info.Signers {
		if IOSTAccount(name) == signer {
			return 1
		}
	}
	return 0
}
