package iost

import "fmt"

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
