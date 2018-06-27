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

		s, e := NewServiPool(len(account.GenesisAccount), 1000)
		So(e, ShouldBeNil)

		s.Restore()
		bu, _ := s.BestUser()

		So(len(bu), ShouldEqual, 0)

		Convey("test init", func() {

			for k, v := range account.GenesisAccount {
				ser, _ := s.User(vm.IOSTAccount(k))
				ser.SetBalance(v)
			}

			bu, _ = s.BestUser()
			So(len(bu), ShouldEqual, len(account.GenesisAccount))

			for k, v := range account.GenesisAccount {
				ser, _ := s.User(vm.IOSTAccount(k))
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

		s, e := NewServiPool(len(account.GenesisAccount), 1000)
		So(e, ShouldBeNil)
		s.Restore()
		bu, _ := s.BestUser()

		// So(len(bu), ShouldEqual, 0)

		for k, v := range account.GenesisAccount {
			ser, _ := s.User(vm.IOSTAccount(k))
			ser.SetBalance(v)
		}

		s.UpdateBtu()
		bu, _ = s.BestUser()
		So(len(bu), ShouldEqual, len(account.GenesisAccount))
		//for _, v := range bu {
		//
		//	fmt.Println("Account id:", v.owner, ", value:", v.b)
		//
		//}
		testAcc := "nnnnnnnnnnnnnnnnnnnnnnnnnnnnnnnnnnnnnnnnnnnn"

		ser, _ := s.User(vm.IOSTAccount(testAcc))
		ser.SetBalance(1000000000000000)
		s.UpdateBtu()

		bu, _ = s.BestUser()
		for _, v := range bu {
			if v.owner == vm.IOSTAccount(testAcc) {
				So(v.b, ShouldEqual, 1000000000000000)
			}
			//fmt.Println("Account id:", v.owner, ", value:", v.b)
		}
		s.Flush()

		ser, _ = s.User(vm.IOSTAccount(testAcc))
		ser.SetBalance(66666666666666666)
		bu, _ = s.BestUser()
		for _, v := range bu {
			if v.owner == vm.IOSTAccount(testAcc) {
				So(v.b, ShouldEqual, 66666666666666666)
			}
			//fmt.Println("Account id:", v.owner, ", value:", v.b)
		}

		ser, _ = s.User(vm.IOSTAccount(testAcc))
		So(ser.v, ShouldEqual, 0)
		ser.IncrBehavior(1)
		So(ser.v, ShouldEqual, 1)

	})
}

func TestClearHm(t *testing.T) {
	Convey("test clear hm", t, func() {
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

		s3, e := NewServiPool(len(account.GenesisAccount), 1)
		So(e, ShouldBeNil)
		s3.Restore()
		bu, _ := s3.BestUser()

		// So(len(bu), ShouldEqual, 0)

		for k, v := range account.GenesisAccount {
			ser, _ := s3.User(vm.IOSTAccount(k))
			ser.SetBalance(v)
		}

		s3.UpdateBtu()
		bu, _ = s3.BestUser()
		So(len(bu), ShouldEqual, len(account.GenesisAccount))
		//for _, v := range bu {
		//
		//	fmt.Println("Account id:", v.owner, ", value:", v.b)
		//
		//}
		s3.hm = make(map[vm.IOSTAccount]*Servi, 0)
		s3.cacheSize = 1
		So(len(s3.hm), ShouldEqual, 0)
		testAcc := "nnnnnnnnnnnnnnnnnnnnnnnnnnnnnnnnnnnnnnn11111"

		ser, _ := s3.User(vm.IOSTAccount(testAcc))
		ser.SetBalance(10)
		So(len(s3.hm), ShouldEqual, 1)
		s3.Flush()
		So(len(s3.hm), ShouldEqual, 1)

		testAcc1 := "nnnnnnnnnnnnnnnnnnnnnnnnnnnnnnnnnnnnnn111111"

		ser1, _ := s3.User(vm.IOSTAccount(testAcc1))
		ser1.SetBalance(11)

		So(len(s3.hm), ShouldEqual, 2)

		s3.Flush()
		So(len(s3.hm), ShouldEqual, 0)

		ser, _ = s3.User(vm.IOSTAccount(testAcc))
		So(ser.b, ShouldEqual, 10)

		So(len(s3.hm), ShouldEqual, 1)

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

		s, e := NewServiPool(len(account.GenesisAccount), 1000)
		s.cacheSize = 1000
		s.hm = make(map[vm.IOSTAccount]*Servi, 0)

		So(e, ShouldBeNil)
		s.Restore()
		bu, _ := s.BestUser()

		// So(len(bu), ShouldEqual, 0)
		s.ClearBtu()
		for k, v := range account.GenesisAccount {
			ser, _ := s.User(vm.IOSTAccount(k))
			ser.SetBalance(v)
		}

		s.UpdateBtu()
		bu, _ = s.BestUser()
		So(len(bu), ShouldEqual, len(account.GenesisAccount))
		//for _, v := range bu {
		//
		//	fmt.Println("Account id:", v.owner, ", value:", v.b)
		//
		//}
		s.Flush()
		for k, v := range account.GenesisAccount {
			ser, _ := s.User(vm.IOSTAccount(k))
			So(ser.b, ShouldEqual, v)
			//fmt.Println("Account id:", vm.IOSTAccount(k), ", value:", ser.b)
		}

		// 清空数据
		for k, _ := range s.btu {
			s.delBtu(vm.IOSTAccount(k))
		}
		bu, _ = s.BestUser()

		So(len(bu), ShouldEqual, 0)

		s.Restore()
		bu, _ = s.BestUser()

		for k, v := range account.GenesisAccount {
			ser, _ := s.User(vm.IOSTAccount(k))
			So(ser.b, ShouldEqual, v)
			//fmt.Println("Account id:", vm.IOSTAccount(k), ", value:", ser.b)
		}
		So(len(bu), ShouldEqual, len(account.GenesisAccount))

		testAcc := "nnnnnnnnnnnnnnnnnnnnnnnnnnnnnnnnnnnnnnnnnnnn"
		sr, _ := s.User(vm.IOSTAccount(testAcc))
		sr.SetBalance(10000)
		sr.IncrBehavior(10)
		s.Flush()

		So(len(s.hm), ShouldEqual, 1)
		So(sr.b, ShouldEqual, 10000)
		So(sr.v, ShouldEqual, 10)

		sr, _ = s.User(vm.IOSTAccount(testAcc))
		So(len(s.hm), ShouldEqual, 1)
		So(sr.b, ShouldEqual, 10000)
		So(sr.v, ShouldEqual, 10)

		err := s.delHm(vm.IOSTAccount(testAcc))
		So(err, ShouldBeNil)
		So(len(s.hm), ShouldEqual, 0)

		sr, _ = s.User(vm.IOSTAccount(testAcc))
		So(len(s.hm), ShouldEqual, 1)
		So(sr.b, ShouldEqual, 0)
		So(sr.v, ShouldEqual, 0)

	})
}
