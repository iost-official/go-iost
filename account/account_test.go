package account

import (
	"testing"

	. "github.com/iost-official/prototype/common"
	. "github.com/smartystreets/goconvey/convey"
)

func TestMember(t *testing.T) {
	Convey("Test of Account", t, func() {
		m, err := NewMember(nil)
		Convey("New member: ", func() {
			So(err, ShouldBeNil)
			So(len(m.Pubkey), ShouldEqual, 33)
			So(len(m.Seckey), ShouldEqual, 32)
		})

		Convey("sign and verify: ", func() {
			info := []byte("hello world")
			sig := SignInSecp256k1(Sha256(info), m.Seckey)
			So(VerifySignInSecp256k1(Sha256(info), m.Pubkey, sig), ShouldBeTrue)
		})
	})
}
