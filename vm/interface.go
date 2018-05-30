/*
Package vm  define vm of smart contract. Use verifier/ to verify txs and blocks
*/
package vm

import (
	"github.com/iost-official/prototype/account"
	"github.com/iost-official/prototype/common"
	"github.com/iost-official/prototype/core/state"
)

// Privilege 设定智能合约的接口权限
type Privilege int

const (
	// Private 接口只能被发布者访问
	Private Privilege = iota
	// Protected 接口只能被签名者访问
	Protected
	// Public 接口谁都能访问
	Public
)

// IOSTAccount iost账户，为base58编码的pubkey
type IOSTAccount string

//go:generate gencode go -schema=structs.schema -package=vm
//go:generate mockgen -destination mocks/mock_contract.go -package vm_mock github.com/iost-official/prototype/vm Contract

// Code type, can be compile to contract
// 代码类型的别名，可以编译为contract
type Code string

// VM 虚拟机interface，定义了虚拟机的接口
//
// 调用流程为prepare - start - call - stop
type VM interface {
	Prepare(contract Contract, monitor Monitor) error
	Start() error
	Stop()
	Call(pool state.Pool, methodName string, args ...state.Value) ([]state.Value, state.Pool, error)
	PC() uint64
}

// Method 方法interface，用来作为接口调用
type Method interface {
	Name() string
	InputCount() int
	OutputCount() int
	Privilege() Privilege
}

// Contract 智能合约interface，其中setPrefix，setSender, AddSigner是从tx构建contract的时候使用
type Contract interface {
	Info() ContractInfo
	SetPrefix(prefix string)
	SetSender(sender IOSTAccount)
	AddSigner(signer IOSTAccount)
	API(apiName string) (Method, error)
	Code() string
	common.Serializable
}

// Monitor 管理虚拟机的管理者，实现在verifier模块
type Monitor interface {
	StartVM(contract Contract) VM
	StopVM(contract Contract)
	Stop()
	GetMethod(contractPrefix, methodName string, caller IOSTAccount) (Method, error)
	Call(pool state.Pool, contractPrefix, methodName string, args ...state.Value) ([]state.Value, state.Pool, uint64, error)
}

func PubkeyToIOSTAccount(pubkey []byte) IOSTAccount {
	return IOSTAccount(account.GetIdByPubkey(pubkey))
}

func HashToPrefix(hash []byte) string {
	return common.Base58Encode(hash)
}

func PrefixToHash(prefix string) []byte {
	return common.Base58Decode(prefix)
}

func CheckPrivilege(info ContractInfo, name string) int {
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
