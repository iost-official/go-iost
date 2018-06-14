package lua

import (
	"testing"

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
			lvm.Prepare(&lc, nil)
			lvm.Start()
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
			lvm.Prepare(&lc, nil)
			lvm.Start()
			rtn, _, err := lvm.call(pool, "main")
			lvm.Stop()
			So(err, ShouldBeNil)

			So(rtn[0].EncodeString(), ShouldEqual, "true")

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
			lvm.Prepare(&lc, nil)
			lvm.Start()
			_, _, err = lvm.call(pool, "main")
			lvm.Stop()
			So(err, ShouldNotBeNil)

		})

	})

	Convey("Test of Lua value converter", t, func() {
		Convey("Lua to core", func() { // todo core2lua的测试平台
			Convey("string", func() {
				lstr := lua.LString("hello")
				cstr := Lua2Core(lstr)
				So(cstr.Type(), ShouldEqual, state.String)
				So(cstr.EncodeString(), ShouldEqual, "shello")
				lbool := lua.LTrue
				cbool := Lua2Core(lbool)
				So(cbool.Type(), ShouldEqual, state.Bool)
				So(cbool.EncodeString(), ShouldEqual, "true")
				lnum := lua.LNumber(3.14)
				cnum := Lua2Core(lnum)
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

		lvm.Prepare(&lc, nil)
		lvm.Start()
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
		lvm.Prepare(&lc, nil)
		lvm.Start()
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
		lvm.Prepare(&lc, nil)
		lvm.Start()
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
		lvm.Prepare(&lc, nil)
		lvm.Start()
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
		lvm.Prepare(&lc, nil)
		lvm.Start()
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
		lvm.Prepare(&lc, nil)
		lvm.Start()
		_, _, err = lvm.call(pool, "main")
		lvm.Stop()
	}

}

func TestCompilerNaive(t *testing.T) {
	Convey("parse结果应该返回一个contract，code部分去掉注释，api部分保存函数的参数信息，info保存gas，price信息", t, func() {
		parser, _ := NewDocCommentParser(
			`--- main 合约主入口
-- 输出hello world
-- @gas_limit 11
-- @gas_price 0.0001
-- @param_cnt 0
-- @return_cnt 1
function main()
 Put("hello", "world")
 return "success"
end
--- foo 乱七八糟的函数
-- 不知道在干啥
-- @gas_limit 12345678910
-- @gas_price 3.14159
-- @param_cnt 3
-- @return_cnt 2
fucntion foo(a,b,c)
	return a,b
end
`)
		contract, _ := parser.Parse()
		So(contract.info.Language, ShouldEqual, "lua")
		So(contract.info.GasLimit, ShouldEqual, 11)
		So(contract.info.Price, ShouldEqual, 0.0001)
		So(contract.code, ShouldEqual, `function main()
 Put("hello", "world")
 return "success"
end
fucntion foo(a,b,c)
	return a,b
end
`)
		So(contract.main, ShouldResemble, Method{"main", 0, 1, vm.Public})
		So(contract.apis, ShouldResemble, map[string]Method{"foo": Method{"foo", 3, 2, vm.Public}})

	})

}

func TestContract(t *testing.T) {
	main := NewMethod(vm.Public, "main", 0, 1)
	Convey("Test of lua Contract", t, func() {
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

func TestCallback(t *testing.T) {

}
