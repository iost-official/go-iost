/*
Package vm  define vm of smart contract. Use verifier/ to verify txs and blocks
*/
package vm

import (
	"github.com/iost-official/prototype/common"
	"github.com/iost-official/prototype/core/state"
)

// Privilege 设定智能合约的接口权限
type Privilege int

const (
	Private Privilege = iota
	Protected
	Public
)

//go:generate gencode go -schema=structs.schema -package=vm
//go:generate mockgen -destination mocks/mock_contract.go -package vm_mock github.com/iost-official/prototype/vm Contract

// code type, can be compile to contract
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
}

// Contract 智能合约interface，其中setPrefix，setSender, AddSigner是从tx构建contract的时候使用
type Contract interface {
	Info() ContractInfo
	SetPrefix(prefix string)
	SetSender(sender []byte)
	AddSigner(signer []byte)
	Api(apiName string) (Method, error)
	common.Serializable
}

// Monitor ...
type Monitor interface {
	StartVM(contract Contract) VM
	StopVM(contract Contract)
	Stop()
	GetMethod(contractPrefix, methodName string) Method
	Call(pool state.Pool, contractPrefix, methodName string, args ...state.Value) ([]state.Value, state.Pool, uint64, error)
}
