package integration

import (
	"fmt"
	"testing"

	"github.com/iost-official/go-iost/core/tx"
	"github.com/iost-official/go-iost/ilog"
	. "github.com/iost-official/go-iost/verifier"
	. "github.com/smartystreets/goconvey/convey"
)

func prepareBase(t *testing.T, s *Simulator, acc *TestAccount) {
	// deploy base.iost
	setNonNativeContract(s, "base.iost", "base.js", ContractPath)
	r, err := s.Call("base.iost", "init", `[]`, acc.ID, acc.KeyPair)
	So(err, ShouldBeNil)
	So(r.Status.Code, ShouldEqual, tx.Success)
	s.Visitor.Commit()
}

func Test_Base(t *testing.T) {
	ilog.Stop()
	Convey("test Base", t, func() {
		s := NewSimulator()
		defer s.Clear()

		s.Head.Number = 0

		createAccountsWithResource(s)
		prepareToken(t, s, acc0)
		prepareProducerVote(t, s, acc0)
		for _, acc := range testAccounts[:6] {
			s.Call("vote_producer.iost", "InitProducer", fmt.Sprintf(`["%v", "%v"]`, acc.ID, acc.KeyPair.ID), acc.ID, acc.KeyPair)
		}

		// deploy bonus.iost
		setNonNativeContract(s, "bonus.iost", "bonus.js", ContractPath)
		s.Call("bonus.iost", "init", `[]`, acc0.ID, acc0.KeyPair)

		prepareBase(t, s, acc0)

		s.Head.Number = 200
		re, err := s.Call("base.iost", "Exec", fmt.Sprintf(`[{"parent":["%v","%v"]}]`, acc0.ID, 12345678), acc0.ID, acc0.KeyPair)
		So(err, ShouldBeNil)
		So(re.Status.Message, ShouldEqual, "")
	})
}

