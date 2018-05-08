package lua

import (
	"fmt"
	"github.com/iost-official/gopher-lua"
	"github.com/iost-official/prototype/common"
	"github.com/iost-official/prototype/core/state"
	"github.com/iost-official/prototype/vm"
	"github.com/iost-official/prototype/vm/host"
)

//go:generate gencode go -schema=structs.schema -package=lua

type api struct {
	name     string
	function func(L *lua.LState) int
}

type VM struct {
	APIs []api
	L    *lua.LState

	Pool, cachePool state.Pool

	Contract *Contract
}

func (l *VM) Start() error {
	for _, api := range l.APIs {
		l.L.SetGlobal(api.name, l.L.NewFunction(api.function))
	}

	if err := l.L.DoString(l.Contract.code); err != nil {
		return err
	}

	return nil
}
func (l *VM) Stop() {
	l.L.Close()
}
func (l *VM) Call(methodName string, args ...state.Value) ([]state.Value, state.Pool, error) {

	l.cachePool = l.Pool.Copy()

	method0, err := l.Contract.Api(methodName)
	if err != nil {
		return nil, nil, err
	}

	method := method0.(*Method)

	err = l.L.CallByParam(lua.P{
		Fn:      l.L.GetGlobal(method.name),
		NRet:    method.outputCount,
		Protect: true,
	})

	if err != nil {
		return nil, nil, err
	}

	rtnValue := make([]state.Value, 0, method.outputCount)
	for i := 0; i < method.outputCount; i++ {
		ret := l.L.Get(-1) // returned value
		l.L.Pop(1)
		rtnValue = append(rtnValue, Lua2Core(ret))
	}

	return rtnValue, l.cachePool, nil
}
func (l *VM) Prepare(contract vm.Contract, pool state.Pool) error {
	var ok bool
	l.Contract, ok = contract.(*Contract)
	if !ok {
		return fmt.Errorf("type error")
	}

	l.L = lua.NewState()
	l.L.PCLimit = uint64(contract.Info().GasLimit)
	l.Pool = pool

	l.APIs = make([]api, 0)

	var Put = api{
		name: "Put",
		function: func(L *lua.LState) int {
			k := L.ToString(1)
			key := state.Key(l.Contract.Info().Prefix + k)
			v := L.Get(2)
			host.Put(l.cachePool, key, Lua2Core(v))
			L.Push(lua.LTrue)
			return 1
		},
	}
	l.APIs = append(l.APIs, Put)

	var Log = api{
		name: "Log",
		function: func(L *lua.LState) int {
			k := L.ToString(1)
			host.Log(k, l.Contract.info.Prefix)
			return 0
		},
	}
	l.APIs = append(l.APIs, Log)

	var Get = api{
		name: "Get",
		function: func(L *lua.LState) int {
			k := L.ToString(1)
			v, err := host.Get(l.cachePool, state.Key(k))
			if err != nil {
				L.Push(lua.LNil)
				return 1
			}
			L.Push(Core2Lua(v))
			return 1
		},
	}
	l.APIs = append(l.APIs, Get)

	var Transfer = api{
		name: "Transfer",
		function: func(L *lua.LState) int {
			src := L.ToString(1)
			if CheckPrivilige(l.Contract.info, src) <= 0 {
				L.Push(lua.LFalse)
				return 1
			}
			des := L.ToString(2)
			value := L.ToNumber(3)
			rtn := host.Transfer(l.cachePool, src, des, float64(value))
			L.Push(Bool2Lua(rtn))
			return 1
		},
	}
	l.APIs = append(l.APIs, Transfer)

	return nil
}
func (l *VM) SetPool(pool state.Pool) {
	l.Pool = pool
}
func (l *VM) PC() uint64 {
	return l.L.PCount
}

func CheckPrivilige(info vm.ContractInfo, name string) int {
	if common.Base58Encode(info.Sender) == name {
		return 2
	}
	for _, signer := range info.Signers {
		if common.Base58Encode(signer) == name {
			return 1
		}
	}
	return 0
}
