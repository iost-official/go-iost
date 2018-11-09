package integration

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/iost-official/go-iost/account"
	"github.com/iost-official/go-iost/common"
	"github.com/iost-official/go-iost/core/tx"
	"github.com/iost-official/go-iost/crypto"
	"github.com/iost-official/go-iost/ilog"
	. "github.com/iost-official/go-iost/verifier"
	"github.com/iost-official/go-iost/vm/native"
	. "github.com/smartystreets/goconvey/convey"
)

func array2json(ss []interface{}) string {
	x, err := json.Marshal(ss)
	if err != nil {
		panic(err)
	}
	return string(x)
}

func TestTransfer(t *testing.T) {
	ilog.Stop()

	s := NewSimulator()
	defer s.Clear()
	kp := prepareAuth(t, s)

	s.SetGas(kp.ID, 1000)
	Convey("test transfer success case", t, func() {
		prepareContract(s)
		createToken(t, s, kp)

		totalGas := s.Visitor.CurrentTotalGas(kp.ID, 0).Value

		r, err := s.Call("iost.token", "transfer", fmt.Sprintf(`["iost","%v","%v","%v"]`, testID[0], testID[2], 0.0001), kp.ID, kp)

		So(err, ShouldBeNil)
		So(r.Status.Message, ShouldEqual, "")
		So(s.Visitor.TokenBalance("iost", testID[0]), ShouldEqual, int64(99999990000))
		So(s.Visitor.TokenBalance("iost", testID[2]), ShouldEqual, int64(10000))
		So(totalGas-s.Visitor.CurrentTotalGas(kp.ID, 0).Value, ShouldEqual, int64(90900000000))
	})
}

func TestSetCode(t *testing.T) {
	ilog.SetLevel(ilog.LevelInfo)
	Convey("set code", t, func() {
		s := NewSimulator()
		defer s.Clear()
		kp := prepareAuth(t, s)
		s.SetAccount(account.NewInitAccount(kp.ID, kp.ID, kp.ID))
		s.SetGas(kp.ID, 10000)
		s.SetRAM(kp.ID, 300)

		c, err := s.Compile("hw", "test_data/helloworld", "test_data/helloworld")
		So(err, ShouldBeNil)
		So(len(c.Encode()), ShouldEqual, 146)
		cname, err := s.DeployContract(c, kp.ID, kp)
		So(err, ShouldBeNil)
		So(cname, ShouldStartWith, "Contract")

		So(s.Visitor.CurrentTotalGas(kp.ID, 0).Value, ShouldEqual, int64(9956000000000))
		So(s.Visitor.TokenBalance("ram", kp.ID), ShouldBeBetweenOrEqual, int64(62), int64(63))

		r, err := s.Call(cname, "hello", "[]", kp.ID, kp)
		So(err, ShouldBeNil)
		So(r.Status.Message, ShouldEqual, "")
	})
}

func TestJS_Database(t *testing.T) {
	//ilog.Stop()
	ilog.SetLevel(ilog.LevelInfo)
	Convey("test of s database", t, func() {
		s := NewSimulator()
		defer s.Clear()

		c, err := s.Compile("datatbase", "test_data/database", "test_data/database")
		So(err, ShouldBeNil)

		kp := prepareAuth(t, s)
		s.SetGas(kp.ID, 1000)
		s.SetRAM(kp.ID, 3000)

		cname, err := s.DeployContract(c, kp.ID, kp)
		So(err, ShouldBeNil)

		So(s.Visitor.Contract(cname), ShouldNotBeNil)
		So(s.Visitor.Get(cname+"-"+"num"), ShouldEqual, "s9")
		So(s.Visitor.Get(cname+"-"+"string"), ShouldEqual, "shello")
		So(s.Visitor.Get(cname+"-"+"bool"), ShouldEqual, "strue")
		So(s.Visitor.Get(cname+"-"+"array"), ShouldEqual, "s[1,2,3]")
		So(s.Visitor.Get(cname+"-"+"obj"), ShouldEqual, `s{"foo":"bar"}`)

		r, err := s.Call(cname, "read", `[]`, kp.ID, kp)

		So(err, ShouldBeNil)
		So(r.Status.Message, ShouldEqual, "")
		So(len(r.Returns), ShouldEqual, 1)
		So(r.Returns[0].Value, ShouldEqual, `["true"]`)
	})

}

