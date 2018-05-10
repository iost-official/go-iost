package lua

import (
	"github.com/iost-official/gopher-lua"
	"github.com/iost-official/prototype/core/state"
)

func Lua2Core(value lua.LValue) state.Value {
	var v state.Value
	switch value.(type) {
	case lua.LNumber:
		vl := value.(*lua.LNumber)
		v = state.MakeVFloat(float64(*vl))
	case lua.LString:
		v = state.MakeVString(value.String())
	case lua.LBool:
		if value == lua.LTrue {
			v = state.VTrue
		} else {
			v = state.VFalse
		}
	}
	return v
}

func Core2Lua(value state.Value) lua.LValue {
	var v lua.LValue
	switch value.(type) {
	case *state.VInt:
		vl := value.(*state.VInt)
		v = lua.LNumber(vl.ToInt())
	case *state.VString:
		v = lua.LString([]rune(value.String())[1:])
	case *state.VBool:
		if value == state.VTrue {
			v = lua.LTrue
		} else {
			v = lua.LFalse
		}
	}
	return v
}

func Bool2Lua(b bool) lua.LValue {
	var rtnl lua.LValue
	if b {
		rtnl = lua.LTrue
	} else {
		rtnl = lua.LFalse
	}
	return rtnl
}
