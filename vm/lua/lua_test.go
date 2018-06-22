package lua

import (
	"testing"

	"fmt"

	"bytes"

	"github.com/iost-official/gopher-lua"
	"github.com/iost-official/prototype/core/state"
	db2 "github.com/iost-official/prototype/db"
	"github.com/iost-official/prototype/vm"
	. "github.com/smartystreets/goconvey/convey"
)

func TestLuaVM(t *testing.T) {
	Convey("Test of Lua VM", t, func() {
		Convey("Normal", func() {
			db, err := db2.DatabaseFactory("redis")
			if err != nil {
				panic(err.Error())
			}
			sdb := state.NewDatabase(db)
			pool := state.NewPool(sdb)

			main := NewMethod(vm.Public, "main", 0, 1)
			lc := Contract{
				info: vm.ContractInfo{Prefix: "test", GasLimit: 10000},
				code: `function main()
	Put("hello", "world")
	return "success"
end`,
				main: main,
			}

			lvm := VM{}
			lvm.Prepare(nil)
			lvm.Start(&lc)
			ret, _, err := lvm.call(pool, "main")
			lvm.Stop()
			So(err, ShouldBeNil)

			v, err := pool.Get("testhello")
			So(err, ShouldBeNil)
			So(ret[0].EncodeString(), ShouldEqual, "ssuccess")
			So(v.EncodeString(), ShouldEqual, "sworld")

		})

		Convey("Transfer", func() {
			db, err := db2.DatabaseFactory("redis")
			if err != nil {
				panic(err.Error())
			}
			sdb := state.NewDatabase(db)
			pool := state.NewPool(sdb)

			pool.PutHM("iost", "a", state.MakeVFloat(100))
			pool.PutHM("iost", "b", state.MakeVFloat(100))

			main := NewMethod(vm.Public, "main", 0, 1)
			lc := Contract{
				info: vm.ContractInfo{Prefix: "test", GasLimit: 11, Publisher: vm.IOSTAccount("a")},

				code: `function main()
	return Transfer("a", "b", 50)
end`,
				main: main,
			}

			lvm := VM{}
			lvm.Prepare(nil)
			lvm.Start(&lc)
			rtn, _, err := lvm.call(pool, "main")
			lvm.Stop()
			So(err, ShouldBeNil)

			So(rtn[0].EncodeString(), ShouldEqual, "true")

		})

		Convey("Fatal", func() {
			db, err := db2.DatabaseFactory("redis")
			if err != nil {
				panic(err.Error())
			}
			sdb := state.NewDatabase(db)
			pool := state.NewPool(sdb)

			pool.PutHM("iost", "a", state.MakeVFloat(10))
			pool.PutHM("iost", "b", state.MakeVFloat(100))

			main := NewMethod(vm.Public, "main", 0, 0)
			lc := Contract{
				info: vm.ContractInfo{Prefix: "test", GasLimit: 10000, Publisher: vm.IOSTAccount("a")},

				code: `function main()
	Assert(Transfer("a", "b", 50))
end`,
				main: main,
			}

			lvm := VM{}
			lvm.Prepare(nil)
			lvm.Start(&lc)
			_, _, err = lvm.call(pool, "main")
			lvm.Stop()
			So(err, ShouldNotBeNil)

			//So(rtn[0].EncodeString(), ShouldEqual, "true")

		})

		Convey("Fatal with no bool assert", func() {
			db, err := db2.DatabaseFactory("redis")
			if err != nil {
				panic(err.Error())
			}
			sdb := state.NewDatabase(db)
			pool := state.NewPool(sdb)

			pool.PutHM("iost", "a", state.MakeVFloat(10))
			pool.PutHM("iost", "b", state.MakeVFloat(100))

			main := NewMethod(vm.Public, "main", 0, 0)
			lc := Contract{
				info: vm.ContractInfo{Prefix: "test", GasLimit: 10000, Publisher: vm.IOSTAccount("a")},

				code: `function main()
	Assert("hello")
end`,
				main: main,
			}

			lvm := VM{}
			lvm.Prepare(nil)
			lvm.Start(&lc)
			_, _, err = lvm.call(pool, "main")
			lvm.Stop()
			So(err, ShouldNotBeNil)

			//So(rtn[0].EncodeString(), ShouldEqual, "true")

		})

		Convey("Out of gas", func() {
			db, err := db2.DatabaseFactory("redis")
			if err != nil {
				panic(err.Error())
			}
			sdb := state.NewDatabase(db)
			pool := state.NewPool(sdb)

			main := NewMethod(vm.Public, "main", 0, 1)
			lc := Contract{
				info: vm.ContractInfo{Prefix: "test", GasLimit: 3},
				code: `function main()
	Put("hello", "world")
	return "success"
end`,
				main: main,
			}

			lvm := VM{}
			lvm.Prepare(nil)
			lvm.Start(&lc)
			_, _, err = lvm.call(pool, "main")
			lvm.Stop()
			So(err, ShouldNotBeNil)

		})

	})

	Convey("Test of Lua value converter", t, func() {
		Convey("Lua to core", func() { // todo core2lua的测试平台
			Convey("string", func() {
				lstr := lua.LString("hello")
				cstr, _ := Lua2Core(lstr)
				So(cstr.Type(), ShouldEqual, state.String)
				So(cstr.EncodeString(), ShouldEqual, "shello")
				lbool := lua.LTrue
				cbool, _ := Lua2Core(lbool)
				So(cbool.Type(), ShouldEqual, state.Bool)
				So(cbool.EncodeString(), ShouldEqual, "true")
				lnum := lua.LNumber(3.14)
				cnum, _ := Lua2Core(lnum)
				So(cnum.Type(), ShouldEqual, state.Float)
				So(cnum.EncodeString(), ShouldEqual, "f3.140000000000000e+00")
			})
		})
	})

}

