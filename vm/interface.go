package vm

import "github.com/iost-official/prototype/state"

type Address string

type Privilege int

const (
	Private Privilege = iota
	SignedPrivate

)

type Signature struct {
	Sig []byte
	Pubkey []byte
}

type Program struct {

}

type Contract struct {

}

type Method struct {

}



func Exec(program Program) (patch state.Patch, gas uint64) {
	return state.Patch{}, 0
}


func getStatus(addr Address, key Key) (Value, error) {
	return nil, nil
}
