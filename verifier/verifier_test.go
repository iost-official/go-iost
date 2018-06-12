package verifier

import (
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/iost-official/prototype/core/mocks"
	"github.com/iost-official/prototype/core/state"
	"github.com/iost-official/prototype/db"
	"github.com/iost-official/prototype/vm"
	"github.com/iost-official/prototype/vm/lua"
	"github.com/iost-official/prototype/vm/mocks"
	. "github.com/smartystreets/goconvey/convey"
)

func TestGenesisVerify(t *testing.T) {
	Convey("Test of Genesis verify", t, func() {
		Convey("Parse Contract", func() {
			mockCtl := gomock.NewController(t)
			pool := core_mock.NewMockPool(mockCtl)
			var count int
			var k, f, k2 state.Key
			var v, v2 state.Value
			pool.EXPECT().PutHM(gomock.Any(), gomock.Any(), gomock.Any()).Times(2).Do(func(key, field state.Key, value state.Value) error {
				k, f, v = key, field, value
				count++
				return nil
			})
			pool.EXPECT().Put(gomock.Any(), gomock.Any()).Do(func(key state.Key, value state.Value) {
				k2, v2 = key, value
			})
			pool.EXPECT().Copy().Return(pool)
			contract := vm_mock.NewMockContract(mockCtl)
			contract.EXPECT().Code().Return(`
-- @PutHM iost abc f10000
-- @PutHM iost def f1000
-- @Put hello sworld
`)
			_, err := ParseGenesis(contract, pool)
			So(err, ShouldBeNil)
			So(count, ShouldEqual, 2)
			So(k, ShouldEqual, state.Key("iost"))
			So(v2.EncodeString(), ShouldEqual, "sworld")

		})
	})
}

func TestCacheVerifier(t *testing.T) {
	Convey("Test of CacheVerifier", t, func() {
		Convey("Verify contract", func() {
			mockCtl := gomock.NewController(t)
			pool := core_mock.NewMockPool(mockCtl)

			var k state.Key
			var v state.Value

			pool.EXPECT().Put(gomock.Any(), gomock.Any()).AnyTimes().Do(func(key state.Key, value state.Value) error {
				k = key
				v = value
				return nil
			})

			pool.EXPECT().Get(gomock.Any()).AnyTimes().Return(state.MakeVFloat(3.14), nil)

			var k2 state.Key
			var f2 state.Key
			var v2 state.Value
			pool.EXPECT().PutHM(gomock.Any(), gomock.Any(), gomock.Any()).Do(func(key, field state.Key, value state.Value) {
				k2 = key
				f2 = field
				v2 = value
			})
			v3 := state.MakeVFloat(float64(10000))
			pool.EXPECT().GetHM(gomock.Any(), gomock.Any()).AnyTimes().Return(v3, nil)
			pool.EXPECT().Copy().AnyTimes().Return(pool)
			main := lua.NewMethod(vm.Public, "main", 0, 1)
			code := `function main()
	a = Get("pi")
	Put("hello", a)
	return "success"
end`
			lc := lua.NewContract(vm.ContractInfo{Prefix: "test", GasLimit: 10000, Price: 1, Publisher: vm.IOSTAccount("ahaha")}, code, main)

			cv := NewCacheVerifier()
			_, err := cv.VerifyContract(&lc, pool)
			So(err, ShouldBeNil)
			So(string(k), ShouldEqual, "testhello")
			So(v.EncodeString(), ShouldEqual, "f3.140000000000000e+00")
			So(string(k2), ShouldEqual, "iost")
			So(string(f2), ShouldEqual, "ahaha")
			vv := v2.(*state.VFloat)
			So(vv.ToFloat64(), ShouldEqual, float64(10000-2010))
		})
		Convey("Verify free contract", func() {
			mockCtl := gomock.NewController(t)
			pool := core_mock.NewMockPool(mockCtl)

			var k state.Key
			var v state.Value

			pool.EXPECT().Put(gomock.Any(), gomock.Any()).AnyTimes().Do(func(key state.Key, value state.Value) error {
				k = key
				v = value
				return nil
			})

			pool.EXPECT().Get(gomock.Any()).AnyTimes().Return(state.MakeVFloat(3.14), nil)

			var k2 state.Key
			var f2 state.Key
			var v2 state.Value
			pool.EXPECT().PutHM(gomock.Any(), gomock.Any(), gomock.Any()).Do(func(key, field state.Key, value state.Value) {
				k2 = key
				f2 = field
				v2 = value
			})
			//v3 := state.MakeVFloat(float64(10000))
			pool.EXPECT().GetHM(gomock.Any(), gomock.Any()).AnyTimes().Return(state.VNil, nil)
			pool.EXPECT().Copy().AnyTimes().Return(pool)
			main := lua.NewMethod(vm.Public, "main", 0, 1)
			code := `function main()
	a = Get("pi")
	Put("hello", a)
	return "success"
end`
			lc := lua.NewContract(vm.ContractInfo{Prefix: "test", GasLimit: 10000, Price: 0, Publisher: vm.IOSTAccount("ahaha")}, code, main)

			cv := NewCacheVerifier()
			_, err := cv.VerifyContract(&lc, pool)
			So(err, ShouldBeNil)
			So(string(k), ShouldEqual, "testhello")
			So(v.EncodeString(), ShouldEqual, "f3.140000000000000e+00")
			//So(string(k2), ShouldEqual, "iost")
			//So(string(f2), ShouldEqual, "ahaha")
			//vv := v2.(*state.VFloat)
			//So(vv.ToFloat64(), ShouldEqual, float64(0))
		})
	})
}

