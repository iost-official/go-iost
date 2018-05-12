package common

import (
	"testing"

	"bytes"

	. "github.com/smartystreets/goconvey/convey"
)

func TestSignature(t *testing.T) {
	Convey("Test of Signature", t, func() {
		Convey("Encode and decode", func() {
			//ac, _ := account.NewAccount(nil)
			//sig, err := Sign(Secp256k1, Sha256([]byte("hello")), ac.Seckey)
			//So(err, ShouldBeNil)
			//So(bytes.Equal(sig.Pubkey, ac.Pubkey), ShouldBeTrue)
			info := Sha256([]byte("hello"))
			seckey := Sha256([]byte("seckey"))
			pubkey := CalcPubkeyInSecp256k1(seckey)
			sig, _ := Sign(Secp256k1, info, seckey)
			So(bytes.Equal(sig.Pubkey, pubkey), ShouldBeTrue)

			ans := VerifySignature(info, sig)
			So(ans, ShouldBeTrue)

			bsig := sig.Encode()
			var sig2 Signature
			sig2.Decode(bsig)
			So(Base58Encode(sig2.Pubkey), ShouldEqual, Base58Encode(sig.Pubkey))
			So(Base58Encode(sig2.Sig), ShouldEqual, Base58Encode(sig.Sig))
			So(sig.Algorithm, ShouldEqual, sig2.Algorithm)
		})

	})
}
