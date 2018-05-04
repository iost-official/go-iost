package lua

import (
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/iost-official/gopher-lua"
	"github.com/iost-official/prototype/core/mocks"
	"github.com/iost-official/prototype/core/state"
	"github.com/iost-official/prototype/vm"
	. "github.com/smartystreets/goconvey/convey"
	"fmt"
)

func TestLuaVM(t *testing.T) {
	Convey("Test of Lua VM", t, func() {
		mockCtl := gomock.NewController(t)
		pool := core_mock.NewMockPool(mockCtl)

		var k state.Key
		var v state.Value

		pool.EXPECT().Put(gomock.Any(), gomock.Any()).AnyTimes().Do(func(key state.Key, value state.Value) error {
			k = key
			v = value
			return nil
		})
		pool.EXPECT().Copy().AnyTimes().Return(pool)
		main := NewLuaMethod("main", 1)
		lc := Contract{
			info: vm.ContractInfo{Prefix:"test"},
			code: `function main()
	Put("hello", "world")
	return "success"
end`,
			main: main,
		}

		lvm := VM{}
		lvm.Prepare(&lc, pool)
		lvm.Pool = pool
		lvm.Start()
		ret, _, err := lvm.Call("main")
		lvm.Stop()

		fmt.Println(lvm.PC())

		So(err, ShouldBeNil)
		So(ret[0].String(), ShouldEqual, "success")
		So(k, ShouldEqual, "hello")
		So(v.String(), ShouldEqual, "world")

	})

	Convey("Test of Lua value converter", t, func() {
		Convey("Lua to core", func() {
			Convey("string", func() {
				lstr := lua.LString("hello")
				cstr := Lua2Core(lstr)
				So(cstr.Type(), ShouldEqual, state.String)
				So(cstr.String(), ShouldEqual, "hello")
			})
		})
	})
}
