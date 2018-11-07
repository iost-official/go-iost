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

func Test_callWithAuth(t *testing.T) {
	ilog.Stop()
	Convey("test of callWithAuth", t, func() {
		s := NewSimulator()
		defer s.Clear()
		prepareContract(s)

		kp, err := account.NewKeyPair(common.Base58Decode(testID[1]), crypto.Secp256k1)
		if err != nil {
			t.Fatal(err)
		}

		ca, err := s.Compile("Contracttransfer", "./test_data/transfer", "./test_data/transfer.js")
		if err != nil || ca == nil {
			t.Fatal(err)
		}
		s.SetContract(ca)

		Convey("test of callWithAuth", func() {
			s.Visitor.SetTokenBalanceFixed("iost", "Contracttransfer", "1000")
			r, err := s.Call("Contracttransfer", "withdraw", fmt.Sprintf(`["%v", "%v"]`, testID[0], "10"), testID[0], kp)
			s.Visitor.Commit()

			So(err, ShouldBeNil)
			So(r.Status.Code, ShouldEqual, tx.Success)
			balance := common.Fixed{Value: s.Visitor.TokenBalance("iost", "Contracttransfer"), Decimal: s.Visitor.Decimal("iost")}
			So(balance.ToString(), ShouldEqual, "990")
		})
	})
}
