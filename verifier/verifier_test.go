package verifier

import (
	"github.com/golang/mock/gomock"
	"github.com/iost-official/prototype/core/mocks"
	"github.com/iost-official/prototype/core/state"
	"github.com/iost-official/prototype/vm"
	"github.com/iost-official/prototype/vm/lua"
	. "github.com/smartystreets/goconvey/convey"
	"testing"
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
			pool.EXPECT().GetHM(gomock.Any(), gomock.Any()).Return(&v3, nil)
			pool.EXPECT().Copy().AnyTimes().Return(pool)
			main := lua.NewMethod("main", 1)
			code := `function main()
	Put("hello", "world")
	return "success"
end`
			lc := lua.NewContract(vm.ContractInfo{Prefix: "test", GasLimit: 100, Price: 1, Sender: []byte("ahaha")}, code, main)

			cv := NewCacheVerifier(pool)
			_, err := cv.VerifyContract(&lc, true)
			So(err, ShouldBeNil)
			So(string(k), ShouldEqual, "testhello")
			So(string(k2), ShouldEqual, "iost")
			So(string(f2), ShouldEqual, "ahaha")
			vv := v2.(*state.VFloat)
			So(vv.ToFloat64(), ShouldEqual, float64(10000-9))
		})
	})
}
