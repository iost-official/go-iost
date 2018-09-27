package crypto

import (
	"testing"

	"bytes"

	"github.com/iost-official/go-iost/common"
	. "github.com/smartystreets/goconvey/convey"
)

func TestSignature(t *testing.T) {
	Convey("Test of Signature", t, func() {
		Convey("Encode and decode", func() {
			//ac, _ := account.NewAccount(nil)
			//sig, err := Sign(Secp256k1, Sha256([]byte("hello")), ac.Seckey)
			//So(err, ShouldBeNil)
			//So(bytes.Equal(sig.Pubkey, ac.Pubkey), ShouldBeTrue)
			info := common.Sha256([]byte("hello"))
			seckey := common.Sha256([]byte("seckey"))
			pubkey := Secp256k1.GetPubkey(seckey)
			sig := NewSignature(Secp256k1, info, seckey)
			So(bytes.Equal(sig.Pubkey, pubkey), ShouldBeTrue)

			ans := sig.Verify(info)
			So(ans, ShouldBeTrue)

			bsig, _ := sig.Encode()
			var sig2 Signature
			sig2.Decode(bsig)
			So(common.Base58Encode(sig2.Pubkey), ShouldEqual, common.Base58Encode(sig.Pubkey))
			So(common.Base58Encode(sig2.Sig), ShouldEqual, common.Base58Encode(sig.Sig))
			So(sig.Algorithm, ShouldEqual, sig2.Algorithm)
		})

	})
}

func TestSign(t *testing.T) {
	testData := "c6e193266883a500c6e51a117e012d96ad113d5f21f42b28eb648be92a78f92f"
	privkey := common.ParseHex(testData)
	var pubkey []byte

	Convey("Test of Crypto", t, func() {
		Convey("Sha256", func() {
			sha := "d4daf0546cb71d90688b45488a8fa000b0821ec14b73677b2fb7788739228c8b"
			So(common.ToHex(common.Sha256(privkey)), ShouldEqual, sha)
		})

		Convey("Calculate public key", func() {
			pub := "0314bf901a6640033ea07b39c6b3acb675fc0af6a6ab526f378216085a93e5c7a2"
			pubkey = Secp256k1.GetPubkey(privkey)
			So(common.ToHex(pubkey), ShouldEqual, pub)
		})

		Convey("Hash-160", func() {
			hash := "9c1185a5c5e9fc54612808977ee8f548b2258d31"
			So(common.ToHex(common.Hash160(Secp256k1.GetPubkey(privkey))), ShouldEqual, hash)
		})

		Convey("SignInSecp256k1 and verify", func() {
			info := common.Sha256([]byte{1, 2, 3, 4})
			sig := Secp256k1.Sign(info, privkey)
			So(Secp256k1.Verify(info, pubkey, sig), ShouldBeTrue)
			So(Secp256k1.Verify(info, pubkey, []byte{5, 6, 7, 8}), ShouldBeFalse)
		})
	})
}
