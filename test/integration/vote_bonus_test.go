package integration

import (
	"fmt"
	"testing"

	"github.com/iost-official/go-iost/core/tx"
	"github.com/iost-official/go-iost/ilog"
	. "github.com/iost-official/go-iost/verifier"
	"github.com/iost-official/go-iost/vm/database"
	. "github.com/smartystreets/goconvey/convey"
)

func Test_VoteBonus(t *testing.T) {
	ilog.Stop()
	Convey("test vote bonus", t, func() {
		s := NewSimulator()
		defer s.Clear()

		s.Head.Number = 0

		createAccountsWithResource(s)
		prepareFakeBase(t, s)
		prepareIssue(s, acc0)
		prepareNewProducerVote(t, s, acc0)
		initProducer(s)

		// deploy bonus.iost
		setNonNativeContract(s, "bonus.iost", "bonus.js", ContractPath)
		s.Call("bonus.iost", "init", `[]`, acc0.ID, acc0.KeyPair)

		s.Head.Number = 1
		for _, acc := range testAccounts[6:] {
			r, err := s.Call("vote_producer.iost", "ApplyRegister", fmt.Sprintf(`["%v", "%v", "loc", "url", "netId"]`, acc.ID, acc.KeyPair.ReadablePubkey()), acc.ID, acc.KeyPair)
			So(err, ShouldBeNil)
			So(r.Status.Code, ShouldEqual, tx.Success)
			r, err = s.Call("vote_producer.iost", "ApproveRegister", fmt.Sprintf(`["%v"]`, acc.ID), acc0.ID, acc0.KeyPair)
			So(err, ShouldBeNil)
			So(r.Status.Code, ShouldEqual, tx.Success)
			r, err = s.Call("vote_producer.iost", "LogInProducer", fmt.Sprintf(`["%v"]`, acc.ID), acc.ID, acc.KeyPair)
			So(err, ShouldBeNil)
			So(r.Status.Code, ShouldEqual, tx.Success)
		}
		for idx, acc := range testAccounts {
			r, err := s.Call("vote_producer.iost", "Vote", fmt.Sprintf(`["%v", "%v", "%v"]`, acc0.ID, acc.ID, (idx+1)*2e7), acc0.ID, acc0.KeyPair)
			So(err, ShouldBeNil)
			So(r.Status.Message, ShouldEqual, "")
			So(database.MustUnmarshal(s.Visitor.MGet("vote.iost-v_1", fmt.Sprintf(`%d`, idx))), ShouldEqual, fmt.Sprintf(`"%d"`, (idx+1)*2e7))
		}

		for idx, acc := range testAccounts {
			s.Head.Witness = acc.KeyPair.ReadablePubkey()
			for i := 0; i <= idx; i++ {
				s.Head.Number++
				r, err := s.Call("base.iost", "IssueContribute", fmt.Sprintf(`[{"parent":["%v","%v"]}]`, acc.ID, 1), acc.ID, acc.KeyPair)
				So(err, ShouldBeNil)
				So(r.Status.Message, ShouldEqual, "")
			}
			So(s.Visitor.TokenBalance("contribute", acc.ID), ShouldEqual, int64(198779440*(idx+1)))
		}
		So(database.MustUnmarshal(s.Visitor.MGet("vote.iost-u_1", acc0.ID)), ShouldEqual, `{"0":["20000000",1,"0"],"1":["40000000",1,"0"],"2":["60000000",1,"0"],"3":["80000000",1,"0"],"4":["100000000",1,"0"],"5":["120000000",1,"0"],"6":["140000000",1,"0"],"7":["160000000",1,"0"],"8":["180000000",1,"0"],"9":["200000000",1,"0"]}`)
		s.Head.Time += 5073358980
		r, err := s.Call("issue.iost", "IssueIOST", `[]`, acc0.ID, acc0.KeyPair)
		So(err, ShouldBeNil)
		So(r.Status.Message, ShouldEqual, "")
		So(s.Visitor.TokenBalance("iost", "vote_producer.iost"), ShouldEqual, int64(103300000109))
		So(s.Visitor.TokenBalance("iost", "bonus.iost"), ShouldEqual, int64(3300000109))

		for i := 0; i < 10; i++ {
			s.Visitor.SetTokenBalance("iost", testAccounts[i].ID, 100000000000)
		}
		r, err = s.Call("bonus.iost", "ExchangeIOST", fmt.Sprintf(`["%s","%s"]`, acc1.ID, "3.9755888"), acc1.ID, acc1.KeyPair)
		So(err, ShouldBeNil)
		So(r.Status.Message, ShouldEqual, "")
		So(s.Visitor.TokenBalance("contribute", acc1.ID), ShouldEqual, int64(0))
		So(s.Visitor.TokenBalance("iost", acc1.ID), ShouldEqual, int64(100198779440))
		So(s.Visitor.TokenBalance("iost", "vote_producer.iost"), ShouldEqual, int64(103498779549))

		s.Head.Time += 86400 * 1e9
		r, err = s.Call("bonus.iost", "ExchangeIOST", fmt.Sprintf(`["%s","%s"]`, acc1.ID, "0.00000001"), acc1.ID, acc1.KeyPair)
		So(err, ShouldBeNil)
		So(r.Status.Message, ShouldContainSubstring, "invalid amount: negative or greater than contribute")

		r, err = s.Call("vote_producer.iost", "CandidateWithdraw", fmt.Sprintf(`["%s"]`, acc1.ID), acc1.ID, acc1.KeyPair)
		So(err, ShouldBeNil)
		So(r.Status.Message, ShouldEqual, "")
		So(s.Visitor.TokenBalance("iost", acc1.ID), ShouldEqual, int64(100198779440+60000001))              // 60000001 = (3300000109*(2/55))/2
		So(s.Visitor.TokenBalance("iost", "vote_producer.iost"), ShouldEqual, int64(103498779549-60000001)) // half to voterBonus

		r, err = s.Call("vote_producer.iost", "CandidateWithdraw", fmt.Sprintf(`["%s"]`, acc1.ID), acc1.ID, acc1.KeyPair)
		So(err, ShouldBeNil)
		So(r.Status.Message, ShouldEqual, "")
		So(s.Visitor.TokenBalance("iost", acc1.ID), ShouldEqual, int64(100198779440+60000001)) // do not change
		So(s.Visitor.TokenBalance("iost", "vote_producer.iost"), ShouldEqual, int64(103498779549-60000001))

		r, err = s.Call("vote_producer.iost", "VoterWithdraw", fmt.Sprintf(`["%s"]`, acc0.ID), acc0.ID, acc0.KeyPair)
		So(err, ShouldBeNil)
		So(r.Status.Message, ShouldEqual, "")
		So(s.Visitor.TokenBalance("iost", acc0.ID), ShouldEqual, int64(100258779441))
		So(s.Visitor.TokenBalance("iost", "vote_producer.iost"), ShouldEqual, int64(103180000107))
	})
}
