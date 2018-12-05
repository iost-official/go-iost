package integration

import (
	"encoding/json"
	"fmt"
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/iost-official/go-iost/account"
	"github.com/iost-official/go-iost/common"
	"github.com/iost-official/go-iost/core/event"
	"github.com/iost-official/go-iost/core/tx"
	"github.com/iost-official/go-iost/crypto"
	"github.com/iost-official/go-iost/ilog"
	. "github.com/iost-official/go-iost/verifier"
)

func Test_callWithAuth(t *testing.T) {
	ilog.Stop()
	Convey("test of callWithAuth", t, func() {
		s := NewSimulator()
		defer s.Clear()

		kp, err := account.NewKeyPair(common.Base58Decode(testID[1]), crypto.Secp256k1)
		if err != nil {
			t.Fatal(err)
		}

		prepareContract(s)
		createToken(t, s, kp)

		ca, err := s.Compile("Contracttransfer", "./test_data/transfer", "./test_data/transfer.js")
		if err != nil || ca == nil {
			t.Fatal(err)
		}
		cname, r, err := s.DeployContract(ca, testID[0], kp)
		So(err, ShouldBeNil)
		So(r.Status.Code, ShouldEqual, tx.Success)

		Convey("test of callWithAuth", func() {
			s.Visitor.SetTokenBalanceFixed("iost", cname, "1000")
			r, err := s.Call(cname, "withdraw", fmt.Sprintf(`["%v", "%v"]`, testID[0], "10"), testID[0], kp)
			s.Visitor.Commit()

			So(err, ShouldBeNil)
			So(r.Status.Message, ShouldEqual, "")
			balance := common.Fixed{Value: s.Visitor.TokenBalance("iost", cname), Decimal: s.Visitor.Decimal("iost")}
			So(balance.ToString(), ShouldEqual, "990")
		})
	})
}

func Test_VMMethod(t *testing.T) {
	ilog.Stop()
	Convey("test of vm method", t, func() {
		s := NewSimulator()
		defer s.Clear()

		kp, err := account.NewKeyPair(common.Base58Decode(testID[1]), crypto.Secp256k1)
		if err != nil {
			t.Fatal(err)
		}

		prepareContract(s)
		createToken(t, s, kp)

		ca, err := s.Compile("", "./test_data/vmmethod", "./test_data/vmmethod")
		if err != nil || ca == nil {
			t.Fatal(err)
		}
		cname, r, err := s.DeployContract(ca, testID[0], kp)
		So(err, ShouldBeNil)
		So(r.Status.Code, ShouldEqual, tx.Success)

		Convey("test of contract name", func() {
			r, err := s.Call(cname, "contractName", "[]", testID[0], kp)
			s.Visitor.Commit()

			So(err, ShouldBeNil)
			So(r.Status.Code, ShouldEqual, tx.Success)
			So(len(r.Returns), ShouldEqual, 1)
			res, err := json.Marshal([]interface{}{cname})
			So(err, ShouldBeNil)
			So(r.Returns[0], ShouldEqual, string(res))
		})

		Convey("test of receipt", func() {
			r, err := s.Call(cname, "receiptf", fmt.Sprintf(`["%v"]`, "receiptdata"), testID[0], kp)
			s.Visitor.Commit()

			So(err, ShouldBeNil)
			So(r.Status.Message, ShouldEqual, "")
			So(len(r.Receipts), ShouldEqual, 1)
			So(r.Receipts[0].Content, ShouldEqual, "receiptdata")
			So(r.Receipts[0].FuncName, ShouldEqual, cname+"/receiptf")
		})

	})
}

func Test_VMMethod_Event(t *testing.T) {
	ilog.Stop()
	Convey("test of vm method event", t, func() {
		s := NewSimulator()
		defer s.Clear()

		kp, err := account.NewKeyPair(common.Base58Decode(testID[1]), crypto.Secp256k1)
		if err != nil {
			t.Fatal(err)
		}

		prepareContract(s)

		ca, err := s.Compile("", "./test_data/vmmethod", "./test_data/vmmethod")
		if err != nil || ca == nil {
			t.Fatal(err)
		}
		cname, r, err := s.DeployContract(ca, testID[0], kp)
		So(err, ShouldBeNil)
		So(r.Status.Code, ShouldEqual, tx.Success)

		eve := event.GetEventCollectorInstance()
		// contract event
		sub1 := event.NewSubscription(100, []event.Event_Topic{event.Event_ContractEvent})
		eve.Subscribe(sub1)
		sub2 := event.NewSubscription(100, []event.Event_Topic{event.Event_ContractReceipt})
		eve.Subscribe(sub2)

		r, err = s.Call(cname, "event", fmt.Sprintf(`["%v"]`, "eventdata"), testID[0], kp)
		s.Visitor.Commit()

		So(err, ShouldBeNil)
		So(r.Status.Message, ShouldEqual, "")

		e := <-sub1.ReadChan()
		So(e.Data, ShouldEqual, "eventdata")
		So(e.Topic, ShouldEqual, event.Event_ContractEvent)

		// receipt event
		r, err = s.Call(cname, "receiptf", fmt.Sprintf(`["%v"]`, "receipteventdata"), testID[0], kp)
		s.Visitor.Commit()

		So(err, ShouldBeNil)
		So(r.Status.Message, ShouldEqual, "")

		e = <-sub2.ReadChan()
		So(e.Data, ShouldEqual, "receipteventdata")
		So(e.Topic, ShouldEqual, event.Event_ContractReceipt)
	})
}

