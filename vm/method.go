package vm

import (
	"github.com/iost-official/prototype/state"
	"github.com/iost-official/gopher-lua"
)

type Method interface {
	Name() string
	Input(...state.Value)
}

type LuaMethod struct {
	name   string
	inputs []lua.LValue
	outputCount int
}

func NewLuaMethod(name string, value ...lua.LValue) LuaMethod {
	var m LuaMethod
	m.name = name
	m.inputs = make([]lua.LValue, 0)

	m.inputs = append(m.inputs, value...)
	return m
}

func (m *LuaMethod) Name() string {
	return m.name
}
func (m *LuaMethod) Input(value ...state.Value) {
}
