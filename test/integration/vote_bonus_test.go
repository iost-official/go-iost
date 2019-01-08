package integration

import (
	"encoding/json"
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
			r, err := s.Call("vote_producer.iost", "ApplyRegister", fmt.Sprintf(`["%v", "%v", "loc", "url", "netId", true]`, acc.ID, acc.KeyPair.ReadablePubkey()), acc.ID, acc.KeyPair)
			So(err, ShouldBeNil)
			So(r.Status.Code, ShouldEqual, tx.Success)
			r, err = s.Call("vote_producer.iost", "ApproveRegister", fmt.Sprintf(`["%v"]`, acc.ID), acc0.ID, acc0.KeyPair)
			So(err, ShouldBeNil)
			So(r.Status.Code, ShouldEqual, tx.Success)
			r, err = s.Call("vote_producer.iost", "LogInProducer", fmt.Sprintf(`["%v"]`, acc.ID), acc.ID, acc.KeyPair)
			So(err, ShouldBeNil)
			So(r.Status.Code, ShouldEqual, tx.Success)
		}
		s.Visitor.SetTokenBalance("iost", acc2.ID, 1e17)
		for idx, acc := range testAccounts {
			voter := acc0
			if idx > 0 {
				r, err := s.Call("vote_producer.iost", "Vote", fmt.Sprintf(`["%v", "%v", "%v"]`, voter.ID, acc.ID, idx*2e7), voter.ID, voter.KeyPair)
				So(err, ShouldBeNil)
				So(r.Status.Message, ShouldEqual, "")
				So(database.MustUnmarshal(s.Visitor.MGet("vote.iost-v_1", fmt.Sprintf(`%d`, idx))), ShouldEqual, fmt.Sprintf(`"%d"`, idx*2e7))
			}
			voter = acc2
			r, err := s.Call("vote_producer.iost", "Vote", fmt.Sprintf(`["%v", "%v", "%v"]`, voter.ID, acc.ID, 2e7), voter.ID, voter.KeyPair)
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
		So(database.MustUnmarshal(s.Visitor.MGet("vote.iost-u_1", acc0.ID)), ShouldEqual, `{"1":["20000000",1,"0"],"2":["40000000",1,"0"],"3":["60000000",1,"0"],"4":["80000000",1,"0"],"5":["100000000",1,"0"],"6":["120000000",1,"0"],"7":["140000000",1,"0"],"8":["160000000",1,"0"],"9":["180000000",1,"0"]}`)
		So(database.MustUnmarshal(s.Visitor.MGet("vote.iost-u_1", acc2.ID)), ShouldEqual, `{"0":["20000000",1,"0"],"1":["20000000",1,"0"],"2":["20000000",1,"0"],"3":["20000000",1,"0"],"4":["20000000",1,"0"],"5":["20000000",1,"0"],"6":["20000000",1,"0"],"7":["20000000",1,"0"],"8":["20000000",1,"0"],"9":["20000000",1,"0"]}`)
		s.Head.Time += 5073358980
		r, err := s.Call("issue.iost", "IssueIOST", `[]`, acc0.ID, acc0.KeyPair)
		So(err, ShouldBeNil)
		So(r.Status.Message, ShouldEqual, "")
		So(s.Visitor.TokenBalance("iost", "vote_producer.iost"), ShouldEqual, int64(103300000109))
		So(s.Visitor.TokenBalance("iost", "bonus.iost"), ShouldEqual, int64(3300000109))

		for i := 0; i < 10; i++ {
			s.Visitor.SetTokenBalance("iost", testAccounts[i].ID, 0)
		}

		// 0. normal withdraw
		r, err = s.Call("bonus.iost", "ExchangeIOST", fmt.Sprintf(`["%s","%s"]`, acc1.ID, "3.9755888"), acc1.ID, acc1.KeyPair)
		So(err, ShouldBeNil)
		So(r.Status.Message, ShouldEqual, "")
		So(s.Visitor.TokenBalance("contribute", acc1.ID), ShouldEqual, int64(0))
		So(s.Visitor.TokenBalance("iost", acc1.ID), ShouldEqual, int64(397558880))
		So(s.Visitor.TokenBalance("iost", "vote_producer.iost"), ShouldEqual, int64(103300000109))

		s.Head.Time += 86400 * 1e9
		r, err = s.Call("bonus.iost", "ExchangeIOST", fmt.Sprintf(`["%s","%s"]`, acc1.ID, "0.00000001"), acc1.ID, acc1.KeyPair)
		So(err, ShouldBeNil)
		So(r.Status.Message, ShouldContainSubstring, "invalid amount: negative or greater than contribute")

		r, err = s.Call("vote_producer.iost", "CandidateWithdraw", fmt.Sprintf(`["%s"]`, acc1.ID), acc1.ID, acc1.KeyPair)
		So(err, ShouldBeNil)
		So(r.Status.Message, ShouldEqual, "")
		So(s.Visitor.TokenBalance("iost", acc1.ID), ShouldEqual, int64(397558880+60000001))                 // 60000001 = (3300000109*(2/55))/2
		So(s.Visitor.TokenBalance("iost", "vote_producer.iost"), ShouldEqual, int64(103300000109-60000001)) // half to voterBonus

		r, err = s.Call("vote_producer.iost", "CandidateWithdraw", fmt.Sprintf(`["%s"]`, acc1.ID), acc1.ID, acc1.KeyPair)
		So(err, ShouldBeNil)
		So(r.Status.Message, ShouldEqual, "")
		So(s.Visitor.TokenBalance("iost", acc1.ID), ShouldEqual, int64(397558880+60000001)) // not change
		So(s.Visitor.TokenBalance("iost", "vote_producer.iost"), ShouldEqual, int64(103300000109-60000001))

		r, err = s.Call("vote_producer.iost", "VoterWithdraw", fmt.Sprintf(`["%s"]`, acc0.ID), acc0.ID, acc0.KeyPair)
		So(err, ShouldBeNil)
		So(r.Status.Message, ShouldEqual, "")
		So(s.Visitor.TokenBalance("iost", acc0.ID), ShouldEqual, int64(30000000))
		So(s.Visitor.TokenBalance("iost", "vote_producer.iost"), ShouldEqual, int64(103210000108))

		r, err = s.Call("vote_producer.iost", "VoterWithdraw", fmt.Sprintf(`["%s"]`, acc0.ID), acc0.ID, acc0.KeyPair)
		So(err, ShouldBeNil)
		So(r.Status.Message, ShouldEqual, "")
		So(s.Visitor.TokenBalance("iost", acc0.ID), ShouldEqual, int64(30000000))
		So(s.Visitor.TokenBalance("iost", "vote_producer.iost"), ShouldEqual, int64(103210000108))

		// 1. unregistered withdraw
		r, err = s.Call("vote_producer.iost", "ForceUnregister", fmt.Sprintf(`["%v"]`, acc3.ID), acc0.ID, acc0.KeyPair)
		So(err, ShouldBeNil)
		So(r.Status.Code, ShouldEqual, tx.Success)
		r, err = s.Call("vote_producer.iost", "Unregister", fmt.Sprintf(`["%v"]`, acc3.ID), acc3.ID, acc3.KeyPair)
		So(err, ShouldBeNil)
		So(r.Status.Code, ShouldEqual, tx.Success)

		r, err = s.Call("bonus.iost", "ExchangeIOST", fmt.Sprintf(`["%s","%s"]`, acc3.ID, "7.9511776"), acc3.ID, acc3.KeyPair)
		So(err, ShouldBeNil)
		So(r.Status.Message, ShouldEqual, "")
		So(s.Visitor.TokenBalance("contribute", acc3.ID), ShouldEqual, int64(0))
		So(s.Visitor.TokenBalance("iost", acc3.ID), ShouldEqual, int64(795117760))
		So(s.Visitor.TokenBalance("iost", "vote_producer.iost"), ShouldEqual, int64(103210000108))

		s.Head.Time += 86400 * 1e9
		r, err = s.Call("bonus.iost", "ExchangeIOST", fmt.Sprintf(`["%s","%s"]`, acc3.ID, "0.00000001"), acc3.ID, acc3.KeyPair)
		So(err, ShouldBeNil)
		So(r.Status.Message, ShouldContainSubstring, "invalid amount: negative or greater than contribute")

		r, err = s.Call("vote_producer.iost", "CandidateWithdraw", fmt.Sprintf(`["%s"]`, acc3.ID), acc3.ID, acc3.KeyPair)
		So(err, ShouldBeNil)
		So(r.Status.Message, ShouldEqual, "")
		So(s.Visitor.TokenBalance("iost", acc3.ID), ShouldEqual, int64(795117760+120000003))                 // 120000003 = (3300000109*(4/55))/2
		So(s.Visitor.TokenBalance("iost", "vote_producer.iost"), ShouldEqual, int64(103210000108-120000003)) // =103090000105. half to voterBonus

		r, err = s.Call("vote_producer.iost", "CandidateWithdraw", fmt.Sprintf(`["%s"]`, acc3.ID), acc3.ID, acc3.KeyPair)
		So(err, ShouldBeNil)
		So(r.Status.Message, ShouldEqual, "")
		So(s.Visitor.TokenBalance("iost", acc3.ID), ShouldEqual, int64(795117760+120000003)) // not change
		So(s.Visitor.TokenBalance("iost", "vote_producer.iost"), ShouldEqual, int64(103210000108-120000003))

		r, err = s.Call("vote_producer.iost", "VoterWithdraw", fmt.Sprintf(`["%s"]`, acc0.ID), acc0.ID, acc0.KeyPair)
		So(err, ShouldBeNil)
		So(r.Status.Message, ShouldEqual, "")
		So(s.Visitor.TokenBalance("iost", acc0.ID), ShouldEqual, int64(30000000+90000002))                  // 120000002
		So(s.Visitor.TokenBalance("iost", "vote_producer.iost"), ShouldEqual, int64(103090000105-90000002)) // 103000000103

		r, err = s.Call("vote_producer.iost", "VoterWithdraw", fmt.Sprintf(`["%s"]`, acc0.ID), acc0.ID, acc0.KeyPair)
		So(err, ShouldBeNil)
		So(r.Status.Message, ShouldEqual, "")
		So(s.Visitor.TokenBalance("iost", acc0.ID), ShouldEqual, int64(30000000+90000002))
		So(s.Visitor.TokenBalance("iost", "vote_producer.iost"), ShouldEqual, int64(103090000105-90000002))

		// 2. re-register withdraw
		r, err = s.Call("vote_producer.iost", "ApplyRegister", fmt.Sprintf(`["%v", "%v", "loc", "url", "netId", true]`, acc3.ID, acc3.KeyPair.ID), acc3.ID, acc3.KeyPair)
		So(err, ShouldBeNil)
		So(r.Status.Message, ShouldEqual, "")
		r, err = s.Call("vote_producer.iost", "ApproveRegister", fmt.Sprintf(`["%v"]`, acc3.ID), acc0.ID, acc0.KeyPair)
		So(err, ShouldBeNil)
		So(r.Status.Message, ShouldEqual, "")
		r, err = s.Call("vote_producer.iost", "LogInProducer", fmt.Sprintf(`["%v"]`, acc3.ID), acc3.ID, acc3.KeyPair)
		So(err, ShouldBeNil)
		So(r.Status.Message, ShouldEqual, "")

		r, err = s.Call("issue.iost", "IssueIOST", `[]`, acc0.ID, acc0.KeyPair)
		So(err, ShouldBeNil)
		So(r.Status.Message, ShouldEqual, "")
		b, _ := json.Marshal(r.Receipts)
		So(string(b), ShouldContainSubstring, "1123989.09997150")
		So(s.Visitor.TokenBalance("iost", "vote_producer.iost"), ShouldEqual, int64(112501909997253))
		So(s.Visitor.TokenBalance("iost", "bonus.iost"), ShouldEqual, int64(112401017320619))

		r, err = s.Call("bonus.iost", "ExchangeIOST", fmt.Sprintf(`["%s","%s"]`, acc3.ID, "0.00000001"), acc3.ID, acc3.KeyPair)
		So(err, ShouldBeNil)
		So(r.Status.Message, ShouldContainSubstring, "invalid amount: negative or greater than contribute")

		r, err = s.Call("vote_producer.iost", "CandidateWithdraw", fmt.Sprintf(`["%s"]`, acc3.ID), acc3.ID, acc3.KeyPair)
		So(err, ShouldBeNil)
		So(r.Status.Message, ShouldEqual, "")
		So(s.Visitor.TokenBalance("iost", acc3.ID), ShouldEqual, int64(915117763+3810012542272))
		So(s.Visitor.TokenBalance("iost", "vote_producer.iost"), ShouldEqual, int64(112501909997253-3810012542272))

		r, err = s.Call("vote_producer.iost", "CandidateWithdraw", fmt.Sprintf(`["%s"]`, acc3.ID), acc3.ID, acc3.KeyPair)
		So(err, ShouldBeNil)
		So(r.Status.Message, ShouldEqual, "")
		So(s.Visitor.TokenBalance("iost", acc3.ID), ShouldEqual, int64(915117763+3810012542272)) // not change
		So(s.Visitor.TokenBalance("iost", "vote_producer.iost"), ShouldEqual, int64(112501909997253-3810012542272))

		r, err = s.Call("vote_producer.iost", "VoterWithdraw", fmt.Sprintf(`["%s"]`, acc0.ID), acc0.ID, acc0.KeyPair)
		So(err, ShouldBeNil)
		So(r.Status.Message, ShouldEqual, "")
		So(s.Visitor.TokenBalance("iost", acc0.ID), ShouldEqual, int64(2857629406706))
		So(s.Visitor.TokenBalance("iost", "vote_producer.iost"), ShouldEqual, int64(105834388048277))

		r, err = s.Call("vote_producer.iost", "VoterWithdraw", fmt.Sprintf(`["%s"]`, acc0.ID), acc0.ID, acc0.KeyPair)
		So(err, ShouldBeNil)
		So(r.Status.Message, ShouldEqual, "")
		So(s.Visitor.TokenBalance("iost", acc0.ID), ShouldEqual, int64(2857629406706))
		So(s.Visitor.TokenBalance("iost", "vote_producer.iost"), ShouldEqual, int64(105834388048277))
	})
}