func TestTransfer(t *testing.T) {
	Convey("Test lua transfer", t, func() {
		main := NewMethod(vm.Public, "main", 0, 1)
		lc := Contract{
			info: vm.ContractInfo{Prefix: "test", GasLimit: 1000, Publisher: vm.IOSTAccount("a")},
			code: `function main()
	Transfer("a", "b", 50)
	return "success"
end`,
			main: main,
		}
		lvm := VM{}

		db, err := db2.DatabaseFactory("redis")
		if err != nil {
			panic(err.Error())
		}
		sdb := state.NewDatabase(db)
		pool := state.NewPool(sdb)
		pool.PutHM("iost", "a", state.MakeVFloat(5000))
		pool.PutHM("iost", "b", state.MakeVFloat(1000))

		lvm.Prepare(nil)
		lvm.Start(&lc)
		//fmt.Print("0 ")
		//fmt.Println(pool.GetHM("iost", "b"))
		_, pool, err = lvm.call(pool, "main")
		lvm.Stop()

		ab, err := pool.GetHM("iost", "a")
		bb, err := pool.GetHM("iost", "b")
		So(err, ShouldBeNil)
		So(ab.(*state.VFloat).ToFloat64(), ShouldEqual, 4950)
		So(bb.(*state.VFloat).ToFloat64(), ShouldEqual, 1050)

	})
}

func BenchmarkLuaVM_SetupAndCalc(b *testing.B) {
	main := NewMethod(vm.Public, "main", 0, 1)
	lc := Contract{
		info: vm.ContractInfo{Prefix: "test", GasLimit: 11},
		code: `function main()
	return "success"
end`,
		main: main,
	}
	lvm := VM{}

	db, err := db2.DatabaseFactory("redis")
	if err != nil {
		panic(err.Error())
	}
	sdb := state.NewDatabase(db)
	pool := state.NewPool(sdb)

	for i := 0; i < b.N; i++ {
		lvm.Prepare(nil)
		lvm.Start(&lc)
		lvm.call(pool, "main")
		lvm.Stop()
	}
}

func BenchmarkLuaVM_10000LuaGas(b *testing.B) {
	main := NewMethod(vm.Public, "main", 0, 1)
	lc := Contract{
		info: vm.ContractInfo{Prefix: "test", GasLimit: 10000},
		code: `function main()
	while( true )
do
   i = 1
end
	return "success"
end`,
		main: main,
	}
	lvm := VM{}

	db, err := db2.DatabaseFactory("redis")
	if err != nil {
		panic(err.Error())
	}
	sdb := state.NewDatabase(db)
	pool := state.NewPool(sdb)

	for i := 0; i < b.N; i++ {
		lvm.Prepare(nil)
		lvm.Start(&lc)
		lvm.call(pool, "main")
		lvm.Stop()
	}
}

