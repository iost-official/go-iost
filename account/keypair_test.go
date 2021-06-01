package account

import (
	"testing"
	"bytes"
	"fmt"

	. "github.com/iost-official/go-iost/v3/common"
	"github.com/iost-official/go-iost/v3/crypto"
	. "github.com/smartystreets/goconvey/convey"
)

func TestKeypair(t *testing.T) {
	Convey("Test of KeyPair", t, func() {
		m, err := NewKeyPair(nil, crypto.Secp256k1)
		Convey("New keypair: ", func() {
			So(err, ShouldBeNil)
			So(len(m.Pubkey), ShouldEqual, 33)
			So(len(m.Seckey), ShouldEqual, 32)
			//So(len(m.ID), ShouldEqual, len(EncodePubkey(m.Pubkey)))
		})

		Convey("sign and verify: ", func() {
			info := []byte("hello world")
			sig := crypto.Secp256k1.Sign(Sha3(info), m.Seckey)
			So(crypto.Secp256k1.Verify(Sha3(info), m.Pubkey, sig), ShouldBeTrue)

			sig2 := m.Sign(Sha3(info))
			So(bytes.Equal(sig2.Pubkey, m.Pubkey), ShouldBeTrue)

		})
		Convey("sec to pub", func() {
			m, err := NewKeyPair(Base58Decode("3BZ3HWs2nWucCCvLp7FRFv1K7RR3fAjjEQccf9EJrTv4"), crypto.Secp256k1)
			So(err, ShouldBeNil)
			fmt.Println(Base58Encode(m.Pubkey))
		})
	})
}

func TestSeckeyToPubkey(t *testing.T) {
	seckey := Base58Decode("1rANSfcRzr4HkhbUFZ7L1Zp69JZZHiDDq5v7dNSbbEqeU4jxy3fszV4HGiaLQEyqVpS1dKT9g7zCVRxBVzuiUzB")
	pubkey := crypto.Ed25519.GetPubkey(seckey)
	fmt.Println("id >", EncodePubkey(pubkey))
}

func TestPubkeyEncodeDecode(t *testing.T) {
	for i := 0; i < 10; i++ {
		seckey := crypto.Secp256k1.GenSeckey()
		pubkey := crypto.Secp256k1.GetPubkey(seckey)
		pubkeyStr := EncodePubkey(pubkey)
		pubkeyStr2 := EncodePubkey(DecodePubkey(pubkeyStr))
		if pubkeyStr != pubkeyStr2 {
			t.Fail()
		}
	}
}

