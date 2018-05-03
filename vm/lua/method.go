package lua

import (
	"github.com/iost-official/prototype/core/state"
	"github.com/iost-official/gopher-lua"
)

type Method struct {
	name        string
	inputs      []lua.LValue
	outputCount int
}

func NewLuaMethod(name string, rtnCount int, value ...lua.LValue) Method {
	var m Method
	m.name = name
	m.inputs = make([]lua.LValue, 0)
	m.inputs = append(m.inputs, value...)
	m.outputCount = rtnCount
	return m
}

func (m *Method) Name() string {
	return m.name
}
func (m *Method) Input(value ...state.Value) {
	m.inputs = make([]lua.LValue, 0)
	for _, v := range value {
		m.inputs = append(m.inputs, Core2Lua(v))
	}
}