func Test_PartnerBonus(t *testing.T) {
	ilog.Stop()
	Convey("test partner bonus", t, func() {
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
			r, err := s.Call("vote_producer.iost", "ApplyRegister", fmt.Sprintf(`["%v", "%v", "loc", "url", "netId", false]`, acc.ID, acc.KeyPair.ID), acc.ID, acc.KeyPair)
			So(err, ShouldBeNil)
			So(r.Status.Code, ShouldEqual, tx.Success)
			r, err = s.Call("vote_producer.iost", "ApproveRegister", fmt.Sprintf(`["%v"]`, acc.ID), acc0.ID, acc0.KeyPair)
			So(err, ShouldBeNil)
			So(r.Status.Code, ShouldEqual, tx.Success)
			r, err = s.Call("vote_producer.iost", "LogInProducer", fmt.Sprintf(`["%v"]`, acc.ID), acc.ID, acc.KeyPair)
			So(err, ShouldBeNil)
			So(r.Status.Code, ShouldEqual, tx.Success)
		}
		s.Visitor.SetTokenBalance("iost", acc2.ID, 1e17)
		for idx, acc := range testAccounts {
			voter := acc0
			if idx > 0 {
				r, err := s.Call("vote_producer.iost", "Vote", fmt.Sprintf(`["%v", "%v", "%v"]`, voter.ID, acc.ID, idx*2e7), voter.ID, voter.KeyPair)
				So(err, ShouldBeNil)
				So(r.Status.Message, ShouldEqual, "")
				So(database.MustUnmarshal(s.Visitor.MGet("vote.iost-v_1", fmt.Sprintf(`%d`, idx))), ShouldEqual, fmt.Sprintf(`"%d"`, idx*2e7))
			}
			voter = acc2
			r, err := s.Call("vote_producer.iost", "Vote", fmt.Sprintf(`["%v", "%v", "%v"]`, voter.ID, acc.ID, 2e7), voter.ID, voter.KeyPair)
			So(err, ShouldBeNil)
			So(r.Status.Message, ShouldEqual, "")
			So(database.MustUnmarshal(s.Visitor.MGet("vote.iost-v_1", fmt.Sprintf(`%d`, idx))), ShouldEqual, fmt.Sprintf(`"%d"`, (idx+1)*2e7))
		}

		So(database.MustUnmarshal(s.Visitor.MGet("vote.iost-u_1", acc0.ID)), ShouldEqual, `{"1":["20000000",1,"0"],"2":["40000000",1,"0"],"3":["60000000",1,"0"],"4":["80000000",1,"0"],"5":["100000000",1,"0"],"6":["120000000",1,"0"],"7":["140000000",1,"0"],"8":["160000000",1,"0"],"9":["180000000",1,"0"]}`)
		So(database.MustUnmarshal(s.Visitor.MGet("vote.iost-u_1", acc2.ID)), ShouldEqual, `{"0":["20000000",1,"0"],"1":["20000000",1,"0"],"2":["20000000",1,"0"],"3":["20000000",1,"0"],"4":["20000000",1,"0"],"5":["20000000",1,"0"],"6":["20000000",1,"0"],"7":["20000000",1,"0"],"8":["20000000",1,"0"],"9":["20000000",1,"0"]}`)
		s.Head.Time += 5073358980
		r, err := s.Call("issue.iost", "IssueIOST", `[]`, acc0.ID, acc0.KeyPair)
		So(err, ShouldBeNil)
		So(r.Status.Message, ShouldEqual, "")
		So(s.Visitor.TokenBalance("iost", "vote_producer.iost"), ShouldEqual, int64(103300000109))
		So(s.Visitor.TokenBalance("iost", "bonus.iost"), ShouldEqual, int64(3300000109))

		for i := 0; i < 10; i++ {
			s.Visitor.SetTokenBalance("iost", testAccounts[i].ID, 0)
		}

		// 0. normal withdraw
		r, err = s.Call("vote_producer.iost", "CandidateWithdraw", fmt.Sprintf(`["%s"]`, acc6.ID), acc6.ID, acc6.KeyPair)
		So(err, ShouldBeNil)
		So(r.Status.Message, ShouldEqual, "")
		So(s.Visitor.TokenBalance("iost", acc6.ID), ShouldEqual, int64(210000006))                           // 210000006 = (3300000109*(7/55))/2
		So(s.Visitor.TokenBalance("iost", "vote_producer.iost"), ShouldEqual, int64(103300000109-210000006)) // half to voterBonus

		r, err = s.Call("vote_producer.iost", "CandidateWithdraw", fmt.Sprintf(`["%s"]`, acc6.ID), acc6.ID, acc6.KeyPair)
		So(err, ShouldBeNil)
		So(r.Status.Message, ShouldEqual, "")
		So(s.Visitor.TokenBalance("iost", acc6.ID), ShouldEqual, int64(210000006)) // not change
		So(s.Visitor.TokenBalance("iost", "vote_producer.iost"), ShouldEqual, int64(103300000109-210000006))

		r, err = s.Call("vote_producer.iost", "VoterWithdraw", fmt.Sprintf(`["%s"]`, acc0.ID), acc0.ID, acc0.KeyPair)
		So(err, ShouldBeNil)
		So(r.Status.Message, ShouldEqual, "")
		So(s.Visitor.TokenBalance("iost", acc0.ID), ShouldEqual, int64(180000005))
		So(s.Visitor.TokenBalance("iost", "vote_producer.iost"), ShouldEqual, int64(102910000098))

		r, err = s.Call("vote_producer.iost", "VoterWithdraw", fmt.Sprintf(`["%s"]`, acc0.ID), acc0.ID, acc0.KeyPair)
		So(err, ShouldBeNil)
		So(r.Status.Message, ShouldEqual, "")
		So(s.Visitor.TokenBalance("iost", acc0.ID), ShouldEqual, int64(180000005))
		So(s.Visitor.TokenBalance("iost", "vote_producer.iost"), ShouldEqual, int64(102910000098))
	})
}
