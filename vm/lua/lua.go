package lua

import (
	"fmt"

	"errors"

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
	contract  *Contract
	callerPC  uint64
	ctx       *vm.Context
}

func (l *VM) Start() error {

	if err := l.L.DoString(l.contract.code); err != nil {
		return err
	}

	return nil
}
func (l *VM) Stop() {
	l.L.Close()
}
func (l *VM) call(pool state.Pool, methodName string, args ...state.Value) ([]state.Value, state.Pool, error) {

	if pool != nil {
		l.cachePool = pool
	} else {
		return nil, nil, errors.New("input pool is nil")
	}

	//fmt.Print("1 ")
	//fmt.Println(l.cachePool.GetHM("iost", "b"))

	method0, err := l.contract.API(methodName)
	if err != nil {
		return nil, pool, err
	}

	method := method0.(*Method)

	var errCrash error = nil

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
				v, err0 := Core2Lua(arg)
				if err0 != nil {
					err = err0
				}
				largs = append(largs, v)
			}
			err = l.L.CallByParam(lua.P{
				Fn:      l.L.GetGlobal(method.name),
				NRet:    method.outputCount,
				Protect: true,
			}, largs...)
		}

		defer func() {
			if err2 := recover(); err2 != nil {
				err3, ok := err2.(error)
				if !ok {
					errCrash = errors.New("recover returns non error value")
				}
				log.Log.E("something wrong in call:", err3.Error())
				errCrash = err3
			}
		}()
	}()

	//fmt.Print("3 ")
	//fmt.Println(l.cachePool.GetHM("iost", "b"))

	if err != nil {
		return nil, pool, err
	}
	if errCrash != nil {
		return nil, pool, errCrash
	}

	rtnValue := make([]state.Value, 0, method.outputCount)
	for i := 0; i < method.outputCount; i++ {
		ret := l.L.Get(-1) // returned value
		l.L.Pop(1)
		v2, err := Lua2Core(ret)
		if err != nil {
			return nil, pool, err
		}
		rtnValue = append(rtnValue, v2)
	}
	return rtnValue, l.cachePool, nil
}
func (l *VM) Call(ctx *vm.Context, pool state.Pool, methodName string, args ...state.Value) ([]state.Value, state.Pool, error) {
	if ctx == nil {
		ctx = vm.BaseContext()
	}
	l.ctx = ctx
	defer func() {
		l.ctx = vm.BaseContext()
	}()
	return l.call(pool, methodName, args...)
}
func (l *VM) Prepare(contract vm.Contract, monitor vm.Monitor) error {
	l.ctx = vm.BaseContext()
	var ok bool
	l.contract, ok = contract.(*Contract)
	if !ok {
		return fmt.Errorf("prepare contract %v : contract type error", contract.Info().Prefix)
	}

	l.L = lua.NewState()
	if contract.Info().GasLimit < 3 {
		return errors.New("gas limit less than 3")
	}
	l.L.PCLimit = uint64(contract.Info().GasLimit)
	l.monitor = monitor

	l.APIs = make([]api, 0)

	var Put = api{
		name: "Put",
		function: func(L *lua.LState) int {
			k := L.ToString(1)
			key := state.Key(l.contract.Info().Prefix + k)
			v := L.Get(2)
			v2, err := Lua2Core(v)
			if err != nil {
				L.Push(lua.LFalse)
				return 1
			}
			host.Put(l.cachePool, key, v2)
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
			host.Log(k, l.contract.info.Prefix)
			return 0
		},
	}
	l.APIs = append(l.APIs, Log)

	var Get = api{
		name: "Get",
		function: func(L *lua.LState) int {
			k := L.ToString(1)
			key := state.Key(l.contract.Info().Prefix + k)
			v, err := host.Get(l.cachePool, key)
			if err != nil {
				L.Push(lua.LNil)
				return 1
			}
			//fmt.Println("get:", v)
			v2, err := Core2Lua(v)
			//fmt.Println(v2.(*lua.LTable).Len())
			if err != nil {
				L.Push(lua.LNil)
				return 1
			}
			L.Push(v2)
			L.PCount += 1000
			return 1
		},
	}
	l.APIs = append(l.APIs, Get)

	var Transfer = api{
		name: "Transfer",
		function: func(L *lua.LState) int {
			src := L.ToString(1) // todo 验证输入
			//fmt.Print("transfer call check")
			if vm.CheckPrivilege(l.ctx, l.contract.info, src) <= 0 {
				L.Push(lua.LString("privilege error"))
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

	var Deposit = api{
		name: "Deposit",
		function: func(L *lua.LState) int {
			src := L.ToString(1) // todo 验证输入
			if vm.CheckPrivilege(l.ctx, l.contract.info, src) <= 0 {
				L.Push(lua.LString("privilege error"))
				return 1
			}
			value := L.ToNumber(2)
			rtn := host.Deposit(l.cachePool, l.contract.Info().Prefix, src, float64(value))
			L.Push(Bool2Lua(rtn))
			return 1
		},
	}
	l.APIs = append(l.APIs, Deposit)

	var ParentHash = api{
		name: "ParentHash",
		function: func(L *lua.LState) int {
			rtn := host.ParentHashLast(l.ctx)
			L.Push(lua.LNumber(float64(rtn)))
			return 1
		},
	}
	l.APIs = append(l.APIs, ParentHash)

	var Withdraw = api{
		name: "Withdraw",
		function: func(L *lua.LState) int {
			des := L.ToString(1)
			value := L.ToNumber(2)
			rtn := host.Withdraw(l.cachePool, l.contract.Info().Prefix, des, float64(value))
			L.Push(Bool2Lua(rtn))
			return 1
		},
	}
	l.APIs = append(l.APIs, Withdraw)

	var Random = api{
		name: "Random",
		function: func(L *lua.LState) int {
			pro := L.ToNumber(1)
			prof := float64(pro)
			rtn := host.RandomByParentHash(l.ctx, prof)
			L.Push(Bool2Lua(rtn))
			return 1
		},
	}
	l.APIs = append(l.APIs, Random)

	var Now = api{
		name: "Now",
		function: func(L *lua.LState) int {
			rtn := lua.LNumber(host.Now(l.ctx))
			L.Push(rtn)
			return 1
		},
	}
	l.APIs = append(l.APIs, Now)

	var Height = api{
		name: "Height",
		function: func(L *lua.LState) int {
			rtn := lua.LNumber(host.BlockHeight(l.ctx))
			L.Push(rtn)
			return 1
		},
	}
	l.APIs = append(l.APIs, Height)

	var Witness = api{
		name: "Witness",
		function: func(L *lua.LState) int {
			rtn := lua.LString(host.Witness(l.ctx))
			L.Push(rtn)
			return 1
		},
	}
	l.APIs = append(l.APIs, Witness)

	var Assert = api{
		name: "Assert",
		function: func(L *lua.LState) int {
			iis := L.Get(1)
			is := iis.Type() == lua.LTBool && iis == lua.LTrue
			if is == false {
				panic("")
			}
			return 0
		},
	}
	l.APIs = append(l.APIs, Assert)

	var Call = api{
		name: "Call",
		function: func(L *lua.LState) int {
			L.PCount += 1000
			contractPrefix := L.ToString(1)
			methodName := L.ToString(2)
			if methodName == "main" {
				L.Push(lua.LFalse)
				return 1
			}
			method, err := l.monitor.GetMethod(contractPrefix, methodName)
			if err != nil {
				fmt.Println("err:", err.Error())
				L.Push(lua.LString(err.Error()))
				return 1
			}

			//fmt.Print("outer call check:")
			p := vm.CheckPrivilege(l.ctx, contract.Info(), string(l.contract.Info().Publisher))
			pri := method.Privilege()
			switch {
			case pri == vm.Private && p > 1:
				fallthrough
			case pri == vm.Protected && p >= 0:
				fallthrough
			case pri == vm.Public:
				args := make([]state.Value, 0)

				for i := 1; i <= method.InputCount(); i++ {
					v2, err := Lua2Core(L.Get(i + 2))
					if err != nil {
						L.Push(lua.LString(err.Error()))
						return 1
					}

					args = append(args, v2)
				}

				ctx := vm.NewContext(l.ctx)
				ctx.Publisher = l.contract.Info().Publisher
				ctx.Signers = l.contract.Info().Signers

				rtn, pool, gas, err := l.monitor.Call(ctx, l.cachePool, contractPrefix, methodName, args...)
				l.callerPC += gas
				if err != nil {
					fmt.Println("err:", err.Error())
					L.Push(lua.LString(err.Error()))
				}
				l.cachePool = pool
				for _, v := range rtn {
					v2, err := Core2Lua(v)
					if err != nil {
						L.Push(lua.LString(err.Error()))
						return 1
					}
					L.Push(v2)
				}
				return len(rtn)
			default:
				L.Push(lua.LString(err.Error()))
				return 1
			}
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
	l.contract = contract.(*Contract)
	if contract.Info().GasLimit < 3 {
		return errors.New("gas limit less than 3")
	}
	l.L.PCLimit = uint64(contract.Info().GasLimit)
	if err := l.L.DoString(l.contract.code); err != nil {
		return err
	}
	return nil
}

func (l *VM) Contract() vm.Contract {
	return l.contract
}
