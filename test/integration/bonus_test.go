package integration

import (
	"fmt"
	"testing"

	"github.com/iost-official/go-iost/ilog"

	"github.com/iost-official/go-iost/account"
	"github.com/iost-official/go-iost/common"
	"github.com/iost-official/go-iost/core/tx"
	"github.com/iost-official/go-iost/crypto"
	. "github.com/iost-official/go-iost/verifier"
	. "github.com/smartystreets/goconvey/convey"
)

func Test_IssueBonus(t *testing.T) {
	ilog.Stop()
	Convey("test bonus.iost", t, func() {
		s := NewSimulator()
		defer s.Clear()

		kp, err := account.NewKeyPair(common.Base58Decode(testID[1]), crypto.Secp256k1)
		if err != nil {
			t.Fatal(err)
		}

		prepareContract(s)

		// deploy issue.iost
		setNonNativeContract(s, "bonus.iost", "bonus.js", ContractPath)
		s.Call("bonus.iost", "init", `[]`, kp.ID, kp)

		Convey("test IssueContribute", func() {
			s.Head.Witness = testID[4]
			s.Head.Number = 1
			wkp, err := account.NewKeyPair(common.Base58Decode(testID[5]), crypto.Secp256k1)
			if err != nil {
				t.Fatal(err)
			}

			r, err := s.Call("bonus.iost", "IssueContribute", fmt.Sprintf(`[{"parent":["%v","12345678"]}]`, wkp.ID), wkp.ID, wkp)
			s.Visitor.Commit()

			So(err, ShouldBeNil)
			So(r.Status.Code, ShouldEqual, tx.Success)
			So(s.Visitor.TokenBalance("contribute", testID[4]), ShouldEqual, int64(912))
		})
	})
}

func Test_ExchangeIOST(t *testing.T) {
	ilog.Stop()
	Convey("test bonus.iost", t, func() {
		s := NewSimulator()
		defer s.Clear()

		kp, err := account.NewKeyPair(common.Base58Decode(testID[1]), crypto.Secp256k1)
		if err != nil {
			t.Fatal(err)
		}

		prepareContract(s)
		prepareIssue(s, kp)

		// deploy bonus.iost
		setNonNativeContract(s, "bonus.iost", "bonus.js", ContractPath)
		s.Call("bonus.iost", "init", `[]`, kp.ID, kp)

		Convey("test ExchangeIOST", func() {
			createToken(t, s, kp)

			// set bonus pool
			s.Call("token.iost", "issue", fmt.Sprintf(`["%v", "%v", "%v"]`, "iost", "bonus.iost", "1000"), kp.ID, kp)

			// gain contribute
			s.Head.Witness = testID[6]
			s.Head.Number = 1
			wkp, _ := account.NewKeyPair(common.Base58Decode(testID[7]), crypto.Secp256k1)
			s.Call("bonus.iost", "IssueContribute", fmt.Sprintf(`[{"parent":["%v","%v"]}]`, wkp.ID, 1), wkp.ID, wkp)
			s.Visitor.Commit()

			So(s.Visitor.TokenBalance("contribute", testID[6]), ShouldEqual, int64(900))

			s.Head.Witness = testID[8]
			s.Head.Number = 2
			wkp2, _ := account.NewKeyPair(common.Base58Decode(testID[9]), crypto.Secp256k1)
			s.Call("bonus.iost", "IssueContribute", fmt.Sprintf(`[{"parent":["%v","%v"]}]`, wkp2.ID, 123456789), wkp2.ID, wkp2)
			s.Visitor.Commit()

			So(s.Visitor.TokenBalance("contribute", testID[8]), ShouldEqual, int64(1000))

			r, err := s.Call("bonus.iost", "ExchangeIOST", fmt.Sprintf(`["%v", "%v"]`, testID[6], "300"), wkp.ID, wkp)
			s.Visitor.Commit()

			So(err, ShouldBeNil)
			So(r.Status.Message, ShouldEqual, "")
			So(s.Visitor.TokenBalance("contribute", testID[6]), ShouldEqual, int64(600))
			So(s.Visitor.TokenBalance("iost", testID[6]), ShouldEqual, int64(15789473684))
			So(s.Visitor.TokenBalance("iost", "bonus.iost"), ShouldEqual, int64(84210526316))
		})
	})
}
