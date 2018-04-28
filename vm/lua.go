package vm

import (
	"github.com/iost-official/prototype/state"
	"github.com/iost-official/gopher-lua"
	"fmt"
)

type LuaAPI struct {
	name     string
	function func(L *lua.LState) int
}

type LuaVM struct {
	APIs []LuaAPI
	L    *lua.LState

	Pool, cachePool state.Pool

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
func (l *LuaVM) Call(methodName string, args ...state.Value) ([]state.Value, state.Pool, error) {

	l.cachePool = l.Pool.Copy()

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

	rtnValue := make([]state.Value, 0, method.outputCount)
	for i := 0; i < method.outputCount; i ++ {
		ret := l.L.Get(-1) // returned value
		l.L.Pop(1)
		rtnValue = append(rtnValue, Lua2Core(ret))
	}

	return rtnValue, l.cachePool, nil
}
func (l *LuaVM) Prepare(contract Contract, pool state.Pool, prefix string) error {
	var ok bool
	l.Contract, ok = contract.(*LuaContract)
	if !ok {
		return fmt.Errorf("type error")
	}

	l.L = lua.NewState()
	l.Pool = pool

	l.APIs = make([]LuaAPI, 0)

	var Put = LuaAPI{
		name: "Put",
		function: func(L *lua.LState) int {
			k := L.ToString(1)
			key := state.Key(prefix + l.Contract.Info().Name + k)
			v := L.Get(2)
			l.cachePool.Put(key, Lua2Core(v))
			L.Push(lua.LTrue)
			return 1
		},
	}
	l.APIs = append(l.APIs, Put)

	var Log = LuaAPI{
		name: "Log",
		function: func(L *lua.LState) int {
			k := L.ToString(1)
			fmt.Println("From Lua :", k)
			return 0
		},
	}
	l.APIs = append(l.APIs, Log)

	return nil
}
func (l *LuaVM) SetPool(pool state.Pool) {
	l.Pool = pool
}
func (l *LuaVM) PC() uint64 {
	return l.L.PCount
}
