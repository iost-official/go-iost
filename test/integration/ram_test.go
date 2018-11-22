package integration

import (
	"testing"

	"github.com/iost-official/go-iost/account"
	"github.com/iost-official/go-iost/common"
	"github.com/iost-official/go-iost/core/tx"
	"github.com/iost-official/go-iost/crypto"
	. "github.com/iost-official/go-iost/verifier"
	. "github.com/smartystreets/goconvey/convey"
)


func TestRAM(t *testing.T) {
	s := NewSimulator()
	defer s.Clear()

	prepareContract(s)
	contractName := "ram.iost"
	err := setNonNativeContract(s, contractName, "ram.js", ContractPath)
	if err != nil {
		t.Fatal(err)
	}

	admin, err := account.NewKeyPair(common.Base58Decode(testID[3]), crypto.Secp256k1)
	if err != nil {
		t.Fatal(err)
	}
	kp := prepareAuth(t, s)
	createToken(t, s, kp)
	s.SetGas(kp.ID, 1000000)

	s.Head.Number = 0
	r, err := s.Call(contractName, "initAdmin", array2json([]interface{}{admin.ID}), admin.ID, admin)
	if err != nil || r.Status.Code != tx.StatusCode(tx.Success) {
		panic("call failed " + err.Error() + " " + r.String())
	}
	r, err = s.Call(contractName, "initContractName", array2json([]interface{}{contractName}), admin.ID, admin)
	if err != nil || r.Status.Code != tx.StatusCode(tx.Success) {
		panic("call failed " + err.Error() + " " + r.String())
	}

	var initialTotal int64 = 128 * 1024 * 1024 * 1024
	var increaseInterval int64 = 24 * 3600
	var increaseAmount int64 = 64 * 1024 * 1024 * 1024 / 365
	r, err = s.Call(contractName, "issue", array2json([]interface{}{initialTotal, increaseInterval, increaseAmount, 0}), admin.ID, admin)
	if err != nil || r.Status.Code != tx.StatusCode(tx.Success) {
		panic("call failed " + err.Error() + " " + r.String())
	}
	initRAM := s.Visitor.TokenBalance("ram", kp.ID)

	s.Head.Number = 1
	Convey("test of ram", t, func() {
		Convey("user has no ram if he did not buy", func() {
			So(s.Visitor.TokenBalance("ram", kp.ID), ShouldEqual, initRAM)
		})
		Convey("test buy", func() {
			var buyAmount int64 = 1024
			Convey("normal buy", func() {
				//balanceBefore := s.Visitor.TokenBalance("iost", kp.ID)
				ramAvailableBefore := s.Visitor.TokenBalance("ram", contractName)
				r, err := s.Call(contractName, "buy", array2json([]interface{}{kp.ID, kp.ID, buyAmount}), kp.ID, kp)
				So(err, ShouldEqual, nil)
				So(r.Status.Message, ShouldEqual, "")
				balanceAfter := s.Visitor.TokenBalance("iost", kp.ID)
				ramAvailableAfter := s.Visitor.TokenBalance("ram", contractName)
				//var priceEstimated int64 = 30 * 1e8 // TODO when the final function is set, update here
				So(balanceAfter, ShouldEqual, 968 * 1e8)
				So(s.Visitor.TokenBalance("ram", kp.ID), ShouldEqual, initRAM+buyAmount)
				So(ramAvailableAfter, ShouldEqual, ramAvailableBefore-buyAmount)
			})
			Convey("when buying triggers increasing total ram", func() {
				head := s.Head
				head.Time = head.Time + increaseInterval*1000*1000*1000
				s.SetBlockHead(head)
				ramAvailableBefore := s.Visitor.TokenBalance("ram", contractName)
				r, err := s.Call(contractName, "buy", array2json([]interface{}{kp.ID, kp.ID, buyAmount}), kp.ID, kp)
				So(err, ShouldEqual, nil)
				So(r.Status.Message, ShouldEqual, "")
				ramAvailableAfter := s.Visitor.TokenBalance("ram", contractName)
				So(ramAvailableAfter, ShouldEqual, ramAvailableBefore+increaseAmount-buyAmount)
			})
			Convey("user can buy for others", func() {
				other := testID[4]
				//balanceBefore := s.Visitor.TokenBalance("iost", kp.ID)
				otherRAMBefore := s.Visitor.TokenBalance("ram", other)
				myRAMBefore := s.Visitor.TokenBalance("ram", kp.ID)
				r, err := s.Call(contractName, "buy", array2json([]interface{}{kp.ID, other, buyAmount}), kp.ID, kp)
				So(err, ShouldEqual, nil)
				So(r.Status.Message, ShouldEqual, "")
				balanceAfter := s.Visitor.TokenBalance("iost", kp.ID)
				otherRAMAfter := s.Visitor.TokenBalance("ram", other)
				myRAMAfter := s.Visitor.TokenBalance("ram", kp.ID)
				//var priceEstimated int64 = 30 * 1e8 // TODO when the final function is set, update here
				So(balanceAfter, ShouldEqual, 906 * 1e8)
				So(myRAMAfter, ShouldEqual, myRAMBefore)
				So(otherRAMAfter, ShouldEqual, otherRAMBefore+buyAmount)
			})
		})
		Convey("test sell", func() {
			Convey("user cannot sell more than he owns", func() {
				r, err := s.Call(contractName, "sell", array2json([]interface{}{kp.ID, kp.ID, 6000}), kp.ID, kp)
				So(err, ShouldEqual, nil)
				So(r.Status.Code, ShouldEqual, tx.StatusCode(tx.ErrorRuntime))
			})
			Convey("normal sell", func() {
				var sellAmount int64 = 341
				//balanceBefore := s.Visitor.TokenBalance("iost", kp.ID)
				ramAvailableBefore := s.Visitor.TokenBalance("ram", contractName)
				myRAMBefore := s.Visitor.TokenBalance("ram", kp.ID)
				r, err := s.Call(contractName, "sell", array2json([]interface{}{kp.ID, kp.ID, sellAmount}), kp.ID, kp)
				So(err, ShouldEqual, nil)
				So(r.Status.Message, ShouldEqual, "")
				balanceAfter := s.Visitor.TokenBalance("iost", kp.ID)
				ramAvailableAfter := s.Visitor.TokenBalance("ram", contractName)
				myRAMAfter := s.Visitor.TokenBalance("ram", kp.ID)
				//var priceEstimated int64 = 10 * 1e8 // TODO when the final function is set, update here
				So(balanceAfter, ShouldEqual, 916 * 1e8)
				So(myRAMAfter, ShouldEqual, myRAMBefore-sellAmount)
				So(ramAvailableAfter, ShouldEqual, ramAvailableBefore+sellAmount)
			})
			Convey("user can sell ram for others", func() {
				var sellAmount int64 = 341
				other := testID[4]
				balanceBefore := s.Visitor.TokenBalance("iost", kp.ID)
				//otherBalanceBefore := s.Visitor.TokenBalance("iost", other)
				myRAMBefore := s.Visitor.TokenBalance("ram", kp.ID)
				r, err := s.Call(contractName, "sell", array2json([]interface{}{kp.ID, other, sellAmount}), kp.ID, kp)
				So(err, ShouldEqual, nil)
				So(r.Status.Message, ShouldEqual, "")
				balanceAfter := s.Visitor.TokenBalance("iost", kp.ID)
				otherBalanceAfter := s.Visitor.TokenBalance("iost", other)
				myRAMAfter := s.Visitor.TokenBalance("ram", kp.ID)
				//var priceEstimated int64 = 10 * 1e8 // TODO when the final function is set, update here
				So(balanceAfter, ShouldEqual, balanceBefore)
				So(myRAMAfter, ShouldEqual, myRAMBefore-sellAmount)
				So(otherBalanceAfter, ShouldEqual, 10 * 1e8)
			})
		})
	})
}

