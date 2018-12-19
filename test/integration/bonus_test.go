package integration

import (
	"fmt"
	"testing"

	"github.com/iost-official/go-iost/ilog"

	"github.com/iost-official/go-iost/core/tx"
	. "github.com/iost-official/go-iost/verifier"
	. "github.com/smartystreets/goconvey/convey"
)

func Test_IssueBonus(t *testing.T) {
	ilog.Stop()
	Convey("test bonus.iost", t, func() {
		s := NewSimulator()
		defer s.Clear()

		acc := testAccounts[0]
		createAccountsWithResource(s)
		prepareFakeBase(t, s)

		// deploy issue.iost
		setNonNativeContract(s, "bonus.iost", "bonus.js", ContractPath)
		s.Call("bonus.iost", "init", `[]`, acc.ID, acc.KeyPair)

		Convey("test IssueContribute", func() {
			acc2 := testAccounts[1]
			s.Head.Witness = acc2.KeyPair.ID
			s.Head.Number = 1

			r, err := s.Call("base.iost", "IssueContribute", fmt.Sprintf(`[{"parent":["%v","12345678"]}]`, acc2.ID), acc2.ID, acc2.KeyPair)
			s.Visitor.Commit()

			So(err, ShouldBeNil)
			So(r.Status.Code, ShouldEqual, tx.Success)
			So(s.Visitor.TokenBalance("contribute", acc2.ID), ShouldEqual, int64(912))
		})
	})
}

func Test_ExchangeIOST(t *testing.T) {
	ilog.Stop()
	Convey("test bonus.iost", t, func() {
		s := NewSimulator()
		defer s.Clear()

		acc := testAccounts[0]
		acc2 := testAccounts[1]
		acc3 := testAccounts[2]
		createAccountsWithResource(s)
		prepareIssue(s, acc)
		prepareFakeBase(t, s)

		// deploy bonus.iost
		setNonNativeContract(s, "bonus.iost", "bonus.js", ContractPath)
		s.Call("bonus.iost", "init", `[]`, acc.ID, acc.KeyPair)

		Convey("test ExchangeIOST", func() {
			createToken(t, s, acc)

			// set bonus pool
			s.Call("token.iost", "issue", fmt.Sprintf(`["%v", "%v", "%v"]`, "iost", "bonus.iost", "1000"), acc.ID, acc.KeyPair)

			// gain contribute
			s.Head.Witness = acc2.KeyPair.ID
			s.Head.Number = 1
			s.Call("base.iost", "IssueContribute", fmt.Sprintf(`[{"parent":["%v","%v"]}]`, acc2.ID, 1), acc2.ID, acc2.KeyPair)
			s.Visitor.Commit()

			So(s.Visitor.TokenBalance("contribute", acc2.ID), ShouldEqual, int64(900))

			s.Head.Witness = acc3.KeyPair.ID
			s.Head.Number = 2
			s.Call("base.iost", "IssueContribute", fmt.Sprintf(`[{"parent":["%v","%v"]}]`, acc3.ID, 123456789), acc3.ID, acc3.KeyPair)
			s.Visitor.Commit()

			So(s.Visitor.TokenBalance("contribute", acc3.ID), ShouldEqual, int64(1000))

			r, err := s.Call("bonus.iost", "ExchangeIOST", fmt.Sprintf(`["%v", "%v"]`, acc2.ID, "300"), acc2.ID, acc2.KeyPair)
			s.Visitor.Commit()

			So(err, ShouldBeNil)
			So(r.Status.Message, ShouldEqual, "")
			So(s.Visitor.TokenBalance("contribute", acc2.ID), ShouldEqual, int64(600))
			So(s.Visitor.TokenBalance("iost", acc2.ID), ShouldEqual, int64(15789473684))
			So(s.Visitor.TokenBalance("iost", "bonus.iost"), ShouldEqual, int64(84210526316))
		})
	})
}
