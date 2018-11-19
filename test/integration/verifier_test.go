package integration

import (
	"fmt"
	"testing"

	"github.com/iost-official/go-iost/account"
	"github.com/iost-official/go-iost/common"
	"github.com/iost-official/go-iost/core/contract"
	"github.com/iost-official/go-iost/core/tx"
	"github.com/iost-official/go-iost/crypto"
	"github.com/iost-official/go-iost/ilog"
	. "github.com/iost-official/go-iost/verifier"
	"github.com/iost-official/go-iost/vm"
	"github.com/iost-official/go-iost/vm/native"
	. "github.com/smartystreets/goconvey/convey"
)

func TestTransfer(t *testing.T) {
	ilog.Stop()

	Convey("test transfer success case", t, func() {
		s := NewSimulator()
		defer s.Clear()
		kp := prepareAuth(t, s)

		s.SetGas(kp.ID, 100000)
		prepareContract(s)
		createToken(t, s, kp)

		Reset(func() {
			s.Visitor.SetTokenBalanceFixed("iost", testID[0], "1000")
			s.Visitor.SetTokenBalanceFixed("iost", testID[2], "0")
			s.SetGas(kp.ID, 100000)
			s.SetRAM(testID[0], 10000)
		})

		Convey("test transfer success case", func() {
			r, err := s.Call("token.iost", "transfer", fmt.Sprintf(`["iost","%v","%v","%v",""]`, testID[0], testID[2], 0.0001), kp.ID, kp)
			So(err, ShouldBeNil)
			So(r.Status.Message, ShouldEqual, "")
			So(s.Visitor.TokenBalance("iost", testID[0]), ShouldEqual, int64(99999990000))
			So(s.Visitor.TokenBalance("iost", testID[2]), ShouldEqual, int64(10000))
			So(r.GasUsage, ShouldEqual, 721)
		})

		Convey("test of token memo", func() {
			r, err := s.Call("token.iost", "transfer", fmt.Sprintf(`["iost","%v","%v","%v","memo"]`, testID[0], testID[2], 0.0001), kp.ID, kp)
			So(err, ShouldBeNil)
			So(r.Status.Code, ShouldEqual, tx.Success)

			memo := "test"
			for i := 0; i < 10; i++ {
				memo = memo + memo
			}
			r, err = s.Call("token.iost", "transfer", fmt.Sprintf(`["iost","%v","%v","%v","%v"]`, testID[0], testID[2], 0.0001, memo), kp.ID, kp)
			So(err, ShouldBeNil)
			So(r.Status.Code, ShouldEqual, tx.ErrorRuntime)
			So(r.Status.Message, ShouldContainSubstring, "memo too large")
		})
	})
}

func TestSetCode(t *testing.T) {
	ilog.SetLevel(ilog.LevelInfo)
	Convey("set code", t, func() {
		s := NewSimulator()
		defer s.Clear()
		kp := prepareAuth(t, s)
		s.SetAccount(account.NewInitAccount(kp.ID, kp.ID, kp.ID))
		s.SetGas(kp.ID, 1000000)
		s.SetRAM(kp.ID, 300)

		c, err := s.Compile("hw", "test_data/helloworld", "test_data/helloworld")
		So(err, ShouldBeNil)
		So(len(c.Encode()), ShouldEqual, 146)
		cname, r, err := s.DeployContract(c, kp.ID, kp)
		So(err, ShouldBeNil)
		So(r.Status.Code, ShouldEqual, tx.Success)
		So(cname, ShouldEqual, "ContractEJuvctjsCVirp9g22As7KbrM71783oq4wYE1Fcy8AXns")
		So(r.GasUsage, ShouldEqual, 1548)
		So(s.Visitor.TokenBalance("ram", kp.ID), ShouldEqual, int64(64))

		r, err = s.Call(cname, "hello", "[]", kp.ID, kp)
		So(err, ShouldBeNil)
		So(r.Status.Message, ShouldEqual, "")
	})
}

