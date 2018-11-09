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

	// deploy iost.issue
	setNonNativeContract(s, "iost.issue", "issue.js", ContractPath)
	s.Call("iost.issue", "init", `[]`, kp.ID, kp)

	witness := common.Witness{
		ID:      testID[0],
		Owner:   testID[0],
		Active:  testID[0],
		Balance: 123000,
	}
	genesisConfig := make(map[string]interface{})
	genesisConfig["iostWitnessInfo"] = []interface{}{witness}
	genesisConfig["iostDecimal"] = 8
	genesisConfig["foundationAcc"] = testID[2]
	genesisConfig["ramGenesisAmount"] = 128
	params := []interface{}{
		testID[0],
		genesisConfig,
	}
	b, _ := json.Marshal(params)
	r, err := s.Call("iost.issue", "InitGenesis", string(b), kp.ID, kp)
	s.Visitor.Commit()
	return r, err
}

func Test_IOSTIssue(t *testing.T) {
	ilog.Stop()
	Convey("test iost.issue", t, func() {
		s := NewSimulator()
		defer s.Clear()
		s.Visitor.SetBalance(testID[0], 10000000)

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
			So(s.Visitor.TokenBalance("ram", "iost.pledge"), ShouldEqual, int64(128))
		})

		Convey("test IssueIOST", func() {
			s.Head.Time += 4 * 3 * 1e9
			r, err := s.Call("iost.issue", "IssueIOST", `[]`, kp.ID, kp)
			s.Visitor.Commit()

			So(err, ShouldBeNil)
			So(r.Status.Message, ShouldEqual, "")

			So(s.Visitor.TokenBalance("iost", "iost.bonus"), ShouldEqual, int64(45654))
			So(s.Visitor.TokenBalance("iost", testID[2]), ShouldEqual, int64(92691))
		})

		Convey("test IssueRAM", func() {
			s.Head.Time += 28801 * 3 * 1e9

			r, err := s.Call("iost.issue", "IssueRAM", `[]`, kp.ID, kp)
			s.Visitor.Commit()

			So(err, ShouldBeNil)
			So(r.Status.Message, ShouldEqual, "")
			So(s.Visitor.TokenBalance("ram", "iost.pledge"), ShouldEqual, int64(128+2179*3*28801))
		})
	})
}
