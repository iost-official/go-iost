package integration

import (
	"fmt"
	"testing"

	"encoding/json"

	"github.com/iost-official/go-iost/account"
	"github.com/iost-official/go-iost/common"
	"github.com/iost-official/go-iost/core/tx"
	"github.com/iost-official/go-iost/crypto"
	"github.com/iost-official/go-iost/ilog"
	. "github.com/iost-official/go-iost/verifier"
	"github.com/iost-official/go-iost/vm/native"
	. "github.com/smartystreets/goconvey/convey"
)

func TestTransfer(t *testing.T) {
	ilog.Stop()

	s := NewSimulator()
	defer s.Clear()
	kp := prepareAuth(t, s)

	s.SetGas(kp.ID, 1000)

	prepareContract(t, s)

	r, err := s.Call("iost.token", "transfer", fmt.Sprintf(`["iost","%v","%v","%v"]`, testID[0], testID[2], 0.0001), kp.ID, kp)

	Convey("test transfer success case", t, func() {
		So(err, ShouldBeNil)
		So(r.Status.Message, ShouldEqual, "")
		So(s.Visitor.TokenBalance("iost", testID[0]), ShouldEqual, int64(99999990000))
		So(s.Visitor.TokenBalance("iost", testID[2]), ShouldEqual, int64(10000))
		So(s.Visitor.CurrentTotalGas(kp.ID, 0).Value, ShouldEqual, int64(99776600000000))
	})
}

func TestSetCode(t *testing.T) {
	ilog.Stop()
	Convey("set code", t, func() {
		s := NewSimulator()
		defer s.Clear()
		kp := prepareAuth(t, s)
		s.SetAccount(account.NewInitAccount(kp.ID, kp.ID, kp.ID))
		s.SetGas(kp.ID, 10000)

		c, err := s.Compile("hw", "test_data/helloworld", "test_data/helloworld")
		So(err, ShouldBeNil)
		cname, err := s.DeployContract(c, kp.ID, kp)
		So(err, ShouldBeNil)
		So(cname, ShouldStartWith, "Contract")

		So(s.Visitor.CurrentTotalGas(kp.ID, 0).Value, ShouldEqual, int64(9998600000000)) // todo check gas

		r, err := s.Call(cname, "hello", "[]", kp.ID, kp)
		So(err, ShouldBeNil)
		So(r.Status.Message, ShouldEqual, "")
	})
}

func TestJS_Database(t *testing.T) {
	t.Skip()
	//ilog.Stop()
	Convey("test of s database", t, func() {
		s := NewSimulator()
		defer s.Clear()

		c, err := s.Compile("datatbase", "test_data/database", "test_data/database")
		if err != nil {
			t.Fatal(err)
		}

		kp := prepareAuth(t, s)
		s.SetGas(kp.ID, 1000)

		cname, err := s.DeployContract(c, kp.ID, kp)
		So(err, ShouldBeNil)
		t.Log("cname ", cname)

		So(s.Visitor.Contract(cname), ShouldNotBeNil)
		So(s.Visitor.Get(cname+"-"+"num"), ShouldEqual, "s9")
		So(s.Visitor.Get(cname+"-"+"string"), ShouldEqual, "shello")
		So(s.Visitor.Get(cname+"-"+"bool"), ShouldEqual, "strue")
		So(s.Visitor.Get(cname+"-"+"array"), ShouldEqual, "s[1,2,3]")
		So(s.Visitor.Get(cname+"-"+"obj"), ShouldEqual, `s{"foo":"bar"}`)

		r, err := s.Call(cname, "read", `[]`, kp.ID, kp)

		So(err, ShouldBeNil)
		So(r.Status.Message, ShouldEqual, "")
	})

}

func TestAmountLimit(t *testing.T) {
	ilog.Stop()
	Convey("test of amount limit", t, func() {
		s := NewSimulator()
		defer s.Clear()
		prepareContract(t, s)

		ca, err := s.Compile("Contracttransfer", "./test_data/transfer", "./test_data/transfer.js")
		if err != nil || ca == nil {
			t.Fatal(err)
		}
		s.SetContract(ca)

		ca, err = s.Compile("Contracttransfer1", "./test_data/transfer1", "./test_data/transfer1.js")
		if err != nil || ca == nil {
			t.Fatal(err)
		}
		s.SetContract(ca)

		kp, err := account.NewKeyPair(common.Base58Decode(testID[1]), crypto.Secp256k1)
		if err != nil {
			t.Fatal(err)
		}

		Reset(func() {
			s.Visitor.SetTokenBalanceFixed("iost", testID[0], "1000")
			s.Visitor.SetTokenBalanceFixed("iost", testID[2], "0")
			s.SetGas(kp.ID, 10000)
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
			So(r.Status.Code, ShouldEqual, tx.ErrorRuntime)
			So(r.Status.Message, ShouldContainSubstring, "exceed amountLimit in abi")
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
		prepareContract(t, s)

		kp, err := account.NewKeyPair(common.Base58Decode(testID[1]), crypto.Secp256k1)
		if err != nil {
			t.Fatal(err)
		}
		s.SetGas(kp.ID, 10000)

		Convey("test out of gas limit", func() {
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

func array2json(ss []interface{}) string {
	x, err := json.Marshal(ss)
	if err != nil {
		panic(err)
	}
	return string(x)
}

func TestAuthority(t *testing.T) {
	s := NewSimulator()
	defer s.Clear()

	ca, err := s.Compile("iost.auth", "../../contract/account", "../../contract/account.js")
	if err != nil {
		t.Fatal(err)
	}
	s.Visitor.SetContract(ca)

	kp := prepareAuth(t, s)
	s.SetGas(kp.ID, 1000)

	Convey("test of Auth", t, func() {
		s.Call("iost.auth", "SignUp", array2json([]interface{}{"myid", "okey", "akey"}), kp.ID, kp)
		So(s.Visitor.MGet("iost.auth-account", "myid"), ShouldEqual, `s{"id":"myid","permissions":{"active":{"name":"active","groups":[],"items":[{"id":"akey","is_key_pair":true,"weight":1}],"threshold":1},"owner":{"name":"owner","groups":[],"items":[{"id":"okey","is_key_pair":true,"weight":1}],"threshold":1}}}`)

		s.Call("iost.auth", "AddPermission", array2json([]interface{}{"myid", "perm1", 1}), kp.ID, kp)
		So(s.Visitor.MGet("iost.auth-account", "myid"), ShouldContainSubstring, `"perm1":{"name":"perm1","groups":[],"items":[],"threshold":1}`)
	})

}