func Test_RamPayer(t *testing.T) {
	ilog.Stop()
	Convey("test of ram payer", t, func() {
		s := NewSimulator()
		defer s.Clear()

		kp, err := account.NewKeyPair(common.Base58Decode(testID[1]), crypto.Secp256k1)
		if err != nil {
			t.Fatal(err)
		}

		prepareContract(s)
		createToken(t, s, kp)

		ca, err := s.Compile("", "./test_data/vmmethod", "./test_data/vmmethod")
		if err != nil || ca == nil {
			t.Fatal(err)
		}
		cname, r, err := s.DeployContract(ca, testID[0], kp)
		So(err, ShouldBeNil)
		So(r.Status.Code, ShouldEqual, tx.Success)

		Convey("test of put and get", func() {
			ram := s.GetRAM(testID[0])
			r, err := s.Call(cname, "putwithpayer", fmt.Sprintf(`["k", "v", "%v"]`, testID[0]), testID[0], kp)
			s.Visitor.Commit()
			So(s.GetRAM(testID[0]), ShouldEqual, ram-111)
			So(err, ShouldBeNil)
			So(r.Status.Code, ShouldEqual, tx.Success)

			r, err = s.Call(cname, "get", fmt.Sprintf(`["k"]`), testID[0], kp)
			So(err, ShouldBeNil)
			So(r.Status.Code, ShouldEqual, tx.Success)
			So(len(r.Returns), ShouldEqual, 1)
			So(r.Returns[0], ShouldEqual, "[\"v\"]")
		})

		Convey("test of map put and get", func() {
			ram := s.GetRAM(testID[0])
			r, err := s.Call(cname, "mapputwithpayer", fmt.Sprintf(`["k", "f", "v", "%v"]`, testID[0]), testID[0], kp)
			s.Visitor.Commit()
			So(err, ShouldBeNil)
			So(r.Status.Code, ShouldEqual, tx.Success)
			So(s.GetRAM(testID[0]), ShouldEqual, ram-113)

			r, err = s.Call(cname, "mapget", fmt.Sprintf(`["k", "f"]`), testID[0], kp)
			So(err, ShouldBeNil)
			So(r.Status.Code, ShouldEqual, tx.Success)
			So(len(r.Returns), ShouldEqual, 1)
			So(r.Returns[0], ShouldEqual, "[\"v\"]")
		})

		Convey("test of map put and get change payer", func() {
			kp2, err := account.NewKeyPair(common.Base58Decode(testID[3]), crypto.Secp256k1)
			if err != nil {
				t.Fatal(err)
			}

			ram := s.GetRAM(testID[0])
			r, err := s.Call(cname, "mapputwithpayer", fmt.Sprintf(`["k", "f", "vv", "%v"]`, testID[0]), testID[0], kp)
			s.Visitor.Commit()
			So(err, ShouldBeNil)
			So(r.Status.Code, ShouldEqual, tx.Success)
			So(s.GetRAM(testID[0]), ShouldEqual, ram-114)

			ram = s.GetRAM(testID[0])
			ram1 := s.GetRAM(testID[2])
			r, err = s.Call(cname, "mapputwithpayer", fmt.Sprintf(`["k", "f", "vvv", "%v"]`, testID[2]), testID[2], kp2)
			s.Visitor.Commit()
			So(err, ShouldBeNil)
			So(r.Status.Code, ShouldEqual, tx.Success)
			So(s.GetRAM(testID[0]), ShouldEqual, ram+114)
			So(s.GetRAM(testID[2]), ShouldEqual, ram1-115)

			ram1 = s.GetRAM(testID[2])
			r, err = s.Call(cname, "mapputwithpayer", fmt.Sprintf(`["k", "f", "v", "%v"]`, testID[2]), testID[2], kp2)
			s.Visitor.Commit()
			So(err, ShouldBeNil)
			So(r.Status.Code, ShouldEqual, tx.Success)
			So(s.GetRAM(testID[2]), ShouldEqual, ram1+2)

			ram1 = s.GetRAM(testID[2])
			r, err = s.Call(cname, "mapputwithpayer", fmt.Sprintf(`["k", "f", "vvvvv", "%v"]`, testID[2]), testID[2], kp2)
			s.Visitor.Commit()
			So(err, ShouldBeNil)
			So(r.Status.Code, ShouldEqual, tx.Success)
			So(s.GetRAM(testID[2]), ShouldEqual, ram1-4)
		})

		Convey("test nested call check payer", func() {
			ram0 := s.GetRAM(testID[0])
			kp4, err := account.NewKeyPair(common.Base58Decode(testID[5]), crypto.Secp256k1)
			if err != nil {
				t.Fatal(err)
			}
			ca, err := s.Compile("", "./test_data/nest0", "./test_data/nest0")
			if err != nil || ca == nil {
				t.Fatal(err)
			}
			cname0, r, err := s.DeployContract(ca, testID[0], kp)
			So(err, ShouldBeNil)
			So(r.Status.Code, ShouldEqual, tx.Success)

			ca, err = s.Compile("", "./test_data/nest1", "./test_data/nest1")
			if err != nil || ca == nil {
				t.Fatal(err)
			}
			cname1, r, err := s.DeployContract(ca, testID[0], kp)
			So(err, ShouldBeNil)
			So(r.Status.Code, ShouldEqual, tx.Success)

			So(s.GetRAM(testID[0]), ShouldEqual, ram0-2465)

			ram0 = s.GetRAM(testID[0])
			ram4 := s.GetRAM(testID[4])
			ram6 := s.GetRAM(testID[6])
			s.Visitor.SetTokenBalanceFixed("iost", testID[4], "100")
			r, err = s.Call(cname0, "call", fmt.Sprintf(`["%v", "test", "%v"]`, cname1,
				fmt.Sprintf(`[\"%v\", \"%v\"]`, testID[4], testID[6])), testID[4], kp4)
			So(err, ShouldBeNil)
			So(r.Status.Message, ShouldEqual, "")
			So(r.Status.Code, ShouldEqual, tx.Success)

			So(s.GetRAM(testID[6]), ShouldEqual, ram6)
			So(s.GetRAM(testID[4]), ShouldEqual, ram4-139)
			So(s.GetRAM(testID[0]), ShouldEqual, ram0-6)
		})
	})
}

