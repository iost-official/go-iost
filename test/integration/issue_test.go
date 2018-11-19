package integration

import (
	"encoding/json"
	"testing"

	"github.com/iost-official/go-iost/core/tx"

	"github.com/iost-official/go-iost/ilog"

	"github.com/iost-official/go-iost/account"
	"github.com/iost-official/go-iost/common"
	"github.com/iost-official/go-iost/crypto"
	. "github.com/iost-official/go-iost/verifier"
	. "github.com/smartystreets/goconvey/convey"
)

func prepareIssue(s *Simulator, kp *account.KeyPair) (*tx.TxReceipt, error) {
	s.Head.Number = 0

	// deploy issue.iost
	setNonNativeContract(s, "issue.iost", "issue.js", ContractPath)
	s.Call("issue.iost", "init", `[]`, kp.ID, kp)

	witness := common.Witness{
		ID:      testID[0],
		Owner:   testID[0],
		Active:  testID[0],
		Balance: 123000,
	}
	params := []interface{}{
		testID[0],
		common.TokenInfo{
			FoundationAccount: testID[2],
			IOSTTotalSupply:   90000000000,
			IOSTDecimal:       8,
			RAMTotalSupply:    9000000000000000000,
			RAMGenesisAmount:  128,
		},
		[]interface{}{witness},
	}
	b, _ := json.Marshal(params)
	r, err := s.Call("issue.iost", "InitGenesis", string(b), kp.ID, kp)
	s.Visitor.Commit()
	return r, err
}

func Test_IOSTIssue(t *testing.T) {
	ilog.Stop()
	Convey("test issue.iost", t, func() {
		s := NewSimulator()
		defer s.Clear()

		kp, err := account.NewKeyPair(common.Base58Decode(testID[1]), crypto.Secp256k1)
		if err != nil {
			t.Fatal(err)
		}

		prepareContract(s)
		r, err := prepareIssue(s, kp)

		Convey("test init", func() {
			So(err, ShouldBeNil)
			So(r.Status.Message, ShouldEqual, "")
			So(s.Visitor.TokenBalance("iost", testID[0]), ShouldEqual, int64(123000*1e8))
			//So(s.Visitor.TokenBalance("ram", "pledge.iost"), ShouldEqual, int64(128))
		})

		Convey("test IssueIOST", func() {
			s.Head.Time += 4 * 3 * 1e9
			r, err := s.Call("issue.iost", "IssueIOST", `[]`, kp.ID, kp)
			s.Visitor.Commit()

			So(err, ShouldBeNil)
			So(r.Status.Message, ShouldEqual, "")

			So(s.Visitor.TokenBalance("iost", "bonus.iost"), ShouldEqual, int64(45654))
			So(s.Visitor.TokenBalance("iost", testID[2]), ShouldEqual, int64(92691))
		})

		/*
			Convey("test IssueRAM", func() {
				s.Head.Time += 28801 * 3 * 1e9

				r, err := s.Call("issue.iost", "IssueRAM", `[]`, kp.ID, kp)
				s.Visitor.Commit()

				So(err, ShouldBeNil)
				So(r.Status.Message, ShouldEqual, "")
				So(s.Visitor.TokenBalance("ram", "pledge.iost"), ShouldEqual, int64(128+2179*3*28801))
			})
		*/
	})
}
