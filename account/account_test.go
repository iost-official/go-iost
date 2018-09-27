package account

import (
	"testing"

	"bytes"

	. "github.com/iost-official/go-iost/common"
	. "github.com/smartystreets/goconvey/convey"
)

func TestMember(t *testing.T) {
	Convey("Test of Account", t, func() {
		m, err := NewAccount(nil)
		Convey("New member: ", func() {
			So(err, ShouldBeNil)
			So(len(m.Pubkey), ShouldEqual, 33)
			So(len(m.Seckey), ShouldEqual, 32)
		})

		Convey("sign and verify: ", func() {
			info := []byte("hello world")
			sig := SignInSecp256k1(Sha256(info), m.Seckey)
			So(VerifySignInSecp256k1(Sha256(info), m.Pubkey, sig), ShouldBeTrue)

			sig2, _ := Sign(Secp256k1, Sha256(info), m.Seckey)
			So(bytes.Equal(sig2.Pubkey, m.Pubkey), ShouldBeTrue)

		})
		Convey("sec to pub", func() {
			m, err := NewAccount(Base58Decode("3BZ3HWs2nWucCCvLp7FRFv1K7RR3fAjjEQccf9EJrTv4"))
			So(err, ShouldBeNil)
			So(Base58Encode(m.Pubkey), ShouldEqual, "iWgLQj3VTPN4dZnomuJMMCggv22LFw4nAkA6bmrVsmCo")
		})
	})
}