//func TestBlockVerifier(t *testing.T) {
//	Convey("Test of BlockVerifier", t, func() {
//
//		a1, _ := account.NewAccount(nil)
//		a2, _ := account.NewAccount(nil)
//
//		ctl := gomock.NewController(t)
//		mockDB := db_mock.NewMockDatabase(ctl)
//		mockDB.EXPECT().GetHM(gomock.Any(), gomock.Any()).AnyTimes().Return(nil, nil)
//		mockDB.EXPECT().Get(gomock.Any()).AnyTimes().Return(nil, nil)
//
//		db := state.NewDatabase(mockDB)
//
//		pool := state.NewPool(db)
//		pool.PutHM(state.Key("iost"), state.Key("ahaha"), state.MakeVFloat(10000))
//		pool.Put(state.Key("a"), state.MakeVFloat(3.14))
//
//		main := lua.NewMethod(vm.Public, "main", 0, 1)
//		code := `function main()
//	a = Get("a")
//	Put("hello", a)
//	return a
//end`
//		lc1 := lua.NewContract(vm.ContractInfo{Prefix: "test", GasLimit: 100, Price: 1, Publisher: vm.IOSTAccount("ahaha")}, code, main)
//
//		code2 := `function main()
//	return Call("con2", "sayHi", "bob")
//end`
//
//		main2 := lua.NewMethod(vm.Public, "main", 0, 1)
//
//		lc2 := lua.NewContract(vm.ContractInfo{Prefix: "test2", GasLimit: 1000, Price: 1, Publisher: vm.IOSTAccount("ahaha")},
//			code2, main2)
//
//		tx1 := tx.NewTx(1, &lc1)
//		tx1, _ = tx.SignTx(tx1, a1)
//		tx2 := tx.NewTx(2, &lc2)
//		tx2, _ = tx.SignTx(tx2, a2)
//
//		blk := block.Block{
//			Content: []tx.Tx{tx1, tx2},
//		}
//
//		bv := NewBlockVerifier(nil)
//		bv.SetPool(pool)
//		pool2, err := bv.VerifyBlock(&blk, false)
//		So(err, ShouldBeNil)
//		So(pool2, ShouldNotBeNil)
//		vt, err := pool2.Get("testhello")
//		So(err, ShouldBeNil)
//		So(vt, ShouldNotBeNil)
//		So(vt.EncodeString(), ShouldEqual, "f3.140000000000000e+00")
//		bal, _ := pool2.GetHM(state.Key("iost"), state.Key("ahaha"))
//		So(bal.(*state.VFloat).ToFloat64(), ShouldEqual, 9985)
//
//	})
//}

func TestCacheVerifier_TransferOnly(t *testing.T) {
	Convey("System test of transfer", t, func() {
		main := lua.NewMethod(vm.Public, "main", 0, 1)
		code := `function main()
	Transfer("a", "b", 50)
end`
		lc := lua.NewContract(vm.ContractInfo{Prefix: "test", GasLimit: 10000, Price: 1, Publisher: vm.IOSTAccount("a")}, code, main)

		dbx, err := db.DatabaseFactor("redis")
		if err != nil {
			panic(err.Error())
		}
		sdb := state.NewDatabase(dbx)
		pool := state.NewPool(sdb)
		pool.PutHM(state.Key("iost"), state.Key("a"), state.MakeVFloat(1000000))
		pool.PutHM(state.Key("iost"), state.Key("b"), state.MakeVFloat(1000000))
		//fmt.Println(pool.GetHM("iost", "b"))
		var pool2 state.Pool

		cv := NewCacheVerifier()
		pool2, err = cv.VerifyContract(&lc, pool)
		if err != nil {
			panic(err)
		}
		aa, err := pool2.GetHM("iost", "a")
		ba, err := pool2.GetHM("iost", "b")
		So(err, ShouldBeNil)
		So(aa.(*state.VFloat).ToFloat64(), ShouldEqual, 999944)
		So(ba.(*state.VFloat).ToFloat64(), ShouldEqual, 1000050)
	})

}

