package verifier

import (
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/iost-official/prototype/core/mocks"
	"github.com/iost-official/prototype/core/state"
	"github.com/iost-official/prototype/core/tx"
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
		sayHi := lua.NewMethod(vm.Public, "sayHi", 1, 1)
		main := lua.NewMethod(vm.Public, "main", 0, 1)

		lc1 := lua.NewContract(vm.ContractInfo{Prefix: "con1", GasLimit: 1000, Price: 1, Publisher: vm.IOSTAccount("ahaha")},
			code1, main)

		lc2 := lua.NewContract(vm.ContractInfo{Prefix: "con2", GasLimit: 1000, Price: 1, Publisher: vm.IOSTAccount("ahaha")},
			code2, sayHi, sayHi)
		//
		//guard := monkey.Patch(FindContract, func(prefix string) vm.Contract { return &lc2 })
		//defer guard.Unpatch()

		verifier := Verifier{
			vmMonitor: newVMMonitor(),
		}
		verifier.StartVM(&lc1)
		verifier.StartVM(&lc2)
		rtn, _, gas, err := verifier.Call(pool, "con2", "sayHi", state.MakeVString("bob"))
		So(err, ShouldBeNil)
		So(gas, ShouldEqual, 4)
		So(rtn[0].EncodeString(), ShouldEqual, "shi bob")
		rtn, _, gas, err = verifier.Call(pool, "con1", "main")
		So(err, ShouldBeNil)
		So(gas, ShouldEqual, 9)
		So(rtn[0].EncodeString(), ShouldEqual, "shi bob")

	})

	Convey("Test of find contract and call", t, func() {

		mockCtl := gomock.NewController(t)
		pool := core_mock.NewMockPool(mockCtl)

		pool.EXPECT().Copy().AnyTimes().Return(pool)
		v3 := state.MakeVFloat(float64(10000))
		pool.EXPECT().GetHM(gomock.Any(), gomock.Any()).Return(v3, nil)
		pool.EXPECT().Get(gomock.Any()).Return(v3, nil)

		code1 := `function main()
	return Call("2uXMjwSgRMCpMwmifCKq5rEPdkDmKmgQxfKfnNrxYGDr", "sayHi", "bob")
end`
		code2 := `function sayHi(name)
			return "hi " .. name
		end`
		sayHi := lua.NewMethod(vm.Public, "sayHi", 1, 1)
		main := lua.NewMethod(vm.Public, "main", 0, 1)
		main3 := lua.NewMethod(vm.Public, "main", 0, 1)

		lc1 := lua.NewContract(vm.ContractInfo{Prefix: "con1", GasLimit: 1000, Price: 1, Publisher: vm.IOSTAccount("ahaha")},
			code1, main)

		lc2 := lua.NewContract(vm.ContractInfo{Prefix: "con2", GasLimit: 1000, Price: 1, Publisher: vm.IOSTAccount("ahaha")},
			code2, sayHi, sayHi)

		lc3 := lua.NewContract(vm.ContractInfo{Prefix: "con3", GasLimit: 1000, Price: 1, Publisher: vm.IOSTAccount("ahaha")},
			`function main()
	return Get("a")
end`, main3)

		//
		//guard := monkey.Patch(FindContract, func(prefix string) vm.Contract { return &lc2 })
		//defer guard.Unpatch()

		txx := tx.NewTx(123, &lc2)
		txx.Time = 1000000
		//fmt.Println("a", txx.Contract.Info().Prefix)
		hash := txx.Hash()
		prefix := vm.HashToPrefix(hash)
		//fmt.Println("b", prefix)
		//txx.Contract.SetPrefix("GmPtEhGJEKH96ieakmfkrXbXiYrZj2xh76XLdnkJxXvi")

		tx.TxDbInstance().Add(&txx)

		//tx2, _ := tx.TxDbInstance().Get(hash)
		//fmt.Println(tx2.Contract.Info().Prefix)

		verifier := Verifier{
			vmMonitor: newVMMonitor(),
		}
		verifier.RestartVM(&lc1)
		//verifier.StartVM(&lc2)
		rtn, _, gas, err := verifier.Call(pool, prefix, "sayHi", state.MakeVString("bob"))
		So(err, ShouldBeNil)
		So(gas, ShouldEqual, 4)
		So(rtn[0].EncodeString(), ShouldEqual, "shi bob")
		rtn, _, gas, err = verifier.Call(pool, "con1", "main")
		So(err, ShouldBeNil)
		So(gas, ShouldEqual, 9)
		So(rtn[0].EncodeString(), ShouldEqual, "shi bob")
		verifier.RestartVM(&lc3)
		rtn, _, gas, err = verifier.Call(pool, "con3", "main")
		So(err, ShouldBeNil)
		So(gas, ShouldEqual, 7)

	})
}
