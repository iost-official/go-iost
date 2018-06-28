package lua

import (
	"fmt"

	"reflect"

	"strconv"

	"github.com/iost-official/gopher-lua"
	"github.com/iost-official/prototype/core/state"
)

var (
	// convert errors
	ErrNotBaseType = fmt.Errorf("not base type")
)

func Lua2Core(value lua.LValue) (rtn state.Value, err error) {
	switch value.(type) {
	case *lua.LTable:
		rtn, err = handleLTable(value.(*lua.LTable))
		if err == ErrNotBaseType {
			err = fmt.Errorf("in %v, %v", value.String(), err.Error())
		}
		return
	default:
		rtn, err = handleLBase(value)
		if err == ErrNotBaseType {
			err = fmt.Errorf("not support convertion: %v", reflect.TypeOf(value).String())
		}
		return
	}
}

func handleLBase(value lua.LValue) (rtn state.Value, err error) {
	var v0 state.Value
	switch value.(type) {
	case *lua.LNilType:
		v0 = state.VDelete
	case lua.LNumber:
		vl := value.(lua.LNumber)
		v0 = state.MakeVFloat(float64(vl))
	case lua.LString:
		v0 = state.MakeVString(value.String())
	case lua.LBool:
		if value == lua.LTrue {
			v0 = state.VTrue
		} else {
			v0 = state.VFalse
		}
	default:
		rtn, err = state.VNil, ErrNotBaseType
		return
	}

	return v0, nil
}

func handleLTable(table *lua.LTable) (rtn *state.VMap, err error) {
	rtn = state.MakeVMap(nil)
	table.ForEach(func(key lua.LValue, value lua.LValue) {
		var k0 state.Key = " "
		switch key.Type() {
		case lua.LTNumber:
			k0 = state.Key(strconv.FormatFloat(float64(key.(lua.LNumber)), 'f', 0, 64))
		case lua.LTString:
			k0 = state.Key(key.(lua.LString))
		default:
			rtn, err = nil, fmt.Errorf("unsatisfied key type")
		}

		var v0 state.Value

		v0, err = handleLBase(value)

		if err == nil {
			rtn.Set(k0, v0)
		}
	})
	return
}

func Core2Lua(value state.Value) (lua.LValue, error) {
	var v lua.LValue
	switch value.(type) {
	case *state.VNilType:
		return lua.LNil, nil
	case *state.VFloat:
		vl := value.(*state.VFloat)
		v = lua.LNumber(vl.ToFloat64())
		return v, nil
	case *state.VString:
		v = lua.LString([]rune(value.EncodeString())[1:])
		return v, nil
	case *state.VBool:
		if value == state.VTrue {
			v = lua.LTrue
		} else {
			v = lua.LFalse
		}
		return v, nil
	case *state.VMap:
		vt := lua.LTable{}
		for k, val := range value.(*state.VMap).Map() {
			lv, err := Core2Lua(val)
			if err != nil {
				return nil, err
			}
			//fmt.Println(k, lv.String())
			i, err := strconv.Atoi(string(k))
			if err != nil {
				vt.RawSetString(string(k), lv)
			} else {
				vt.RawSetInt(i, lv)
			}
		}
		return &vt, nil
	case *state.VDeleteType:
		return lua.LNil, nil
	}
	panic(fmt.Errorf("not support convertion: %v", reflect.TypeOf(value).String()))
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
