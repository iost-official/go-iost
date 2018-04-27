package vm

import (
	"github.com/iost-official/prototype/state"
	"github.com/iost-official/gopher-lua"
	"reflect"
	"fmt"
	"unsafe"
)

type LuaAPI struct {
	name     string
	function func(L *lua.LState) int
}

type LuaVM struct {
	APIs []LuaAPI
	L    *lua.LState

	Pool     state.Pool
	Contract *LuaContract
}

func (l *LuaVM) Run(methodName string, args ...state.Value) (state.Pool, error) { // TODO 抛弃输出

	method0, err := l.Contract.Api(methodName)
	if err != nil {
		return nil, err
	}

	method := (*LuaMethod)(unsafe.Pointer(&method0))

	for _, api := range l.APIs {
		l.L.SetGlobal(api.name, l.L.NewFunction(api.function))
	}

	if err := l.L.DoString(l.Contract.code); err != nil {
		return nil, err
	}

	err = l.L.CallByParam(method.Entry, method.inputs...)

	if err != nil {
		return nil, err
	}


	return l.Pool, nil
}
func (l *LuaVM) Prepare(contract Contract, pool state.Pool, prefix string) error {
	if reflect.TypeOf(contract) != reflect.TypeOf(LuaContract{}) {
		return fmt.Errorf("contract type error")
	}

	l.Contract = (*LuaContract)(unsafe.Pointer(&contract))

	l.L = lua.NewState()
	l.Pool = pool.Copy()

	l.APIs = make([]LuaAPI, 0)

	var Return = LuaAPI{
		name: "Return",
		function: func(L *lua.LState) int {
			k := L.ToString(1)
			key := state.Key(prefix + l.Contract.Info().Name + k)

			v := L.Get(2)

			switch v.Type() {
			case lua.LTString:
				val := state.MakeVString(v.String())
				l.Pool.Put(key, &val)
			}

			return 1
		},
	}
	l.APIs = append(l.APIs, Return)

	var Finish = LuaAPI{
		name: "Finish",
		function: func(L *lua.LState) int {
			defer L.Close()
			return 0
		},
	}
	l.APIs = append(l.APIs, Finish)
	return nil
}