func BenchmarkLuaVM_GetInPatch(b *testing.B) {
	main := NewMethod(vm.Public, "main", 0, 1)
	lc := Contract{
		info: vm.ContractInfo{Prefix: "test", GasLimit: 1000},
		code: `function main()
	Get("hello")
	return "success"
end`,
		main: main,
	}
	lvm := VM{}

	db, err := db2.DatabaseFactory("redis")
	if err != nil {
		panic(err.Error())
	}
	sdb := state.NewDatabase(db)
	pool := state.NewPool(sdb)
	pool.Put("hello", state.MakeVString("world"))

	for i := 0; i < b.N; i++ {
		lvm.Prepare(nil)
		lvm.Start(&lc)
		lvm.call(pool, "main")
		lvm.Stop()
	}
}

func BenchmarkLuaVM_PutToPatch(b *testing.B) {
	main := NewMethod(vm.Public, "main", 0, 1)
	lc := Contract{
		info: vm.ContractInfo{Prefix: "test", GasLimit: 1000},
		code: `function main()
	Put("hello", "world")
	return "success"
end`,
		main: main,
	}
	lvm := VM{}

	db, err := db2.DatabaseFactory("redis")
	if err != nil {
		panic(err.Error())
	}
	sdb := state.NewDatabase(db)
	pool := state.NewPool(sdb)

	for i := 0; i < b.N; i++ {
		lvm.Prepare(nil)
		lvm.Start(&lc)
		lvm.call(pool, "main")
		lvm.Stop()
	}
}

func BenchmarkLuaVM_Transfer(b *testing.B) {
	main := NewMethod(vm.Public, "main", 0, 1)
	lc := Contract{
		info: vm.ContractInfo{Prefix: "test", GasLimit: 1000, Publisher: vm.IOSTAccount("a")},
		code: `function main()
	Transfer("a", "b", 50)
	return "success"
end`,
		main: main,
	}
	lvm := VM{}

	db, err := db2.DatabaseFactory("redis")
	if err != nil {
		panic(err.Error())
	}
	sdb := state.NewDatabase(db)
	pool := state.NewPool(sdb)
	pool.PutHM("iost", "a", state.MakeVFloat(5000))
	pool.PutHM("iost", "b", state.MakeVFloat(50))

	for i := 0; i < b.N; i++ {
		lvm.Prepare(nil)
		lvm.Start(&lc)
		_, _, err = lvm.call(pool, "main")
		lvm.Stop()
	}

}

func TestCompilerNaive(t *testing.T) {
	Convey("test of parse", t, func() {
		parser, _ := NewDocCommentParser(
			`
--- main 合约主入口
-- server1转账server2
-- @gas_limit 11
-- @gas_price 0.0001
-- @param_cnt 0
-- @return_cnt 1
function main()
	print("hello")
	Transfer("abc","mSS7EdV7WvBAiv7TChww7WE3fKDkEYRcVguznbQspj4K", 10)
end--f
`)
		contract, err := parser.Parse()
		So(err, ShouldBeNil)
		So(contract.info.Language, ShouldEqual, "lua")
		So(contract.info.GasLimit, ShouldEqual, 11)
		So(contract.info.Price, ShouldEqual, 0.0001)
		fmt.Println(contract.code)
		So(contract.main, ShouldResemble, Method{"main", 0, 1, vm.Public})
		//So(contract.apis, ShouldResemble, map[string]Method{"foo": Method{"foo", 3, 2, vm.Public}})

	})

}

func TestContract(t *testing.T) {
	main := NewMethod(vm.Public, "main", 0, 1)
	Convey("Test of lua contract", t, func() {
		lc := Contract{
			info: vm.ContractInfo{GasLimit: 1000, Price: 0.1},
			code: `function main()
	Transfer("a", "b", 50)
	return "success"
end`,
			main: main,
		}
		buf := lc.Encode()
		var lc2 Contract
		err := lc2.Decode(buf)
		So(err, ShouldBeNil)
		So(lc2.info.GasLimit, ShouldEqual, lc.info.GasLimit)
		So(lc2.Code(), ShouldEqual, lc.code)
	})
}