func TestAmountLimit(t *testing.T) {
	ilog.Stop()
	Convey("test of amount limit", t, func() {
		s := NewSimulator()
		defer s.Clear()
		prepareContract(s)

		ca, err := s.Compile("Contracttransfer", "./test_data/transfer", "./test_data/transfer.js")
		So(err, ShouldBeNil)
		So(ca, ShouldNotBeNil)
		s.SetContract(ca)

		ca, err = s.Compile("Contracttransfer1", "./test_data/transfer1", "./test_data/transfer1.js")
		So(err, ShouldBeNil)
		So(ca, ShouldNotBeNil)
		s.SetContract(ca)

		kp, err := account.NewKeyPair(common.Base58Decode(testID[1]), crypto.Secp256k1)
		So(err, ShouldBeNil)

		s.SetRAM(testID[0], 10000)

		createToken(t, s, kp)

		Reset(func() {
			s.Visitor.SetTokenBalanceFixed("iost", testID[0], "1000")
			s.Visitor.SetTokenBalanceFixed("iost", testID[2], "0")
			s.SetGas(kp.ID, 10000)
			s.SetRAM(testID[0], 10000)
		})

		Convey("test of amount limit", func() {
			r, err := s.Call("Contracttransfer", "transfer", fmt.Sprintf(`["%v", "%v", "%v"]`, testID[0], testID[2], "10"), testID[0], kp)
			s.Visitor.Commit()

			So(err, ShouldBeNil)
			So(r.Status.Code, ShouldEqual, tx.Success)
			balance0 := common.Fixed{Value: s.Visitor.TokenBalance("iost", testID[0]), Decimal: s.Visitor.Decimal("iost")}
			balance2 := common.Fixed{Value: s.Visitor.TokenBalance("iost", testID[2]), Decimal: s.Visitor.Decimal("iost")}
			So(balance0.ToString(), ShouldEqual, "990")
			So(balance2.ToString(), ShouldEqual, "10")
		})

		Convey("test out of amount limit", func() {
			r, err := s.Call("Contracttransfer", "transfer", fmt.Sprintf(`["%v", "%v", "%v"]`, testID[0], testID[2], "110"), testID[0], kp)
			s.Visitor.Commit()

			So(err, ShouldBeNil)
			So(r.Status.Message, ShouldContainSubstring, "exceed amountLimit in abi")
			So(r.Status.Code, ShouldEqual, tx.ErrorRuntime)
			//balance0 := common.Fixed{Value:s.Visitor.TokenBalance("iost", testID[0]), Decimal:s.Visitor.Decimal("iost")}
			//balance2 := common.Fixed{Value:s.Visitor.TokenBalance("iost", testID[2]), Decimal:s.Visitor.Decimal("iost")}
			// todo exit when monitor.Call return err
			// So(balance0.ToString(), ShouldEqual, "990")
			// So(balance2.ToString(), ShouldEqual, "10")
		})

		Convey("test amount limit two level invocation", func() {
			r, err := s.Call("Contracttransfer1", "transfer", fmt.Sprintf(`["%v", "%v", "%v"]`, testID[0], testID[2], "120"), testID[0], kp)
			s.Visitor.Commit()

			So(err, ShouldBeNil)
			So(r.Status.Code, ShouldEqual, tx.Success)
			balance0 := common.Fixed{Value: s.Visitor.TokenBalance("iost", testID[0]), Decimal: s.Visitor.Decimal("iost")}
			balance2 := common.Fixed{Value: s.Visitor.TokenBalance("iost", testID[2]), Decimal: s.Visitor.Decimal("iost")}
			So(balance0.ToString(), ShouldEqual, "880")
			So(balance2.ToString(), ShouldEqual, "120")
		})

	})
}