func TestStringGas(t *testing.T) {
	ilog.SetLevel(ilog.LevelInfo)
	Convey("string op gas", t, func() {
		s := NewSimulator()
		defer s.Clear()
		kp := prepareAuth(t, s)
		s.SetAccount(account.NewInitAccount(kp.ID, kp.ID, kp.ID))
		s.SetGas(kp.ID, 1000000)
		s.SetRAM(kp.ID, 1000)

		c, err := s.Compile("so", "test_data/stringop", "test_data/stringop")
		So(err, ShouldBeNil)
		So(c, ShouldNotBeNil)
		cname, r, err := s.DeployContract(c, kp.ID, kp)
		So(err, ShouldBeNil)
		So(r.Status.Code, ShouldEqual, tx.Success)

		r, err = s.Call(cname, "add2", "[]", kp.ID, kp)
		So(err, ShouldBeNil)
		So(r.Status.Code, ShouldEqual, 0)
		gas2 := r.GasUsage

		r, err = s.Call(cname, "add9", "[]", kp.ID, kp)
		So(err, ShouldBeNil)
		So(r.Status.Code, ShouldEqual, 0)
		So(r.GasUsage-gas2, ShouldBeBetweenOrEqual, 12, 14)

		r, err = s.Call(cname, "equal9", "[]", kp.ID, kp)
		So(err, ShouldBeNil)
		So(r.Status.Code, ShouldEqual, 0)
		So(r.GasUsage-gas2, ShouldEqual, 14)

		r, err = s.Call(cname, "superadd9", "[]", kp.ID, kp)
		So(err, ShouldBeNil)
		So(r.Status.Code, ShouldEqual, 0)
		So(r.GasUsage-gas2, ShouldBeGreaterThan, 14)
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
		s.SetGas(kp.ID, 100000)
		s.SetRAM(kp.ID, 3000)

		cname, _, err := s.DeployContract(c, kp.ID, kp)
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
		So(r.Returns[0], ShouldEqual, `["true"]`)
	})

}

