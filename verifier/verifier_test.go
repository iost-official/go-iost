package verifier

import (
	"testing"

	"errors"

	"fmt"

	"github.com/golang/mock/gomock"
	"github.com/iost-official/prototype/account"
	"github.com/iost-official/prototype/core/block"
	"github.com/iost-official/prototype/core/mocks"
	"github.com/iost-official/prototype/core/state"
	"github.com/iost-official/prototype/core/tx"
	"github.com/iost-official/prototype/db"
	"github.com/iost-official/prototype/db/mocks"
	"github.com/iost-official/prototype/vm"
	"github.com/iost-official/prototype/vm/lua"
	. "github.com/smartystreets/goconvey/convey"
)

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

			var k2 state.Key
			var f2 state.Key
			var v2 state.Value
			pool.EXPECT().PutHM(gomock.Any(), gomock.Any(), gomock.Any()).Do(func(key, field state.Key, value state.Value) {
				k2 = key
				f2 = field
				v2 = value
			})
			v3 := state.MakeVFloat(float64(10000))
			pool.EXPECT().GetHM(gomock.Any(), gomock.Any()).Return(v3, nil)
			pool.EXPECT().Copy().AnyTimes().Return(pool)
			main := lua.NewMethod("main", 0, 1)
			code := `function main()
	Put("hello", "world")
	return "success"
end`
			lc := lua.NewContract(vm.ContractInfo{Prefix: "test", GasLimit: 100, Price: 1, Publisher: vm.IOSTAccount("ahaha")}, code, main)

			cv := NewCacheVerifier(pool)
			_, err := cv.VerifyContract(&lc, true)
			So(err, ShouldBeNil)
			So(string(k), ShouldEqual, "testhello")
			So(string(k2), ShouldEqual, "iost")
			So(string(f2), ShouldEqual, "ahaha")
			vv := v2.(*state.VFloat)
			So(vv.ToFloat64(), ShouldEqual, float64(10000-6))
		})
	})
}

func TestBlockVerifier(t *testing.T) {
	Convey("Test of BlockVerifier", t, func() {

		a1, _ := account.NewAccount(nil)
		a2, _ := account.NewAccount(nil)

		ctl := gomock.NewController(t)
		mockDB := db_mock.NewMockDatabase(ctl)
		mockDB.EXPECT().GetHM(gomock.Any(), gomock.Any()).AnyTimes().Return(nil, errors.New("not found"))
		mockDB.EXPECT().Get(gomock.Any()).AnyTimes().Return(nil, errors.New("not found"))

		db := state.NewDatabase(mockDB)

		pool := state.NewPool(db)
		pool.PutHM(state.Key("iost"), state.Key("ahaha"), state.MakeVFloat(10000))

		main := lua.NewMethod("main", 0, 1)
		code := `function main()
	Put("hello", "world")
	return "success"
end`
		lc1 := lua.NewContract(vm.ContractInfo{Prefix: "test", GasLimit: 100, Price: 1, Publisher: vm.IOSTAccount("ahaha")}, code, main)

		code1 := `function main()
	return Call("con2", "sayHi", "bob")
end`

		main2 := lua.NewMethod("main", 0, 1)

		lc2 := lua.NewContract(vm.ContractInfo{Prefix: "test2", GasLimit: 1000, Price: 1, Publisher: vm.IOSTAccount("ahaha")},
			code1, main2)

		tx1 := tx.NewTx(1, &lc1)
		tx1, _ = tx.SignTx(tx1, a1)
		tx2 := tx.NewTx(2, &lc2)
		tx2, _ = tx.SignTx(tx2, a2)

		blk := block.Block{
			Content: []tx.Tx{tx1, tx2},
		}

		bv := NewBlockVerifier(nil)
		bv.SetPool(pool)
		pool2, err := bv.VerifyBlock(&blk, false)
		So(err, ShouldBeNil)
		So(pool2, ShouldNotBeNil)
		bal, _ := pool2.GetHM(state.Key("iost"), state.Key("ahaha"))
		So(bal.(*state.VFloat).ToFloat64(), ShouldEqual, 9985)

	})
}

func BenchmarkCacheVerifier_TransferOnly(b *testing.B) {
	main := lua.NewMethod("main", 0, 1)
	code := `function main()
	Transfer("a", "b", 50)
end`
	lc := lua.NewContract(vm.ContractInfo{Prefix: "test", GasLimit: 100, Price: 1, Publisher: vm.IOSTAccount("a")}, code, main)

	dbx, err := db.DatabaseFactor("redis")
	if err != nil {
		panic(err.Error())
	}
	sdb := state.NewDatabase(dbx)
	pool := state.NewPool(sdb)
	pool.PutHM(state.Key("iost"), state.Key("a"), state.MakeVFloat(1000000))
	pool.PutHM(state.Key("iost"), state.Key("b"), state.MakeVFloat(1000000))

	cv := NewCacheVerifier(pool)
	for i := 0; i < b.N; i++ {
		_, err = cv.VerifyContract(&lc, false)
		if err != nil {
			panic(err)
		}
	}
	fmt.Println(pool.GetHM("iost", "a"))
	fmt.Println(pool.GetHM("iost", "b"))

}