func Test_StackHeight(t *testing.T) {
	ilog.Stop()
	Convey("test of stack height", t, func() {
		s := NewSimulator()
		defer s.Clear()

		kp, err := account.NewKeyPair(common.Base58Decode(testID[1]), crypto.Secp256k1)
		if err != nil {
			t.Fatal(err)
		}

		prepareContract(s)
		createToken(t, s, kp)

		ca, err := s.Compile("", "./test_data/nest0", "./test_data/nest0")
		if err != nil || ca == nil {
			t.Fatal(err)
		}
		cname0, r, err := s.DeployContract(ca, testID[0], kp)
		So(err, ShouldBeNil)
		So(r.Status.Code, ShouldEqual, tx.Success)

		ca, err = s.Compile("", "./test_data/nest1", "./test_data/nest1")
		if err != nil || ca == nil {
			t.Fatal(err)
		}
		cname1, r, err := s.DeployContract(ca, testID[0], kp)
		So(err, ShouldBeNil)
		So(r.Status.Code, ShouldEqual, tx.Success)

		Convey("test of out of stack height", func() {
			r, err := s.Call(cname0, "sh0", fmt.Sprintf(`["%v"]`, cname1), testID[0], kp)
			s.Visitor.Commit()
			So(err, ShouldBeNil)
			So(r.Status.Message, ShouldContainSubstring, "stack height exceed.")
		})
	})
}

func Test_Validate(t *testing.T) {
	ilog.SetLevel(ilog.LevelInfo)
	Convey("test validate", t, func() {
		s := NewSimulator()
		defer s.Clear()
		kp := prepareAuth(t, s)
		s.SetAccount(account.NewInitAccount(kp.ID, kp.ID, kp.ID))
		s.SetGas(kp.ID, 1000000)
		s.SetRAM(kp.ID, 300)

		c, err := s.Compile("validate", "test_data/validate", "test_data/validate")
		So(err, ShouldBeNil)
		So(len(c.Encode()), ShouldEqual, 133)
		_, r, err := s.DeployContract(c, kp.ID, kp)
		s.Visitor.Commit()
		So(err.Error(), ShouldContainSubstring, "abi not defined in source code: c")
		So(r.Status.Message, ShouldEqual, "validate code error: , result: Error: abi not defined in source code: c")
	})
}