func TestContract2(t *testing.T) {
	errlua := `--- main 一元夺宝
-- snatch treasure with 1 coin !
-- @gas_limit 100000
-- @gas_price 0.01
-- @param_cnt 0
-- @return_cnt 1
-- @publisher walleta
function main()
	Put("max_user_number", 20)
	Put("user_number", 0)
	Put("winner", "")
	Put("claimed", "false")
    return "success"
end--f

--- BuyCoin buy coins
-- buy some coins
-- @param_cnt 2
-- @return_cnt 1
function BuyCoin(account, buyNumber)
	if (buyNumber <= 0)
	then
	    return "buy number should be more than zero"
	end

	maxUserNumber = Get("max_user_number")
    number = Get("user_number")
	if (number >= maxUserNumber or number + buyNumber > maxUserNumber)
	then
	    return string.format("max user number exceed, only %d coins left", maxUserNumber - number)
	end

	-- print(string.format("deposit account = %s, number = %d", account, buyNumber))
	Deposit(account, buyNumber)

	win = false
	for i = 0, buyNumber - 1, 1 do
	    win = win or winAfterBuyOne(number)
	    number = number + 1
	end
	Put("user_number", number)

	if (win)
	then
	    Put("winner", account)
	end

    return "success"
end--f

--- winAfterBuyOne win after buy one
-- @param_cnt 1
-- @return_cnt 1
function winAfterBuyOne(number)
	win = Random(1 - 1.0 / (number + 1))
	return win
end--f

--- QueryWinner query winner
-- @param_cnt 0
-- @return_cnt 1
function QueryWinner()
	return Get("winner")
end--f

--- QueryClaimed query claimed
-- @param_cnt 0
-- @return_cnt 1
function QueryClaimed()
	return Get("claimed")
end--f

--- QueryUserNumber query user number 
-- @param_cnt 0
-- @return_cnt 1
function QueryUserNumber()
	return Get("user_number")
end--f

--- QueryMaxUserNumber query max user number 
-- @param_cnt 0
-- @return_cnt 1
function QueryMaxUserNumber()
	return Get("max_user_number")
end--f

--- Claim claim prize
-- @param_cnt 0
-- @return_cnt 1
function Claim()
	claimed = Get("claimed")
	if (claimed == "true")
	then
		return "price has been claimed"
	end
	number = Get("user_number")
	maxUserNumber = Get("max_user_number")
	if (number < maxUserNumber)
	then
		return string.format("game not end yet! user_number = %d, max_user_number = %d", number, maxUserNumber)
	end
	winner = Get("winner")

	Put("claimed", "true")

	Withdraw(winner, number)
	return "success"
end--f
`

	Convey("test of encode", t, func() {
		parser, err := NewDocCommentParser(errlua)
		So(err, ShouldBeNil)
		sc, err := parser.Parse()
		So(err, ShouldBeNil)
		buf := sc.Encode()

		var sc2 Contract
		sc2.Decode(buf)
		fmt.Println()
		//fmt.Println(sc.Encode())
		//fmt.Println(sc2.Encode())
		So(bytes.Equal(sc.Encode(), sc2.Encode()), ShouldBeTrue)
	})

}

