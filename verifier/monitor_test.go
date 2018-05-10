package verifier

import (
	"testing"

	"github.com/bouk/monkey"
	"github.com/golang/mock/gomock"
	"github.com/iost-official/prototype/core/mocks"
	"github.com/iost-official/prototype/core/state"
	"github.com/iost-official/prototype/vm"
	"github.com/iost-official/prototype/vm/lua"
	. "github.com/smartystreets/goconvey/convey"
)

func TestContractCall(t *testing.T) {
	Convey("Test of trans-contract call", t, func() {

		mockCtl := gomock.NewController(t)
		pool := core_mock.NewMockPool(mockCtl)

		pool.EXPECT().Copy().AnyTimes().Return(pool)
		v3 := state.MakeVFloat(float64(10000))
		pool.EXPECT().GetHM(gomock.Any(), gomock.Any()).Return(v3, nil)

		code1 := `function main()
	return Call("con2", "sayHi", "bob")
end`
		code2 := `function sayHi(name)
	return "hi " .. name
end`
		sayHi := lua.NewMethod("sayHi", 1, 1)
		main := lua.NewMethod("main", 0, 1)

		lc1 := lua.NewContract(vm.ContractInfo{Prefix: "con1", GasLimit: 1000, Price: 1, Sender: []byte("ahaha")},
			code1, main)

		lc2 := lua.NewContract(vm.ContractInfo{Prefix: "con2", GasLimit: 1000, Price: 1, Sender: []byte("ahaha")},
			code2, sayHi, sayHi)

		guard := monkey.Patch(FindContract, func(prefix string) vm.Contract { return &lc2 })
		defer guard.Unpatch()

		verifier := Verifier{
			Pool:      pool,
			VMMonitor: NewVMMonitor(),
		}
		verifier.StartVM(&lc1)
		//verifier.StartVM(&lc2)
		rtn, _, gas, err := verifier.Call(pool, "con2", "sayHi", state.MakeVString("bob"))
		So(err, ShouldBeNil)
		So(gas, ShouldEqual, 7)
		So(rtn[0].String(), ShouldEqual, "shi bob")
		rtn, _, gas, err = verifier.Call(pool, "con1", "main")
		So(err, ShouldBeNil)
		So(gas, ShouldEqual, 19)
		So(rtn[0].String(), ShouldEqual, "shi bob")

	})
}
