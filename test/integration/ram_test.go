package integration

import (
	"github.com/iost-official/go-iost/vm/database"
	"github.com/iost-official/go-iost/vm/native"
	"testing"

	"github.com/iost-official/go-iost/core/tx"
	"github.com/iost-official/go-iost/ilog"
	. "github.com/iost-official/go-iost/verifier"
	. "github.com/smartystreets/goconvey/convey"
)


func TestRAM(t *testing.T) {
	s := NewSimulator()
	defer s.Clear()
	ilog.Stop()

	createAccountsWithResource(s)
	contractName := "ram.iost"
	err := setNonNativeContract(s, contractName, "ram.js", ContractPath)
	if err != nil {
		t.Fatal(err)
	}

	acc := prepareAuth(t, s)
	createToken(t, s, acc)
	s.SetGas(acc.ID, 10000000)

	s.Head.Number = 0
	admin := acc1
	r, err := s.Call(contractName, "initAdmin", array2json([]interface{}{admin.ID}), admin.ID, admin.KeyPair)
	if err != nil || r.Status.Code != tx.StatusCode(tx.Success) {
		panic("call failed " + err.Error() + " " + r.String())
	}

	var initialTotal int64 = 128 * 1024 * 1024 * 1024
	var increaseInterval int64 = 10 * 60                   // increase every 10 mins
	var increaseAmount int64 = 10 * (64*1024*1024*1024) / (365 * 24 * 60) // 64GB per year
	r, err = s.Call(contractName, "issue", array2json([]interface{}{initialTotal, increaseInterval, increaseAmount, 0}), admin.ID, admin.KeyPair)
	if err != nil || r.Status.Code != tx.StatusCode(tx.Success) {
		panic("call failed " + err.Error() + " " + r.String())
	}
	dbKey := "token.iost" + database.Separator + native.TokenInfoMapPrefix + "ram"
	if database.MustUnmarshal(s.Visitor.MGet(dbKey, "fullName")) != "IOST system ram" {
		panic("incorrect token full name")
	}
	initRAM := s.Visitor.TokenBalance("ram", acc.ID)
	s.Head.Number = 1
	other := acc2.ID

	Convey("test of ram", t, func() {
		Convey("user has no ram if he did not buy", func() {
			So(s.Visitor.TokenBalance("ram", acc.ID), ShouldEqual, initRAM)
		})
		Convey("test buy", func() {
			var buyAmount int64 = 1024
			Convey("normal buy", func() {
				//balanceBefore := s.Visitor.TokenBalance("iost", acc.ID)
				ramAvailableBefore := s.Visitor.TokenBalance("ram", contractName)
				s.Visitor.SetTokenBalance("iost", contractName, 0)
				r, err := s.Call(contractName, "buy", array2json([]interface{}{acc.ID, acc.ID, buyAmount}), acc.ID, acc.KeyPair)
				So(err, ShouldEqual, nil)
				So(r.Status.Message, ShouldEqual, "")
				balanceAfter := s.Visitor.TokenBalance("iost", acc.ID)
				ramAvailableAfter := s.Visitor.TokenBalance("ram", contractName)
				//var priceEstimated int64 = 30 * 1e8 // TODO when the final function is set, update here
				So(balanceAfter, ShouldEqual, 96970 * 1e6)
				So(s.Visitor.TokenBalance("ram", acc.ID), ShouldEqual, initRAM+buyAmount)
				So(ramAvailableAfter, ShouldEqual, ramAvailableBefore-buyAmount)
			})
			Convey("when buying triggers increasing total ram", func() {
				head := s.Head
				head.Time = head.Time + 144 * increaseInterval*1000*1000*1000
				s.SetBlockHead(head)
				ramAvailableBefore := s.Visitor.TokenBalance("ram", contractName)
				r, err := s.Call(contractName, "buy", array2json([]interface{}{acc.ID, acc.ID, buyAmount}), acc.ID, acc.KeyPair)
				So(err, ShouldEqual, nil)
				So(r.Status.Message, ShouldEqual, "")
				ramAvailableAfter := s.Visitor.TokenBalance("ram", contractName)
				So(ramAvailableAfter, ShouldEqual, ramAvailableBefore+ 144 * increaseAmount-buyAmount)
			})
			Convey("user can buy for others", func() {
				//balanceBefore := s.Visitor.TokenBalance("iost", acc.ID)
				otherRAMBefore := s.Visitor.TokenBalance("ram", other)
				myRAMBefore := s.Visitor.TokenBalance("ram", acc.ID)
				r, err := s.Call(contractName, "buy", array2json([]interface{}{acc.ID, other, buyAmount}), acc.ID, acc.KeyPair)
				So(err, ShouldEqual, nil)
				So(r.Status.Message, ShouldEqual, "")
				balanceAfter := s.Visitor.TokenBalance("iost", acc.ID)
				otherRAMAfter := s.Visitor.TokenBalance("ram", other)
				myRAMAfter := s.Visitor.TokenBalance("ram", acc.ID)
				//var priceEstimated int64 = 30 * 1e8 // TODO when the final function is set, update here
				So(balanceAfter, ShouldEqual, 90918 * 1e6)
				So(myRAMAfter, ShouldEqual, myRAMBefore)
				So(otherRAMAfter, ShouldEqual, otherRAMBefore+buyAmount)
			})
		})
		Convey("test sell", func() {
			Convey("user cannot sell more than he owns", func() {
				r, err := s.Call(contractName, "sell", array2json([]interface{}{acc.ID, acc.ID, 6000}), acc.ID, acc.KeyPair)
				So(err, ShouldEqual, nil)
				So(r.Status.Code, ShouldEqual, tx.StatusCode(tx.ErrorRuntime))
			})
			Convey("normal sell", func() {
				var sellAmount int64 = 341
				//balanceBefore := s.Visitor.TokenBalance("iost", acc.ID)
				ramAvailableBefore := s.Visitor.TokenBalance("ram", contractName)
				myRAMBefore := s.Visitor.TokenBalance("ram", acc.ID)
				r, err := s.Call(contractName, "sell", array2json([]interface{}{acc.ID, acc.ID, sellAmount}), acc.ID, acc.KeyPair)
				So(err, ShouldEqual, nil)
				So(r.Status.Message, ShouldEqual, "")
				balanceAfter := s.Visitor.TokenBalance("iost", acc.ID)
				ramAvailableAfter := s.Visitor.TokenBalance("ram", contractName)
				myRAMAfter := s.Visitor.TokenBalance("ram", acc.ID)
				//var priceEstimated int64 = 10 * 1e8 // TODO when the final function is set, update here
				So(balanceAfter, ShouldEqual, 91916 * 1e6)
				So(myRAMAfter, ShouldEqual, myRAMBefore-sellAmount)
				So(ramAvailableAfter, ShouldEqual, ramAvailableBefore+sellAmount)
			})
			Convey("user can sell ram for others", func() {
				var sellAmount int64 = 341
				balanceBefore := s.Visitor.TokenBalance("iost", acc.ID)
				//otherBalanceBefore := s.Visitor.TokenBalance("iost", other)
				myRAMBefore := s.Visitor.TokenBalance("ram", acc.ID)
				r, err := s.Call(contractName, "sell", array2json([]interface{}{acc.ID, other, sellAmount}), acc.ID, acc.KeyPair)
				So(err, ShouldEqual, nil)
				So(r.Status.Message, ShouldEqual, "")
				balanceAfter := s.Visitor.TokenBalance("iost", acc.ID)
				otherBalanceAfter := s.Visitor.TokenBalance("iost", other)
				myRAMAfter := s.Visitor.TokenBalance("ram", acc.ID)
				//var priceEstimated int64 = 10 * 1e8 // TODO when the final function is set, update here
				So(balanceAfter, ShouldEqual, balanceBefore)
				So(myRAMAfter, ShouldEqual, myRAMBefore-sellAmount)
				So(otherBalanceAfter, ShouldEqual, 998 * 1e6)
			})
		})
	})
}

