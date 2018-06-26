package tx

import (
	"fmt"
	"github.com/iost-official/prototype/account"
	"github.com/iost-official/prototype/vm"
	. "github.com/smartystreets/goconvey/convey"
	"os"
	"testing"
)

func TestNewServi(t *testing.T) {
	Convey("test servi", t, func() {
		gopath := os.Getenv("GOPATH")
		fmt.Println(gopath)
		serviDb := gopath + "/src/github.com/iost-official/prototype/core/tx/serviDb"
		serviDb2 := gopath + "/src/github.com/iost-official/serviDb"

		delDir := os.RemoveAll(serviDb)
		if delDir != nil {
			fmt.Println(delDir)
		}

		delDir = os.RemoveAll(serviDb2)
		if delDir != nil {
			fmt.Println(delDir)
		}

		s, e := NewServiPool(len(account.GenesisAccount))
		So(e, ShouldBeNil)

		s.Restore()
		bu := s.BestUser()

		So(len(bu), ShouldEqual, 0)

		Convey("test init", func() {

			for k, v := range account.GenesisAccount {
				ser := s.User(vm.IOSTAccount(k))
				ser.SetBalance(v)
			}

			bu = s.BestUser()
			So(len(bu), ShouldEqual, len(account.GenesisAccount))

			for k, v := range account.GenesisAccount {
				ser := s.User(vm.IOSTAccount(k))
				So(ser.b, ShouldEqual, v)
				//fmt.Println("Account id:", vm.IOSTAccount(k), ", value:", ser.b)
			}

			//for _, v := range bu {
			//	fmt.Println("Account id:", v.owner, ", value:", v.b)
			//}

			s.Flush()
		})

	})
}

func TestInitServi(t *testing.T) {
	Convey("test servi", t, func() {
		gopath := os.Getenv("GOPATH")
		fmt.Println(gopath)
		serviDb := gopath + "/src/github.com/iost-official/prototype/core/tx/serviDb"
		serviDb2 := gopath + "/src/github.com/iost-official/serviDb"

		delDir := os.RemoveAll(serviDb)
		if delDir != nil {
			fmt.Println(delDir)
		}

		delDir = os.RemoveAll(serviDb2)
		if delDir != nil {
			fmt.Println(delDir)
		}

		s, e := NewServiPool(len(account.GenesisAccount))
		So(e, ShouldBeNil)
		s.Restore()
		bu := s.BestUser()

		// So(len(bu), ShouldEqual, 0)

		for k, v := range account.GenesisAccount {
			ser := s.User(vm.IOSTAccount(k))
			ser.SetBalance(v)
		}

		s.UpdateBtu()
		bu = s.BestUser()
		So(len(bu), ShouldEqual, len(account.GenesisAccount))
		//for _, v := range bu {
		//
		//	fmt.Println("Account id:", v.owner, ", value:", v.b)
		//
		//}
		testAcc := "nnnnnnnnnnnnnnnnnnnnnnnnnnnnnnnnnnnnnnnnnnnn"

		ser := s.User(vm.IOSTAccount(testAcc))
		ser.SetBalance(1000000000000000)
		s.UpdateBtu()

		bu = s.BestUser()
		for _, v := range bu {
			if v.owner == vm.IOSTAccount(testAcc) {
				So(v.b, ShouldEqual, 1000000000000000)
			}
			//fmt.Println("Account id:", v.owner, ", value:", v.b)
		}
		s.Flush()

		ser = s.User(vm.IOSTAccount(testAcc))
		ser.SetBalance(66666666666666666)
		bu = s.BestUser()
		for _, v := range bu {
			if v.owner == vm.IOSTAccount(testAcc) {
				So(v.b, ShouldEqual, 66666666666666666)
			}
			//fmt.Println("Account id:", v.owner, ", value:", v.b)
		}

		ser = s.User(vm.IOSTAccount(testAcc))
		So(ser.v, ShouldEqual, 0)
		ser.IncrBehavior(1)
		So(ser.v, ShouldEqual, 1)

	})
}

func TestStoreServi(t *testing.T) {
	Convey("test servi", t, func() {
		gopath := os.Getenv("GOPATH")
		fmt.Println(gopath)
		serviDb := gopath + "/src/github.com/iost-official/prototype/core/tx/serviDb"
		serviDb2 := gopath + "/src/github.com/iost-official/serviDb"

		delDir := os.RemoveAll(serviDb)
		if delDir != nil {
			fmt.Println(delDir)
		}

		delDir = os.RemoveAll(serviDb2)
		if delDir != nil {
			fmt.Println(delDir)
		}

		s, e := NewServiPool(len(account.GenesisAccount))
		So(e, ShouldBeNil)
		s.Restore()
		bu := s.BestUser()

		// So(len(bu), ShouldEqual, 0)
		s.ClearBtu()
		for k, v := range account.GenesisAccount {
			ser := s.User(vm.IOSTAccount(k))
			ser.SetBalance(v)
		}

		s.UpdateBtu()
		bu = s.BestUser()
		So(len(bu), ShouldEqual, len(account.GenesisAccount))
		//for _, v := range bu {
		//
		//	fmt.Println("Account id:", v.owner, ", value:", v.b)
		//
		//}
		s.Flush()
		for k, v := range account.GenesisAccount {
			ser := s.User(vm.IOSTAccount(k))
			So(ser.b, ShouldEqual, v)
			//fmt.Println("Account id:", vm.IOSTAccount(k), ", value:", ser.b)
		}

		// 清空数据
		for k, _ := range s.btu {
			s.delBtu(vm.IOSTAccount(k))
		}
		bu = s.BestUser()

		So(len(bu), ShouldEqual, 0)

		s.Restore()
		bu = s.BestUser()

		for k, v := range account.GenesisAccount {
			ser := s.User(vm.IOSTAccount(k))
			So(ser.b, ShouldEqual, v)
			//fmt.Println("Account id:", vm.IOSTAccount(k), ", value:", ser.b)
		}
		So(len(bu), ShouldEqual, len(account.GenesisAccount))

		testAcc := "nnnnnnnnnnnnnnnnnnnnnnnnnnnnnnnnnnnnnnnnnnnn"
		sr := s.User(vm.IOSTAccount(testAcc))
		sr.SetBalance(10000)
		sr.IncrBehavior(10)
		s.Flush()

		So(len(s.hm), ShouldEqual, 1)
		So(sr.b, ShouldEqual, 10000)
		So(sr.v, ShouldEqual, 10)

		sr = s.User(vm.IOSTAccount(testAcc))
		So(len(s.hm), ShouldEqual, 1)
		So(sr.b, ShouldEqual, 10000)
		So(sr.v, ShouldEqual, 10)

		err := s.delHm(vm.IOSTAccount(testAcc))
		So(err, ShouldBeNil)
		So(len(s.hm), ShouldEqual, 0)

		sr = s.User(vm.IOSTAccount(testAcc))
		So(len(s.hm), ShouldEqual, 1)
		So(sr.b, ShouldEqual, 0)
		So(sr.v, ShouldEqual, 0)

	})
}