func TestCacheVerifier_Multiple(t *testing.T) {
	Convey("System test of transfer", t, func() {
		main := lua.NewMethod(vm.Public, "main", 0, 1)
		code := `function main()
	Transfer("a", "b", 50)
end`
		lc := lua.NewContract(vm.ContractInfo{Prefix: "test", GasLimit: 10000, Price: 1, Publisher: vm.IOSTAccount("a")}, code, main)

		dbx, err := db.DatabaseFactor("redis")
		if err != nil {
			panic(err.Error())
		}
		sdb := state.NewDatabase(dbx)
		pool := state.NewPool(sdb)
		pool.PutHM(state.Key("iost"), state.Key("a"), state.MakeVFloat(1000000))
		pool.PutHM(state.Key("iost"), state.Key("b"), state.MakeVFloat(1000000))
		//fmt.Println(pool.GetHM("iost", "b"))
		var pool2 state.Pool

		cv := NewCacheVerifier()
		pool2, err = cv.VerifyContract(&lc, pool)
		pool3, err := cv.VerifyContract(&lc, pool2)
		if err != nil {
			panic(err)
		}
		_, err = pool2.GetHM("iost", "a")
		ba, err := pool2.GetHM("iost", "b")

		aa2, err := pool3.GetHM("iost", "a")
		So(err, ShouldBeNil)
		So(aa2.(*state.VFloat).ToFloat64(), ShouldEqual, 999888)
		So(ba.(*state.VFloat).ToFloat64(), ShouldEqual, 1000100)
	})

}

func BenchmarkCacheVerifier_TransferOnly(b *testing.B) {
	main := lua.NewMethod(vm.Public, "main", 0, 1)
	code := `function main()
	Transfer("a", "b", 50)
    return "success"
end`
	lc := lua.NewContract(vm.ContractInfo{Prefix: "test", GasLimit: 10000, Price: 1, Publisher: vm.IOSTAccount("a")}, code, main)

	dbx, err := db.DatabaseFactor("redis")
	if err != nil {
		panic(err.Error())
	}
	sdb := state.NewDatabase(dbx)
	pool := state.NewPool(sdb)
	pool.PutHM(state.Key("iost"), state.Key("a"), state.MakeVFloat(1000000))
	pool.PutHM(state.Key("iost"), state.Key("b"), state.MakeVFloat(1000000))

	var pool2 state.Pool

	cv := NewCacheVerifier()
	for i := 0; i < b.N; i++ {
		pool2, err = cv.VerifyContract(&lc, pool)
		if err != nil {
			panic(err)
		}
		//cv.SetPool(pool2)
	}

	_ = pool2
	//fmt.Println()
	//fmt.Print("1.a: ")
	//fmt.Println(pool2.GetHM("iost", "a"))
	//fmt.Print("1.b: ")
	//fmt.Println(pool2.GetHM("iost", "b"))

}

func BenchmarkCacheVerifierWithCache_TransferOnly(b *testing.B) {
	main := lua.NewMethod(vm.Public, "main", 0, 1)
	code := `function main()
	Transfer("a", "b", 50)
    return "success"
end`
	lc := lua.NewContract(vm.ContractInfo{Prefix: "test", GasLimit: 10000, Price: 1, Publisher: vm.IOSTAccount("a")}, code, main)

	dbx, err := db.DatabaseFactor("redis")
	if err != nil {
		panic(err.Error())
	}
	sdb := state.NewDatabase(dbx)
	pool := state.NewPool(sdb)
	pool.PutHM(state.Key("iost"), state.Key("a"), state.MakeVFloat(1000000))
	pool.PutHM(state.Key("iost"), state.Key("b"), state.MakeVFloat(1000000))

	var pool2 state.Pool

	cv := NewCacheVerifier()
	for i := 0; i < b.N; i++ {
		pool2, err = cv.VerifyContract(&lc, pool)
		if err != nil {
			panic(err)
		}
		pool = pool2
	}

	_ = pool2
	//fmt.Println()
	//fmt.Print("1.a: ")
	//fmt.Println(pool2.GetHM("iost", "a"))
	//fmt.Print("1.b: ")
	//fmt.Println(pool2.GetHM("iost", "b"))

}
