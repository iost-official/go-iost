package common

import (
	"testing"

	"bytes"

	"github.com/iost-official/Go-IOS-Protocol/crypto"
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
			pubkey := crypto.Secp256k1.GetPubkey(seckey)
			sig := NewSignature(crypto.Secp256k1, info, seckey, true)
			So(bytes.Equal(sig.Pubkey, pubkey), ShouldBeTrue)

			ans := sig.Verify(info)
			So(ans, ShouldBeTrue)

			bsig, _ := sig.Encode()
			var sig2 Signature
			sig2.Decode(bsig)
			So(Base58Encode(sig2.Pubkey), ShouldEqual, Base58Encode(sig.Pubkey))
			So(Base58Encode(sig2.Sig), ShouldEqual, Base58Encode(sig.Sig))
			So(sig.Algorithm, ShouldEqual, sig2.Algorithm)
		})

	})
}
