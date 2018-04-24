package vm

import (
	"github.com/iost-official/prototype/state"
	"github.com/iost-official/prototype/core"
)

type Address string

type Privilege int

const (
	Private       Privilege = iota
	SignedPrivate
)

type Signature struct {
	Sig    []byte
	Pubkey []byte
}

type Program struct {
}

type Contract struct {
}

type Method struct {
}

func Transition(sp state.Pool, tx core.Tx) (state.Pool, uint64, error) {
	return nil, 0, nil
}

func getStatus(addr Address, key state.Key) (state.Value, error) {
	return nil, nil
}
