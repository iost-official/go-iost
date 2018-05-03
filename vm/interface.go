package vm

import (
	"github.com/iost-official/prototype/core/state"
)

type Address string

type Privilege int

const (
	Private Privilege = iota
	Protected
	public
)

type Code []byte

type Pubkey []byte

func getStatus(addr Address, key state.Key) (state.Value, error) {
	return nil, nil
}

type VM interface {
	Prepare(contract Contract, pool state.Pool) error
	Start() error
	Stop()
	Call(methodName string, args ...state.Value) ([]state.Value, state.Pool, error)
	SetPool(pool state.Pool)
	PC() uint64
}
