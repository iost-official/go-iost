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
		s.SetContract(ca)

		Convey("test of callWithAuth", func() {
			s.Visitor.SetTokenBalanceFixed("iost", "Contracttransfer", "1000")
			r, err := s.Call("Contracttransfer", "withdraw", fmt.Sprintf(`["%v", "%v"]`, testID[0], "10"), testID[0], kp)
			s.Visitor.Commit()

			So(err, ShouldBeNil)
			So(r.Status.Message, ShouldEqual, "")
			balance := common.Fixed{Value: s.Visitor.TokenBalance("iost", "Contracttransfer"), Decimal: s.Visitor.Decimal("iost")}
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
			So(r.Receipts[0].FuncName, ShouldEqual, cname + "/receiptf")
		})

		Convey("test of event", func() {
			eve := event.GetEventCollectorInstance()
			// contract event
			sub1 := event.NewSubscription(100, []event.Event_Topic{event.Event_ContractEvent})
			eve.Subscribe(sub1)
			sub2 := event.NewSubscription(100, []event.Event_Topic{event.Event_ContractReceipt})
			eve.Subscribe(sub2)

			r, err := s.Call(cname, "event", fmt.Sprintf(`["%v"]`, "eventdata"), testID[0], kp)
			s.Visitor.Commit()

			So(err, ShouldBeNil)
			So(r.Status.Message, ShouldEqual, "")

			e := <- sub1.ReadChan()
			So(e.Data, ShouldEqual, "eventdata")
			So(e.Topic, ShouldEqual, event.Event_ContractEvent)

			// receipt event
			r, err = s.Call(cname, "receiptf", fmt.Sprintf(`["%v"]`, "receipteventdata"), testID[0], kp)
			s.Visitor.Commit()

			So(err, ShouldBeNil)
			So(r.Status.Message, ShouldEqual, "")

			e = <- sub2.ReadChan()
			So(e.Data, ShouldEqual, "receipteventdata")
			So(e.Topic, ShouldEqual, event.Event_ContractReceipt)
		})
	})
}