func TestContext(t *testing.T) {
	Convey("Test context privilege", t, func() {
		main := NewMethod(vm.Public, "main", 0, 1)
		lc := Contract{
			info: vm.ContractInfo{Prefix: "test", GasLimit: 10000, Publisher: vm.IOSTAccount("b")},
			code: `function main()
	Transfer("a", "b", 50)
	return "success"
end`,
			main: main,
		}
		lvm := VM{}

		db, err := db2.DatabaseFactory("redis")
		if err != nil {
			panic(err.Error())
		}
		sdb := state.NewDatabase(db)
		pool := state.NewPool(sdb)
		pool.PutHM("iost", "a", state.MakeVFloat(5000))
		pool.PutHM("iost", "b", state.MakeVFloat(1000))

		lvm.Prepare(nil)
		lvm.Start(&lc)
		//fmt.Print("0 ")
		//fmt.Println(pool.GetHM("iost", "b"))

		ctx := &vm.Context{
			Publisher: vm.IOSTAccount("a"),
		}

		_, pool, err = lvm.Call(ctx, pool, "main")
		lvm.Stop()

		ab, err := pool.GetHM("iost", "a")
		bb, err := pool.GetHM("iost", "b")
		So(err, ShouldBeNil)
		So(ab.(*state.VFloat).ToFloat64(), ShouldEqual, 4950)
		So(bb.(*state.VFloat).ToFloat64(), ShouldEqual, 1050)

	})
}

func TestVM_Restart(t *testing.T) {
	Convey("test of restart a vm", t, func() {
		db, err := db2.DatabaseFactory("redis")
		if err != nil {
			panic(err.Error())
		}
		sdb := state.NewDatabase(db)
		pool := state.NewPool(sdb)
		pool.PutHM("iost", "a", state.MakeVFloat(5000))
		pool.PutHM("iost", "b", state.MakeVFloat(1000))

		main := NewMethod(vm.Public, "main", 0, 1)
		lc := Contract{
			info: vm.ContractInfo{Prefix: "test", GasLimit: 3, Publisher: vm.IOSTAccount("a")},
			code: `function main()
	Transfer("a", "b", 50)
	return "success"
end`,
			main: main,
		}

		lc2 := Contract{
			info: vm.ContractInfo{Prefix: "test", GasLimit: 10000, Publisher: vm.IOSTAccount("a")},
			code: `function main()
	Transfer("a", "b", 100)
	return "success"
end`,
			main: main,
		}

		var lvm VM

		lvm.Prepare(nil)
		lvm.Start(&lc)
		//fmt.Println(*lvm.L)
		So(lvm.PC(), ShouldEqual, 3)

		lvm.Call(vm.BaseContext(), pool, "main")
		//fmt.Println(pool.GetHM("iost", "b"))
		So(lvm.PC(), ShouldEqual, 4)

		//fmt.Println(*lvm.L)
		So(lvm.L.PCLimit, ShouldEqual, 3)
		lvm.Restart(&lc2)
		//fmt.Println(lvm.Call(vm.BaseContext(), pool, "main"))
		//fmt.Println(pool.GetHM("iost", "b"))

		So(lvm.L.PCLimit, ShouldEqual, 10000)
	})

}

func TestTable(t *testing.T) {
	Convey("test of table", t, func() {
		db, err := db2.DatabaseFactory("redis")
		if err != nil {
			panic(err.Error())
		}
		sdb := state.NewDatabase(db)
		pool := state.NewPool(sdb)

		main := NewMethod(vm.Public, "main", 0, 1)
		lc := Contract{
			info: vm.ContractInfo{GasLimit: 1000, Price: 0.1},
			code: `function main()
	test = {}
	test["a"] = 1
	test["b"] = 2
	Put("test", test)
	return "success"
end`,
			main: main,
		}

		lvm := VM{}
		lvm.Prepare(nil)
		lvm.Start(&lc)
		_, pool, err = lvm.call(pool, "main")
		lvm.Stop()
		So(err, ShouldNotBeNil)

		pool.Flush()

		fmt.Println(pool.GetHM("test", "a"))

	})
}

func TestLog(t *testing.T) {
	Convey("test of table", t, func() {
		db, err := db2.DatabaseFactory("redis")
		if err != nil {
			panic(err.Error())
		}
		sdb := state.NewDatabase(db)
		pool := state.NewPool(sdb)

		main := NewMethod(vm.Public, "main", 0, 1)
		lc := Contract{
			info: vm.ContractInfo{GasLimit: 1000, Price: 0.1},
			code: `function main()
	Log("hello world")
end`,
			main: main,
		}

		lvm := VM{}
		lvm.Prepare(nil)
		lvm.Start(&lc)
		_, pool, err = lvm.call(pool, "main")
		lvm.Stop()
		So(err, ShouldBeNil)

		pool.Flush()

	})
}
