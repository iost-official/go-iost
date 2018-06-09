package lua

import (
	"fmt"

	"github.com/iost-official/gopher-lua"
	"github.com/iost-official/prototype/core/state"
	"github.com/iost-official/prototype/log"
	"github.com/iost-official/prototype/vm"
	"github.com/iost-official/prototype/vm/host"
)

//go:generate gencode go -schema=structs.schema -package=lua

type api struct {
	name     string
	function func(L *lua.LState) int
}

// VM lua 虚拟机的实现
type VM struct {
	APIs []api
	L    *lua.LState

	cachePool state.Pool
	monitor   vm.Monitor
	Contract  *Contract
	callerPC  uint64
}

func (l *VM) Start() error {

	if err := l.L.DoString(l.Contract.code); err != nil {
		return err
	}

	return nil
}
func (l *VM) Stop() {
	l.L.Close()
}
func (l *VM) Call(pool state.Pool, methodName string, args ...state.Value) ([]state.Value, state.Pool, error) {

	if pool != nil {
		l.cachePool = pool
	} else {
		panic("input pool is nill")
	}

	//fmt.Print("1 ")
	//fmt.Println(l.cachePool.GetHM("iost", "b"))

	method0, err := l.Contract.API(methodName)
	if err != nil {
		return nil, nil, err
	}

	method := method0.(*Method)

	func() {
		if len(args) == 0 {
			err = l.L.CallByParam(lua.P{
				Fn:      l.L.GetGlobal(method.name),
				NRet:    method.outputCount,
				Protect: true,
			})
		} else {
			largs := make([]lua.LValue, 0)
			for _, arg := range args {
				largs = append(largs, Core2Lua(arg))
			}
			err = l.L.CallByParam(lua.P{
				Fn:      l.L.GetGlobal(method.name),
				NRet:    method.outputCount,
				Protect: true,
			}, largs...)
		}

		defer func() {
			if err2 := recover(); err2 != nil {
				log.Log.E("something wrong in call:", err.Error())
			}
		}()
	}()

	//fmt.Print("3 ")
	//fmt.Println(l.cachePool.GetHM("iost", "b"))

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
func (l *VM) Prepare(contract vm.Contract, monitor vm.Monitor) error {
	var ok bool
	l.Contract, ok = contract.(*Contract)
	if !ok {
		return fmt.Errorf("prepare contract %v : contract type error", contract.Info().Prefix)
	}

	l.L = lua.NewState()
	l.L.PCLimit = uint64(contract.Info().GasLimit)
	l.monitor = monitor

	l.APIs = make([]api, 0)

	var Put = api{
		name: "Put",
		function: func(L *lua.LState) int {
			k := L.ToString(1)
			key := state.Key(l.Contract.Info().Prefix + k)
			v := L.Get(2)
			host.Put(l.cachePool, key, Lua2Core(v))
			L.Push(lua.LTrue)
			L.PCount += 1000
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
			L.PCount += 1000
			return 1
		},
	}
	l.APIs = append(l.APIs, Get)

	var Transfer = api{
		name: "Transfer",
		function: func(L *lua.LState) int {
			src := L.ToString(1) // todo 验证输入
			if vm.CheckPrivilege(l.Contract.info, src) <= 0 {
				L.Push(lua.LFalse)
				return 1
			}
			des := L.ToString(2)
			value := L.ToNumber(3)
			//fmt.Print("0 ")
			//fmt.Println(l.cachePool.GetHM("iost", state.Key(des)))
			rtn := host.Transfer(l.cachePool, src, des, float64(value))
			//fmt.Print("4 ")
			//fmt.Println(l.cachePool.GetHM("iost", state.Key(des)))
			L.Push(Bool2Lua(rtn))
			return 1
		},
	}
	l.APIs = append(l.APIs, Transfer)

	var Call = api{
		name: "Call",
		function: func(L *lua.LState) int {
			L.PCount += 1000
			blockName := L.ToString(1)
			methodName := L.ToString(2)
			method, err := l.monitor.GetMethod(blockName, methodName, l.Contract.Info().Publisher)

			if err != nil {
				L.Push(lua.LString(err.Error()))
				return 1
			}

			args := make([]state.Value, 0)

			for i := 1; i <= method.InputCount(); i++ {
				args = append(args, Lua2Core(L.Get(i+2)))
			}

			rtn, pool, gas, err := l.monitor.Call(l.cachePool, blockName, methodName, args...)
			l.callerPC += gas
			if err != nil {
				L.Push(lua.LString(err.Error()))
			}
			l.cachePool = pool
			for _, v := range rtn {
				L.Push(Core2Lua(v))
			}
			return len(rtn)
		},
	}
	l.APIs = append(l.APIs, Call)

	for _, api := range l.APIs {
		l.L.SetGlobal(api.name, l.L.NewFunction(api.function))
	}
	return nil
}
func (l *VM) PC() uint64 {
	rtn := l.L.PCount + l.callerPC
	l.L.PCount = 0
	return rtn
}
func (l *VM) Restart(contract vm.Contract) error {
	l.Contract = contract.(*Contract)
	if err := l.L.DoString(l.Contract.code); err != nil {
		return err
	}
	return nil
}
