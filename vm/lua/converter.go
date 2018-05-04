package lua

import (
	"github.com/iost-official/gopher-lua"
	"github.com/iost-official/prototype/core/state"
)

func Lua2Core(value lua.LValue) state.Value {

	switch value.(type) {
	case lua.LString:
		v := state.MakeVString(value.String())
		return &v
	}
	return nil
}

func Core2Lua(value state.Value) lua.LValue {
	return nil
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
