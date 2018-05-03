package vm

import (
	"github.com/iost-official/prototype/core/state"
	"github.com/iost-official/prototype/common"
)

type Address string

type Privilege int

const (
	Private   Privilege = iota
	Protected
	public
)

type Code string

type Pubkey []byte

type VM interface {
	Prepare(contract Contract, pool state.Pool) error
	Start() error
	Stop()
	Call(methodName string, args ...state.Value) ([]state.Value, state.Pool, error)
	SetPool(pool state.Pool)
	PC() uint64
}
type Method interface {
	Name() string
	Input(...state.Value)
}

//go:generate gencode go -schema=structs.schema -package=vm

type Contract interface {
	Info() ContractInfo
	SetPrefix(prefix string)
	SetSender(sender []byte)
	AddSigner(signer []byte)
	Api(apiName string) (Method, error)
	common.Serializable
}
