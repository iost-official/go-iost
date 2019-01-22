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
			r, err := s.Call("vote_producer.iost", "applyRegister", fmt.Sprintf(`["%v", "%v", "loc", "url", "netId", true]`, acc.ID, acc.KeyPair.ReadablePubkey()), acc.ID, acc.KeyPair)
			So(err, ShouldBeNil)
			So(r.Status.Code, ShouldEqual, tx.Success)
			r, err = s.Call("vote_producer.iost", "approveRegister", fmt.Sprintf(`["%v"]`, acc.ID), acc0.ID, acc0.KeyPair)
			So(err, ShouldBeNil)
			So(r.Status.Code, ShouldEqual, tx.Success)
			r, err = s.Call("vote_producer.iost", "logInProducer", fmt.Sprintf(`["%v"]`, acc.ID), acc.ID, acc.KeyPair)
			So(err, ShouldBeNil)
			So(r.Status.Code, ShouldEqual, tx.Success)
		}
		s.Visitor.SetTokenBalance("iost", acc2.ID, 1e17)
		for idx, acc := range testAccounts {
			voter := acc0
			if idx > 0 {
				r, err := s.Call("vote_producer.iost", "vote", fmt.Sprintf(`["%v", "%v", "%v"]`, voter.ID, acc.ID, idx*2e7), voter.ID, voter.KeyPair)
				So(err, ShouldBeNil)
				So(r.Status.Message, ShouldEqual, "")
				So(database.MustUnmarshal(s.Visitor.MGet("vote.iost-v_1", acc.ID)), ShouldEqual, fmt.Sprintf(`{"votes":"%d","deleted":0,"clearTime":-1}`, idx*2e7))
			}
			voter = acc2
			r, err := s.Call("vote_producer.iost", "vote", fmt.Sprintf(`["%v", "%v", "%v"]`, voter.ID, acc.ID, 2e7), voter.ID, voter.KeyPair)
			So(err, ShouldBeNil)
			So(r.Status.Message, ShouldEqual, "")
			So(database.MustUnmarshal(s.Visitor.MGet("vote.iost-v_1", acc.ID)), ShouldEqual, fmt.Sprintf(`{"votes":"%d","deleted":0,"clearTime":-1}`, (idx+1)*2e7))
		}

		for idx, acc := range testAccounts {
			s.Head.Witness = acc.KeyPair.ReadablePubkey()
			for i := 0; i <= idx; i++ {
				s.Head.Number++
				r, err := s.Call("base.iost", "issueContribute", fmt.Sprintf(`[{"parent":["%v","%v"]}]`, acc.ID, 1), acc.ID, acc.KeyPair)
				So(err, ShouldBeNil)
				So(r.Status.Message, ShouldEqual, "")
			}
			So(s.Visitor.TokenBalance("contribute", acc.ID), ShouldEqual, int64(198779440*(idx+1)))
		}
		So(database.MustUnmarshal(s.Visitor.MGet("vote.iost-u_1", acc0.ID)), ShouldEqual, `{"user_1":["20000000",1,"0"],"user_2":["40000000",1,"0"],"user_3":["60000000",1,"0"],"user_4":["80000000",1,"0"],"user_5":["100000000",1,"0"],"user_6":["120000000",1,"0"],"user_7":["140000000",1,"0"],"user_8":["160000000",1,"0"],"user_9":["180000000",1,"0"]}`)
		So(database.MustUnmarshal(s.Visitor.MGet("vote.iost-u_1", acc2.ID)), ShouldEqual, `{"user_0":["20000000",1,"0"],"user_1":["20000000",1,"0"],"user_2":["20000000",1,"0"],"user_3":["20000000",1,"0"],"user_4":["20000000",1,"0"],"user_5":["20000000",1,"0"],"user_6":["20000000",1,"0"],"user_7":["20000000",1,"0"],"user_8":["20000000",1,"0"],"user_9":["20000000",1,"0"]}`)
		s.Head.Time += 5073358980
		r, err := s.Call("issue.iost", "issueIOST", `[]`, acc0.ID, acc0.KeyPair)
		So(err, ShouldBeNil)
		So(r.Status.Message, ShouldEqual, "")
		So(s.Visitor.TokenBalance("iost", "vote_producer.iost"), ShouldEqual, int64(103333333410))
		So(s.Visitor.TokenBalance("iost", "bonus.iost"), ShouldEqual, int64(3333333410))

		for i := 0; i < 10; i++ {
			s.Visitor.SetTokenBalance("iost", testAccounts[i].ID, 0)
		}

		// 0. normal withdraw
		r, err = s.Call("bonus.iost", "exchangeIOST", fmt.Sprintf(`["%s","%s"]`, acc1.ID, "3.9755888"), acc1.ID, acc1.KeyPair)
		So(err, ShouldBeNil)
		So(r.Status.Message, ShouldEqual, "")
		So(s.Visitor.TokenBalance("contribute", acc1.ID), ShouldEqual, int64(0))
		So(s.Visitor.TokenBalance("iost", acc1.ID), ShouldEqual, int64(397558880))
		So(s.Visitor.TokenBalance("iost", "vote_producer.iost"), ShouldEqual, int64(103333333410))

		s.Head.Time += 86400 * 1e9
		r, err = s.Call("bonus.iost", "exchangeIOST", fmt.Sprintf(`["%s","%s"]`, acc1.ID, "0.00000001"), acc1.ID, acc1.KeyPair)
		So(err, ShouldBeNil)
		So(r.Status.Message, ShouldContainSubstring, "invalid amount: negative or greater than contribute")

		r, err = s.Call("vote_producer.iost", "candidateWithdraw", fmt.Sprintf(`["%s"]`, acc1.ID), acc1.ID, acc1.KeyPair)
		So(err, ShouldBeNil)
		So(r.Status.Message, ShouldEqual, "")
		So(s.Visitor.TokenBalance("iost", acc1.ID), ShouldEqual, int64(397558880+60606062))                 // 60606062 = (3333333410*(2/55))/2
		So(s.Visitor.TokenBalance("iost", "vote_producer.iost"), ShouldEqual, int64(103333333410-60606062)) // half to voterBonus

		r, err = s.Call("vote_producer.iost", "candidateWithdraw", fmt.Sprintf(`["%s"]`, acc1.ID), acc1.ID, acc1.KeyPair)
		So(err, ShouldBeNil)
		So(r.Status.Message, ShouldEqual, "")
		So(s.Visitor.TokenBalance("iost", acc1.ID), ShouldEqual, int64(397558880+60606062)) // not change
		So(s.Visitor.TokenBalance("iost", "vote_producer.iost"), ShouldEqual, int64(103333333410-60606062))

		r, err = s.Call("vote_producer.iost", "voterWithdraw", fmt.Sprintf(`["%s"]`, acc0.ID), acc0.ID, acc0.KeyPair)
		So(err, ShouldBeNil)
		So(r.Status.Message, ShouldEqual, "")
		So(s.Visitor.TokenBalance("iost", acc0.ID), ShouldEqual, int64(30303031))
		So(s.Visitor.TokenBalance("iost", "vote_producer.iost"), ShouldEqual, int64(103242424317))

		r, err = s.Call("vote_producer.iost", "voterWithdraw", fmt.Sprintf(`["%s"]`, acc0.ID), acc0.ID, acc0.KeyPair)
		So(err, ShouldBeNil)
		So(r.Status.Message, ShouldEqual, "")
		So(s.Visitor.TokenBalance("iost", acc0.ID), ShouldEqual, int64(30303031))
		So(s.Visitor.TokenBalance("iost", "vote_producer.iost"), ShouldEqual, int64(103242424317))

		// 1. unregistered withdraw
		r, err = s.Call("vote_producer.iost", "forceUnregister", fmt.Sprintf(`["%v"]`, acc3.ID), acc0.ID, acc0.KeyPair)
		So(err, ShouldBeNil)
		So(r.Status.Message, ShouldEqual, "")
		r, err = s.Call("vote_producer.iost", "unregister", fmt.Sprintf(`["%v"]`, acc3.ID), acc3.ID, acc3.KeyPair)
		So(err, ShouldBeNil)
		So(r.Status.Message, ShouldContainSubstring, "producer in pending list or in current list, can't unregister")

		r, err = s.Call("bonus.iost", "exchangeIOST", fmt.Sprintf(`["%s","%s"]`, acc3.ID, "7.9511776"), acc3.ID, acc3.KeyPair)
		So(err, ShouldBeNil)
		So(r.Status.Message, ShouldEqual, "")
		So(s.Visitor.TokenBalance("contribute", acc3.ID), ShouldEqual, int64(0))
		So(s.Visitor.TokenBalance("iost", acc3.ID), ShouldEqual, int64(795117760))
		So(s.Visitor.TokenBalance("iost", "vote_producer.iost"), ShouldEqual, int64(103242424317))

		s.Head.Time += 86400 * 1e9
		r, err = s.Call("bonus.iost", "exchangeIOST", fmt.Sprintf(`["%s","%s"]`, acc3.ID, "0.00000001"), acc3.ID, acc3.KeyPair)
		So(err, ShouldBeNil)
		So(r.Status.Message, ShouldContainSubstring, "invalid amount: negative or greater than contribute")

		r, err = s.Call("vote_producer.iost", "candidateWithdraw", fmt.Sprintf(`["%s"]`, acc3.ID), acc3.ID, acc3.KeyPair)
		So(err, ShouldBeNil)
		So(r.Status.Message, ShouldEqual, "")
		So(s.Visitor.TokenBalance("iost", acc3.ID), ShouldEqual, int64(795117760+121212124))                 // 121212124 = (3333333410*(4/55))/2
		So(s.Visitor.TokenBalance("iost", "vote_producer.iost"), ShouldEqual, int64(103242424317-121212124)) // =103121212193. half to voterBonus

		r, err = s.Call("vote_producer.iost", "candidateWithdraw", fmt.Sprintf(`["%s"]`, acc3.ID), acc3.ID, acc3.KeyPair)
		So(err, ShouldBeNil)
		So(r.Status.Message, ShouldEqual, "")
		So(s.Visitor.TokenBalance("iost", acc3.ID), ShouldEqual, int64(795117760+121212124)) // not change
		So(s.Visitor.TokenBalance("iost", "vote_producer.iost"), ShouldEqual, int64(103242424317-121212124))

		r, err = s.Call("vote_producer.iost", "voterWithdraw", fmt.Sprintf(`["%s"]`, acc0.ID), acc0.ID, acc0.KeyPair)
		So(err, ShouldBeNil)
		So(r.Status.Message, ShouldEqual, "")
		So(s.Visitor.TokenBalance("iost", acc0.ID), ShouldEqual, int64(30303031+90909093))                  // 121212124
		So(s.Visitor.TokenBalance("iost", "vote_producer.iost"), ShouldEqual, int64(103121212193-90909093)) // 103030303100

		r, err = s.Call("vote_producer.iost", "voterWithdraw", fmt.Sprintf(`["%s"]`, acc0.ID), acc0.ID, acc0.KeyPair)
		So(err, ShouldBeNil)
		So(r.Status.Message, ShouldEqual, "")
		So(s.Visitor.TokenBalance("iost", acc0.ID), ShouldEqual, int64(30303031+90909093))
		So(s.Visitor.TokenBalance("iost", "vote_producer.iost"), ShouldEqual, int64(103121212193-90909093))

		// 2. re-register withdraw
		r, err = s.Call("vote_producer.iost", "approveRegister", fmt.Sprintf(`["%v"]`, acc3.ID), acc0.ID, acc0.KeyPair)
		So(err, ShouldBeNil)
		So(r.Status.Message, ShouldEqual, "")
		r, err = s.Call("vote_producer.iost", "logInProducer", fmt.Sprintf(`["%v"]`, acc3.ID), acc3.ID, acc3.KeyPair)
		So(err, ShouldBeNil)
		So(r.Status.Message, ShouldEqual, "")

		r, err = s.Call("issue.iost", "issueIOST", `[]`, acc0.ID, acc0.KeyPair)
		So(err, ShouldBeNil)
		So(r.Status.Message, ShouldEqual, "")
		b, _ := json.Marshal(r.Receipts)
		So(string(b), ShouldContainSubstring, "1135342.51387031")
		So(s.Visitor.TokenBalance("iost", "vote_producer.iost"), ShouldEqual, int64(113637281690131))
		So(s.Visitor.TokenBalance("iost", "bonus.iost"), ShouldEqual, int64(113536392043801))

		r, err = s.Call("bonus.iost", "exchangeIOST", fmt.Sprintf(`["%s","%s"]`, acc3.ID, "0.00000001"), acc3.ID, acc3.KeyPair)
		So(err, ShouldBeNil)
		So(r.Status.Message, ShouldContainSubstring, "invalid amount: negative or greater than contribute")

		r, err = s.Call("vote_producer.iost", "candidateWithdraw", fmt.Sprintf(`["%s"]`, acc3.ID), acc3.ID, acc3.KeyPair)
		So(err, ShouldBeNil)
		So(r.Status.Message, ShouldEqual, "")
		So(s.Visitor.TokenBalance("iost", acc3.ID), ShouldEqual, int64(916329884+3848497478961))
		So(s.Visitor.TokenBalance("iost", "vote_producer.iost"), ShouldEqual, int64(113637281690131-3848497478961))

		r, err = s.Call("vote_producer.iost", "candidateWithdraw", fmt.Sprintf(`["%s"]`, acc3.ID), acc3.ID, acc3.KeyPair)
		So(err, ShouldBeNil)
		So(r.Status.Message, ShouldEqual, "")
		So(s.Visitor.TokenBalance("iost", acc3.ID), ShouldEqual, int64(916329884+3848497478961)) // not change
		So(s.Visitor.TokenBalance("iost", "vote_producer.iost"), ShouldEqual, int64(113637281690131-3848497478961))

		r, err = s.Call("vote_producer.iost", "voterWithdraw", fmt.Sprintf(`["%s"]`, acc0.ID), acc0.ID, acc0.KeyPair)
		So(err, ShouldBeNil)
		So(r.Status.Message, ShouldEqual, "")
		So(s.Visitor.TokenBalance("iost", acc0.ID), ShouldEqual, int64(2886494321345))
		So(s.Visitor.TokenBalance("iost", "vote_producer.iost"), ShouldEqual, int64(106902411101949))

		r, err = s.Call("vote_producer.iost", "voterWithdraw", fmt.Sprintf(`["%s"]`, acc0.ID), acc0.ID, acc0.KeyPair)
		So(err, ShouldBeNil)
		So(r.Status.Message, ShouldEqual, "")
		So(s.Visitor.TokenBalance("iost", acc0.ID), ShouldEqual, int64(2886494321345))
		So(s.Visitor.TokenBalance("iost", "vote_producer.iost"), ShouldEqual, int64(106902411101949))
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
			r, err := s.Call("vote_producer.iost", "applyRegister", fmt.Sprintf(`["%v", "%v", "loc", "url", "netId", false]`, acc.ID, acc.KeyPair.ReadablePubkey()), acc.ID, acc.KeyPair)
			So(err, ShouldBeNil)
			So(r.Status.Code, ShouldEqual, tx.Success)
			r, err = s.Call("vote_producer.iost", "approveRegister", fmt.Sprintf(`["%v"]`, acc.ID), acc0.ID, acc0.KeyPair)
			So(err, ShouldBeNil)
			So(r.Status.Code, ShouldEqual, tx.Success)
			r, err = s.Call("vote_producer.iost", "logInProducer", fmt.Sprintf(`["%v"]`, acc.ID), acc.ID, acc.KeyPair)
			So(err, ShouldBeNil)
			So(r.Status.Code, ShouldEqual, tx.Success)
		}
		s.Visitor.SetTokenBalance("iost", acc2.ID, 1e17)
		for idx, acc := range testAccounts {
			voter := acc0
			if idx > 0 {
				r, err := s.Call("vote_producer.iost", "vote", fmt.Sprintf(`["%v", "%v", "%v"]`, voter.ID, acc.ID, idx*2e7), voter.ID, voter.KeyPair)
				So(err, ShouldBeNil)
				So(r.Status.Message, ShouldEqual, "")
				So(database.MustUnmarshal(s.Visitor.MGet("vote.iost-v_1", acc.ID)), ShouldEqual, fmt.Sprintf(`{"votes":"%d","deleted":0,"clearTime":-1}`, idx*2e7))
			}
			voter = acc2
			r, err := s.Call("vote_producer.iost", "vote", fmt.Sprintf(`["%v", "%v", "%v"]`, voter.ID, acc.ID, 2e7), voter.ID, voter.KeyPair)
			So(err, ShouldBeNil)
			So(r.Status.Message, ShouldEqual, "")
			So(database.MustUnmarshal(s.Visitor.MGet("vote.iost-v_1", acc.ID)), ShouldEqual, fmt.Sprintf(`{"votes":"%d","deleted":0,"clearTime":-1}`, (idx+1)*2e7))
		}

		So(database.MustUnmarshal(s.Visitor.MGet("vote.iost-u_1", acc0.ID)), ShouldEqual, `{"user_1":["20000000",1,"0"],"user_2":["40000000",1,"0"],"user_3":["60000000",1,"0"],"user_4":["80000000",1,"0"],"user_5":["100000000",1,"0"],"user_6":["120000000",1,"0"],"user_7":["140000000",1,"0"],"user_8":["160000000",1,"0"],"user_9":["180000000",1,"0"]}`)
		So(database.MustUnmarshal(s.Visitor.MGet("vote.iost-u_1", acc2.ID)), ShouldEqual, `{"user_0":["20000000",1,"0"],"user_1":["20000000",1,"0"],"user_2":["20000000",1,"0"],"user_3":["20000000",1,"0"],"user_4":["20000000",1,"0"],"user_5":["20000000",1,"0"],"user_6":["20000000",1,"0"],"user_7":["20000000",1,"0"],"user_8":["20000000",1,"0"],"user_9":["20000000",1,"0"]}`)
		s.Head.Time += 5073358980
		r, err := s.Call("issue.iost", "issueIOST", `[]`, acc0.ID, acc0.KeyPair)
		So(err, ShouldBeNil)
		So(r.Status.Message, ShouldEqual, "")
		So(s.Visitor.TokenBalance("iost", "vote_producer.iost"), ShouldEqual, int64(103333333410))
		So(s.Visitor.TokenBalance("iost", "bonus.iost"), ShouldEqual, int64(3333333410))

		for i := 0; i < 10; i++ {
			s.Visitor.SetTokenBalance("iost", testAccounts[i].ID, 0)
		}

		// 0. normal withdraw
		r, err = s.Call("vote_producer.iost", "candidateWithdraw", fmt.Sprintf(`["%s"]`, acc6.ID), acc6.ID, acc6.KeyPair)
		So(err, ShouldBeNil)
		So(r.Status.Message, ShouldEqual, "")
		So(s.Visitor.TokenBalance("iost", acc6.ID), ShouldEqual, int64(212121217))                           // 212121217 = (3333333410*(7/55))/2
		So(s.Visitor.TokenBalance("iost", "vote_producer.iost"), ShouldEqual, int64(103333333410-212121217)) // half to voterBonus

		r, err = s.Call("vote_producer.iost", "candidateWithdraw", fmt.Sprintf(`["%s"]`, acc6.ID), acc6.ID, acc6.KeyPair)
		So(err, ShouldBeNil)
		So(r.Status.Message, ShouldEqual, "")
		So(s.Visitor.TokenBalance("iost", acc6.ID), ShouldEqual, int64(212121217)) // not change
		So(s.Visitor.TokenBalance("iost", "vote_producer.iost"), ShouldEqual, int64(103333333410-212121217))

		r, err = s.Call("vote_producer.iost", "voterWithdraw", fmt.Sprintf(`["%s"]`, acc0.ID), acc0.ID, acc0.KeyPair)
		So(err, ShouldBeNil)
		So(r.Status.Message, ShouldEqual, "")
		So(s.Visitor.TokenBalance("iost", acc0.ID), ShouldEqual, int64(181818186))
		So(s.Visitor.TokenBalance("iost", "vote_producer.iost"), ShouldEqual, int64(102939394007))

		r, err = s.Call("vote_producer.iost", "voterWithdraw", fmt.Sprintf(`["%s"]`, acc0.ID), acc0.ID, acc0.KeyPair)
		So(err, ShouldBeNil)
		So(r.Status.Message, ShouldEqual, "")
		So(s.Visitor.TokenBalance("iost", acc0.ID), ShouldEqual, int64(181818186))
		So(s.Visitor.TokenBalance("iost", "vote_producer.iost"), ShouldEqual, int64(102939394007))
	})
}