func TestAmountLimit(t *testing.T) {
	ilog.Stop()
	Convey("test of amount limit", t, func() {
		s := NewSimulator()
		defer s.Clear()
		prepareContract(s)

		kp, err := account.NewKeyPair(common.Base58Decode(testID[1]), crypto.Secp256k1)
		So(err, ShouldBeNil)

		createToken(t, s, kp)

		ca, err := s.Compile("Contracttransfer", "./test_data/transfer", "./test_data/transfer.js")
		So(err, ShouldBeNil)
		So(ca, ShouldNotBeNil)
		s.SetContract(ca)

		ca, err = s.Compile("Contracttransfer1", "./test_data/transfer1", "./test_data/transfer1.js")
		So(err, ShouldBeNil)
		So(ca, ShouldNotBeNil)
		contractTransfer1, _, err := s.DeployContract(ca, testID[0], kp)
		So(err, ShouldBeNil)

		s.SetRAM(testID[0], 10000)

		Reset(func() {
			s.Visitor.SetTokenBalanceFixed("iost", testID[0], "1000")
			s.Visitor.SetTokenBalanceFixed("iost", testID[2], "0")
			s.SetGas(kp.ID, 100000)
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

		Convey("test out of amount limit, use signers ID", func() {
			s.SetAccount(account.NewInitAccount("test0", testID[0], testID[0]))
			s.Visitor.SetTokenBalanceFixed("iost", "test0", "1000")
			s.SetGas("test0", 100000)
			s.SetRAM("test0", 10000)

			r, err := s.Call("Contracttransfer", "transfer", fmt.Sprintf(`["%v", "%v", "%v"]`, "test0", testID[2], "200"), "test0", kp)
			s.Visitor.Commit()

			So(err, ShouldBeNil)
			So(r.Status.Message, ShouldContainSubstring, "exceed amountLimit in abi")
			So(r.Status.Code, ShouldEqual, tx.ErrorRuntime)
		})

		Convey("test out of amount limit", func() {
			r, err := s.Call("Contracttransfer", "transfer", fmt.Sprintf(`["%v", "%v", "%v"]`, testID[0], testID[2], "110"), testID[0], kp)
			s.Visitor.Commit()

			So(err, ShouldBeNil)
			So(r.Status.Message, ShouldContainSubstring, "exceed amountLimit in abi")
			So(r.Status.Code, ShouldEqual, tx.ErrorRuntime)
		})

		Convey("test amount limit two level invocation", func() {
			r, err := s.Call(contractTransfer1, "transfer", fmt.Sprintf(`["%v", "%v", "%v"]`, testID[0], testID[2], "120"), testID[0], kp)
			s.Visitor.Commit()

			So(err, ShouldBeNil)
			So(r.Status.Code, ShouldEqual, tx.Success)
			balance0 := common.Fixed{Value: s.Visitor.TokenBalance("iost", testID[0]), Decimal: s.Visitor.Decimal("iost")}
			balance2 := common.Fixed{Value: s.Visitor.TokenBalance("iost", testID[2]), Decimal: s.Visitor.Decimal("iost")}
			So(balance0.ToString(), ShouldEqual, "880")
			So(balance2.ToString(), ShouldEqual, "120")
		})

		Convey("test invalid amount limit", func() {
			ca, err = s.Compile("Contracttransfer2", "./test_data/transfer2", "./test_data/transfer2.js")
			So(err, ShouldBeNil)
			So(ca, ShouldNotBeNil)
			_, _, err := s.DeployContract(ca, testID[0], kp)
			So(err.Error(), ShouldContainSubstring, "abnormal char in amount")
		})

	})
}

func TestTxAmountLimit(t *testing.T) {
	ilog.Stop()
	Convey("test of tx amount limit", t, func() {
		s := NewSimulator()
		defer s.Clear()
		prepareContract(s)

		kp, err := account.NewKeyPair(common.Base58Decode(testID[1]), crypto.Secp256k1)
		So(err, ShouldBeNil)

		createToken(t, s, kp)
		s.SetRAM(testID[0], 10000)

		Reset(func() {
			s.Visitor.SetTokenBalanceFixed("iost", testID[0], "1000")
			s.Visitor.SetTokenBalanceFixed("iost", testID[2], "0")
			s.SetGas(kp.ID, 100000)
			s.SetRAM(testID[0], 10000)
		})

		Convey("test of tx amount limit", func() {
			trx := tx.NewTx([]*tx.Action{{
				Contract:   "token.iost",
				ActionName: "transfer",
				Data:       fmt.Sprintf(`["iost", "%v", "%v", "%v", ""]`, testID[0], testID[2], "10"),
			}}, nil, 100000, 100, 10000000, 0)
			trx.AmountLimit = append(trx.AmountLimit, &contract.Amount{Token: "iost", Val: "100"})
			r, err := s.CallTx(trx, testID[0], kp)
			s.Visitor.Commit()

			So(err, ShouldBeNil)
			So(r.Status.Code, ShouldEqual, tx.Success)
			balance0 := common.Fixed{Value: s.Visitor.TokenBalance("iost", testID[0]), Decimal: s.Visitor.Decimal("iost")}
			balance2 := common.Fixed{Value: s.Visitor.TokenBalance("iost", testID[2]), Decimal: s.Visitor.Decimal("iost")}
			So(balance0.ToString(), ShouldEqual, "990")
			So(balance2.ToString(), ShouldEqual, "10")
		})

		Convey("test out of amount limit", func() {
			trx := tx.NewTx([]*tx.Action{{
				Contract:   "token.iost",
				ActionName: "transfer",
				Data:       fmt.Sprintf(`["iost", "%v", "%v", "%v", ""]`, testID[0], testID[2], "110"),
			}}, nil, 100000, 100, 10000000, 0)
			trx.AmountLimit = append(trx.AmountLimit, &contract.Amount{Token: "iost", Val: "100"})
			r, err := s.CallTx(trx, testID[0], kp)
			s.Visitor.Commit()

			So(err, ShouldBeNil)
			So(r.Status.Code, ShouldEqual, tx.ErrorRuntime)
			So(r.Status.Message, ShouldContainSubstring, "exceed amountLimit in abi")
			// todo
			// balance2 := common.Fixed{Value: s.Visitor.TokenBalance("iost", testID[2]), Decimal: s.Visitor.Decimal("iost")}
			// So(balance2.ToString(), ShouldEqual, "0")
		})

		Convey("test invalid amount limit", func() {
			trx := tx.NewTx([]*tx.Action{{
				Contract:   "token.iost",
				ActionName: "transfer",
				Data:       fmt.Sprintf(`["iost", "%v", "%v", "%v", ""]`, testID[0], testID[2], "110"),
			}}, nil, 100000, 100, 10000000, 0)
			trx.AmountLimit = append(trx.AmountLimit, &contract.Amount{Token: "iost1", Val: "100"})

			err = vm.CheckAmountLimit(s.Mvcc, trx)
			So(err.Error(), ShouldContainSubstring, "token not exists in amountLimit")
		})
	})
}

func TestTokenMemo(t *testing.T) {
	ilog.Stop()
	Convey("test of token memo", t, func() {
		s := NewSimulator()
		defer s.Clear()
		prepareContract(s)

		kp, err := account.NewKeyPair(common.Base58Decode(testID[1]), crypto.Secp256k1)
		So(err, ShouldBeNil)

		createToken(t, s, kp)
		s.SetRAM(testID[0], 10000)

		Reset(func() {
			s.Visitor.SetTokenBalanceFixed("iost", testID[0], "1000")
			s.Visitor.SetTokenBalanceFixed("iost", testID[2], "0")
			s.SetGas(kp.ID, 100000)
			s.SetRAM(testID[0], 10000)
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
		s.SetGas(kp.ID, 100000)

		tx0 := tx.NewTx([]*tx.Action{{
			Contract:   "token.iost",
			ActionName: "transfer",
			Data:       fmt.Sprintf(`["iost", "%v", "%v", "%v", ""]`, testID[0], testID[2], "10"),
		}}, nil, 550, 100, 10000000, 0)

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
		s.SetGas(kp.ID, 100000)
		s.SetRAM(kp.ID, 3000)

		cname, _, err := s.DeployContract(c, kp.ID, kp)
		So(err, ShouldBeNil)
		s.Visitor.SetContract(native.ABI("domain.iost", native.DomainABIs))
		r1, err := s.Call("domain.iost", "Link", fmt.Sprintf(`["abcde","%v"]`, cname), kp.ID, kp)
		So(err, ShouldBeNil)
		So(r1.Status.Message, ShouldEqual, "")
		r2, err := s.Call("abcde", "read", "[]", kp.ID, kp)
		So(err, ShouldBeNil)
		So(r2.Status.Message, ShouldEqual, "")
	})
}

func TestAuthority(t *testing.T) {
	ilog.SetLevel(ilog.LevelInfo)
	s := NewSimulator()
	defer s.Clear()
	Convey("test of Auth", t, func() {

		ca, err := s.Compile("auth.iost", "../../contract/account", "../../contract/account.js")
		So(err, ShouldBeNil)
		s.Visitor.SetContract(ca)

		kp := prepareAuth(t, s)
		s.SetGas(kp.ID, 100000)

		r, err := s.Call("auth.iost", "SignUp", array2json([]interface{}{"myid", kp.ID, "akey"}), kp.ID, kp)
		So(err, ShouldBeNil)
		So(r.Status.Message, ShouldEqual, "")
		So(s.Visitor.MGet("auth.iost-account", "myid"), ShouldStartWith, `s{"id":"myid",`)

		r, err = s.Call("auth.iost", "AddPermission", array2json([]interface{}{"myid", "perm1", 1}), kp.ID, kp)
		So(err, ShouldBeNil)
		So(r.Status.Message, ShouldEqual, "")
		So(s.Visitor.MGet("auth.iost-account", "myid"), ShouldContainSubstring, `"perm1":{"name":"perm1","groups":[],"items":[],"threshold":1}`)
	})

}
