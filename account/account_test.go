package account

import (
	"testing"

	"bytes"

	"fmt"

	"strings"

	. "github.com/iost-official/go-iost/common"
	"github.com/iost-official/go-iost/crypto"
	. "github.com/smartystreets/goconvey/convey"
)

func TestMember(t *testing.T) {
	Convey("Test of KeyPair", t, func() {
		m, err := NewKeyPair(nil, crypto.Secp256k1)
		Convey("New member: ", func() {
			So(err, ShouldBeNil)
			So(len(m.Pubkey), ShouldEqual, 33)
			So(len(m.Seckey), ShouldEqual, 32)
			//So(len(m.ID), ShouldEqual, len(GetIDByPubkey(m.Pubkey)))
		})

		Convey("sign and verify: ", func() {
			info := []byte("hello world")
			sig := crypto.Secp256k1.Sign(Sha256(info), m.Seckey)
			So(crypto.Secp256k1.Verify(Sha256(info), m.Pubkey, sig), ShouldBeTrue)

			sig2 := m.Sign(Sha256(info))
			So(bytes.Equal(sig2.Pubkey, m.Pubkey), ShouldBeTrue)

		})
		Convey("sec to pub", func() {
			m, err := NewKeyPair(Base58Decode("3BZ3HWs2nWucCCvLp7FRFv1K7RR3fAjjEQccf9EJrTv4"), crypto.Secp256k1)
			So(err, ShouldBeNil)
			fmt.Println(Base58Encode(m.Pubkey))
		})
	})
}

func TestPubkeyAndID(t *testing.T) {
	for i := 0; i < 10; i++ {
		seckey := crypto.Secp256k1.GenSeckey()
		pubkey := crypto.Secp256k1.GetPubkey(seckey)
		id := GetIDByPubkey(pubkey)
		fmt.Println(`"`, id, `", "`, Base58Encode(seckey), `"`)
		pub2 := GetPubkeyByID(id)
		id2 := GetIDByPubkey(pub2)
		if !strings.HasPrefix(id, "IOST") {
			t.Failed()
		}
		if id != id2 {
			t.Failed()
		}
	}
}
