package iost

import (
	"bytes"
	"math/big"
	"fmt"
)

type State interface {
	GetValue() *big.Int
	GetScript() string
	GetHash() []byte
	Bytes() []byte
}

type StatePool struct {
	stateMap map[string]State
}

func (sp *StatePool) Add(state State) error {
	if sp.stateMap == nil {
		sp.stateMap = make(map[string]State)
	}
	if sp.Get(state.GetHash()) != nil {
		return fmt.Errorf("state_exist")
	} else {
		sp.stateMap[string(state.GetHash())] = state
		return nil
	}
}

func (sp *StatePool) Get(key []byte) State {
	if sp.stateMap == nil {
		sp.stateMap = make(map[string]State)
	}
	state, ok := sp.stateMap[string(key)]
	if ok {
		return state
	} else {
		return nil
	}
}

func (sp *StatePool) Del(key []byte) error {
	if sp.stateMap == nil {
		sp.stateMap = make(map[string]State)
	}
	_, ok := sp.stateMap[string(key)]
	if ok {
		delete(sp.stateMap, string(key))
		return nil
	} else {
		return fmt.Errorf("state_not_exist")
	}
}


// 比特币UTXO的实现
type UTXO struct {
	value  uint64
	script string
	hash   []byte
}

func (u *UTXO) Bytes() []byte {
	var b Binary
	b.putULong(u.value)
	scriptBytes := bytes.NewBufferString(u.script).Bytes()
	b.putUInt(uint32(len(scriptBytes)))
	b.putBytes(scriptBytes)
	return b.Bytes()
}

func (u *UTXO) GetValue() *big.Int {
	return big.NewInt(int64(u.value))
}

func (u *UTXO) GetScript() string {
	return u.script
}

func (u *UTXO) GetHash() []byte {
	if u.hash == nil || len(u.hash) > 32 {
		u.hash = Sha256(u.Bytes())
	}
	return u.hash
}

func NewUtxo(value uint64, script string) UTXO {
	return UTXO{
		value:  value,
		script: script,
		hash:   nil,
	}
}

// end of UTXO
