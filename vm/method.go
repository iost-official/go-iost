package vm

import (
	"github.com/iost-official/prototype/state"
	"github.com/iost-official/gopher-lua"
	"fmt"
)

type Method interface {
	Name() string
	Input(...state.Value)
}

type LuaMethod struct {
	name      string
	code      string
	inputType []lua.LValueType

	inputs []lua.LValue
	Entry  lua.P
}

func (m *LuaMethod) Name() string {
	return m.name
}
func (m *LuaMethod) Input(value ...state.Value) error {
	m.inputs = make([]lua.LValue, 0)
	for i, val := range value {
		if m.inputType[i] != val.Type() {
			return fmt.Errorf("type error")
		}
		m.inputs = append(m.inputs, val)
	}
	return nil
}


//type Method struct {
//	Name   string
//	Code   Code
//	Owner  Pubkey
//	prefix state.Key // 通过prefix + name 可以获得唯一的Method，类似于state
//}