func TestNativeVM_GasLimit(t *testing.T) {
	ilog.Stop()
	Convey("test of amount limit", t, func() {
		s := NewSimulator()
		defer s.Clear()
		prepareContract(s)

		kp, err := account.NewKeyPair(common.Base58Decode(testID[1]), crypto.Secp256k1)
		if err != nil {
			t.Fatal(err)
		}
		createToken(t, s, kp)
		s.SetGas(kp.ID, 10000)

		tx0 := tx.NewTx([]*tx.Action{{
			Contract:   "iost.token",
			ActionName: "transfer",
			Data:       fmt.Sprintf(`["iost", "%v", "%v", "%v"]`, testID[0], testID[2], "10"),
		}}, nil, 100, 100, 10000000, 0)

		r, err := s.CallTx(tx0, testID[0], kp)
		s.Visitor.Commit()
		So(err, ShouldBeNil)
		So(r.Status.Code, ShouldEqual, tx.ErrorRuntime)
		So(r.Status.Message, ShouldContainSubstring, "gas limit exceeded")

	})
}

func TestDomain(t *testing.T) {
	Convey("test of domain", t, func() {
		s := NewSimulator()
		defer s.Clear()

		c, err := s.Compile("datatbase", "test_data/database", "test_data/database")
		So(err, ShouldBeNil)

		kp := prepareAuth(t, s)
		s.SetGas(kp.ID, 1000)
		s.SetRAM(kp.ID, 3000)

		cname, err := s.DeployContract(c, kp.ID, kp)
		So(err, ShouldBeNil)
		s.Visitor.SetContract(native.ABI("iost.domain", native.DomainABIs))
		r1, err := s.Call("iost.domain", "Link", fmt.Sprintf(`["abcde","%v"]`, cname), kp.ID, kp)
		So(err, ShouldBeNil)
		So(r1.Status.Message, ShouldEqual, "")
		r2, err := s.Call("abcde", "read", "[]", kp.ID, kp)
		So(err, ShouldBeNil)
		So(r2.Status.Message, ShouldEqual, "")
	})
}

func TestAuthority(t *testing.T) {
	s := NewSimulator()
	defer s.Clear()
	Convey("test of Auth", t, func() {

		ca, err := s.Compile("iost.auth", "../../contract/account", "../../contract/account.js")
		So(err, ShouldBeNil)
		s.Visitor.SetContract(ca)

		kp := prepareAuth(t, s)
		s.SetGas(kp.ID, 1000)

		r, err := s.Call("iost.auth", "SignUp", array2json([]interface{}{"myid", "okey", "akey"}), kp.ID, kp)
		So(err, ShouldBeNil)
		So(r.Status.Message, ShouldEqual, "")
		So(s.Visitor.MGet("iost.auth-account", "myid"), ShouldEqual, `s{"id":"myid","permissions":{"active":{"name":"active","groups":[],"items":[{"id":"akey","is_key_pair":true,"weight":1}],"threshold":1},"owner":{"name":"owner","groups":[],"items":[{"id":"okey","is_key_pair":true,"weight":1}],"threshold":1}}}`)

		r, err = s.Call("iost.auth", "AddPermission", array2json([]interface{}{"myid", "perm1", 1}), kp.ID, kp)
		So(err, ShouldBeNil)
		So(r.Status.Message, ShouldEqual, "")
		So(s.Visitor.MGet("iost.auth-account", "myid"), ShouldContainSubstring, `"perm1":{"name":"perm1","groups":[],"items":[],"threshold":1}`)
	})

}

