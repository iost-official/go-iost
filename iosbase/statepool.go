package iosbase

import "fmt"

type StatePool interface {
	Add(utxo State) error
	Find(StateHash []byte) (State, error)
	GetSlice() ([]State, error)
	Del(StateHash []byte) error
}

type StatePoolImpl struct {
	stateMap map[string]State
}

func (sp *StatePoolImpl) Init() error {
	return nil
}

func (sp *StatePoolImpl) Add(state State) error {
	if sp.stateMap == nil {
		sp.stateMap = make(map[string]State)
	}
	if s, err := sp.Find(state.Hash()); err == nil {
		return fmt.Errorf("state_exist")
		_ = s
	} else {
		sp.stateMap[string(state.Hash())] = state
		return nil
	}
	return nil
}

func (sp *StatePoolImpl) Find(stateHash []byte) (State, error) {
	if sp.stateMap == nil {
		sp.stateMap = make(map[string]State)
	}
	state, ok := sp.stateMap[string(stateHash)]
	if ok {
		return state, nil
	} else {
		return State{}, fmt.Errorf("not found")
	}
}

func (sp *StatePoolImpl) Del(stateHash []byte) error {
	if sp.stateMap == nil {
		sp.stateMap = make(map[string]State)
	}
	_, ok := sp.stateMap[string(stateHash)]
	if ok {
		delete(sp.stateMap, string(stateHash))
		return nil
	} else {
		return fmt.Errorf("state_not_exist")
	}
}

func (sp *StatePoolImpl) Transact(t Tx) error {
	return nil
}
