package vm

import (
	"github.com/iost-official/prototype/state"
	"github.com/iost-official/gopher-lua"
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

func (l *LuaVM) Start() error {
	for _, api := range l.APIs {
		l.L.SetGlobal(api.name, l.L.NewFunction(api.function))
	}

	if err := l.L.DoString(l.Contract.code); err != nil {
		return err
	}

	return nil
}
func (l *LuaVM) Stop() {
	l.L.Close()
}
func (l *LuaVM) Call(methodName string, args ...state.Value) ([]state.Value, state.Pool, error) { // TODO 输出的转换

	method0, err := l.Contract.Api(methodName)
	if err != nil {
		return nil, nil, err
	}

	method := method0.(*LuaMethod)

	err = l.L.CallByParam(lua.P{
		Fn:      l.L.GetGlobal(method.name),
		NRet:    method.outputCount,
		Protect: true,
	})

	if err != nil {
		return nil, nil, err
	}

	return nil, l.Pool, nil
}
func (l *LuaVM) Prepare(contract *LuaContract, pool state.Pool, prefix string) error {
	l.Contract = contract

	l.L = lua.NewState()
	l.Pool = pool.Copy()

	l.APIs = make([]LuaAPI, 0)

	var Put = LuaAPI{
		name: "Put",
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
	l.APIs = append(l.APIs, Put)

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