func TestRAM(t *testing.T) {
	s := NewSimulator()
	defer s.Clear()

	prepareContract(s)
	contractName := "iost.ram"
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
	s.SetGas(kp.ID, 1000)

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
	r, err = s.Call(contractName, "issue", array2json([]interface{}{initialTotal, increaseInterval, increaseAmount}), admin.ID, admin)
	if err != nil || r.Status.Code != tx.StatusCode(tx.Success) {
		panic("call failed " + err.Error() + " " + r.String())
	}
	initRAM := s.Visitor.TokenBalance("ram", kp.ID)

	Convey("test of ram", t, func() {
		Convey("user has no ram if he did not buy", func() {
			So(s.Visitor.TokenBalance("ram", kp.ID), ShouldEqual, initRAM)
		})
		Convey("test buy", func() {
			var buyAmount int64 = 30
			Convey("user can only buy for himself", func() {
				r, err := s.Call(contractName, "buy", array2json([]interface{}{testID[4], 1234}), kp.ID, kp)
				So(err, ShouldEqual, nil)
				So(r.Status.Code, ShouldEqual, tx.StatusCode(tx.ErrorRuntime))
			})
			Convey("normal buy", func() {
				balanceBefore := s.Visitor.TokenBalance("iost", kp.ID)
				ramAvailableBefore := s.Visitor.TokenBalance("ram", contractName)
				r, err := s.Call(contractName, "buy", array2json([]interface{}{kp.ID, buyAmount}), kp.ID, kp)
				So(err, ShouldEqual, nil)
				So(r.Status.Message, ShouldEqual, "")
				balanceAfter := s.Visitor.TokenBalance("iost", kp.ID)
				ramAvailableAfter := s.Visitor.TokenBalance("ram", contractName)
				var priceEstimated int64 = 30 * 1e8 // TODO when the final function is set, update here
				So(balanceAfter, ShouldEqual, balanceBefore-priceEstimated)
				So(s.Visitor.TokenBalance("ram", kp.ID), ShouldEqual, initRAM+buyAmount)
				So(ramAvailableAfter, ShouldEqual, ramAvailableBefore-buyAmount)
			})
			Convey("when buying triggers increasing total ram", func() {
				head := s.Head
				head.Time = head.Time + increaseInterval*1000*1000*1000
				s.SetBlockHead(head)
				ramAvailableBefore := s.Visitor.TokenBalance("ram", contractName)
				r, err := s.Call(contractName, "buy", array2json([]interface{}{kp.ID, buyAmount}), kp.ID, kp)
				So(err, ShouldEqual, nil)
				So(r.Status.Message, ShouldEqual, "")
				ramAvailableAfter := s.Visitor.TokenBalance("ram", contractName)
				So(ramAvailableAfter, ShouldEqual, ramAvailableBefore+increaseAmount-buyAmount)
			})
		})
		Convey("test sell", func() {
			Convey("user can only sell ram of himself", func() {
				r, err := s.Call(contractName, "sell", array2json([]interface{}{testID[4], 10}), kp.ID, kp)
				So(err, ShouldEqual, nil)
				So(r.Status.Code, ShouldEqual, tx.StatusCode(tx.ErrorRuntime))
			})
			Convey("user cannot sell more than he owns", func() {
				r, err := s.Call(contractName, "sell", array2json([]interface{}{kp.ID, 6000}), kp.ID, kp)
				So(err, ShouldEqual, nil)
				So(r.Status.Code, ShouldEqual, tx.StatusCode(tx.ErrorRuntime))
			})
			Convey("normal sell", func() {
				var sellAmount int64 = 10
				balanceBefore := s.Visitor.TokenBalance("iost", kp.ID)
				ramAvailableBefore := s.Visitor.TokenBalance("ram", contractName)
				r, err := s.Call(contractName, "sell", array2json([]interface{}{kp.ID, sellAmount}), kp.ID, kp)
				So(err, ShouldEqual, nil)
				So(r.Status.Message, ShouldEqual, "")
				balanceAfter := s.Visitor.TokenBalance("iost", kp.ID)
				ramAvailableAfter := s.Visitor.TokenBalance("ram", contractName)
				var priceEstimated int64 = 10 * 1e8 // TODO when the final function is set, update here
				So(balanceAfter, ShouldEqual, balanceBefore+priceEstimated)
				So(s.Visitor.TokenBalance("ram", kp.ID), ShouldEqual, initRAM+50)
				So(ramAvailableAfter, ShouldEqual, ramAvailableBefore+sellAmount)
			})
		})
	})
}
