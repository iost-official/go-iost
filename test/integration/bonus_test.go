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

		createAccountsWithResource(s)
		prepareFakeBase(t, s)

		// deploy issue.iost
		setNonNativeContract(s, "bonus.iost", "bonus.js", ContractPath)
		s.Call("bonus.iost", "init", `[]`, acc0.ID, acc0.KeyPair)

		Convey("test IssueContribute", func() {
			s.Head.Witness = acc1.KeyPair.ReadablePubkey()
			s.Head.Number = 1

			r, err := s.Call("base.iost", "IssueContribute", fmt.Sprintf(`[{"parent":["%v","12345678"]}]`, acc1.ID), acc1.ID, acc1.KeyPair)
			s.Visitor.Commit()

			So(err, ShouldBeNil)
			So(r.Status.Code, ShouldEqual, tx.Success)
			So(s.Visitor.TokenBalance("contribute", acc1.ID), ShouldEqual, int64(198779440))
		})
	})
}

func Test_ExchangeIOST(t *testing.T) {
	ilog.Stop()
	Convey("test bonus.iost", t, func() {
		s := NewSimulator()
		defer s.Clear()

		s.Head.Number = 0
		createAccountsWithResource(s)
		prepareIssue(s, acc0)
		prepareNewProducerVote(t, s, acc0)
		initProducer(s)
		prepareFakeBase(t, s)

		// deploy bonus.iost
		setNonNativeContract(s, "bonus.iost", "bonus.js", ContractPath)
		s.Call("bonus.iost", "init", `[]`, acc0.ID, acc0.KeyPair)

		Convey("test ExchangeIOST", func() {
			createToken(t, s, acc0)

			// set bonus pool
			s.Call("token.iost", "issue", fmt.Sprintf(`["%v", "%v", "%v"]`, "iost", "bonus.iost", "1000"), acc0.ID, acc0.KeyPair)

			// gain contribute
			s.Head.Witness = acc1.KeyPair.ReadablePubkey()
			s.Head.Number = 1
			r, err := s.Call("base.iost", "IssueContribute", fmt.Sprintf(`[{"parent":["%v","%v"]}]`, acc1.ID, 1), acc1.ID, acc1.KeyPair)
			So(err, ShouldBeNil)
			So(r.Status.Message, ShouldEqual, "")
			s.Visitor.Commit()

			So(s.Visitor.TokenBalance("contribute", acc1.ID), ShouldEqual, int64(198779440))

			s.Head.Witness = acc2.KeyPair.ReadablePubkey()
			s.Head.Number = 2
			r, err = s.Call("base.iost", "IssueContribute", fmt.Sprintf(`[{"parent":["%v","%v"]}]`, acc2.ID, 123456789), acc2.ID, acc2.KeyPair)
			So(err, ShouldBeNil)
			So(r.Status.Message, ShouldEqual, "")
			s.Visitor.Commit()

			So(s.Visitor.TokenBalance("contribute", acc2.ID), ShouldEqual, int64(198779440))

			r, err = s.Call("bonus.iost", "ExchangeIOST", fmt.Sprintf(`["%v", "%v"]`, acc1.ID, "1.9"), acc1.ID, acc1.KeyPair)
			s.Visitor.Commit()

			So(err, ShouldBeNil)
			So(r.Status.Message, ShouldEqual, "")
			So(s.Visitor.TokenBalance("contribute", acc1.ID), ShouldEqual, int64(8779440))
			So(s.Visitor.TokenBalance("iost", acc1.ID), ShouldEqual, int64(190000000))
			So(s.Visitor.TokenBalance("iost", "bonus.iost"), ShouldEqual, int64(99810000000))
		})
	})
}
