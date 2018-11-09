package integration

import (
	"fmt"
	"testing"

	"github.com/iost-official/go-iost/account"
	"github.com/iost-official/go-iost/common"
	"github.com/iost-official/go-iost/crypto"
	"github.com/iost-official/go-iost/ilog"
	. "github.com/iost-official/go-iost/verifier"
	. "github.com/smartystreets/goconvey/convey"
)

func prepareBase(t *testing.T, s *Simulator, kp *account.KeyPair) {
	// deploy iost.base
	setNonNativeContract(s, "iost.base", "base.js", ContractPath)
	s.Call("iost.base", "Exec", `[]`, kp.ID, kp)

	s.Visitor.Commit()
}

func Test_Base(t *testing.T) {
	ilog.Stop()
	Convey("test Base", t, func() {
		s := NewSimulator()
		defer s.Clear()

		s.Head.Number = 0
		kp, err := account.NewKeyPair(common.Base58Decode(testID[1]), crypto.Secp256k1)
		if err != nil {
			t.Fatal(err)
		}

		prepareContract(s)
		prepareToken(t, s, kp)
		prepareProducerVote(t, s, kp)
		prepareBase(t, s, kp)
		for i := 0; i < 12; i += 2 {
			s.Call("iost.vote_producer", "InitProducer", fmt.Sprintf(`["%v"]`, testID[i]), kp.ID, kp)
		}

		re, err := s.Call("iost.base", "Exec", `[]`, kp.ID, kp)
		So(err, ShouldBeNil)
		So(re.Status.Code, ShouldEqual, 0)

	})
}
