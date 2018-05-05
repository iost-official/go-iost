/*
Vm package define vm of smart contract. Use verifier/ to verify txs and blocks
*/
package vm

import (
	"github.com/iost-official/prototype/common"
	"github.com/iost-official/prototype/core/state"
)

type Privilege int

const (
	Private Privilege = iota
	Protected
	Public
)

// code type, can be compile to contract
// 代码类型的别名，可以编译为contract
type Code string

// 虚拟机interface，定义了虚拟机的接口
//
// 调用流程为prepare - start - call - stop
type VM interface {
	Prepare(contract Contract, pool state.Pool) error
	Start() error
	Stop()
	Call(methodName string, args ...state.Value) ([]state.Value, state.Pool, error)
	SetPool(pool state.Pool)
	PC() uint64
}

// 方法interface，用来作为接口调用
type Method interface {
	Name() string
	Input(...state.Value)
}

//go:generate gencode go -schema=structs.schema -package=vm

// 智能合约interface，其中setPrefix，setSender, AddSigner是从tx构建contract的时候使用
type Contract interface {
	Info() ContractInfo
	SetPrefix(prefix string)
	SetSender(sender []byte)
	AddSigner(signer []byte)
	Api(apiName string) (Method, error)
	common.Serializable
}
