package crypto

import (
	"testing"

	"bytes"
	"encoding/base64"
	"fmt"

	"github.com/iost-official/go-iost/v3/common"
	. "github.com/smartystreets/goconvey/convey"
)

func TestSignature(t *testing.T) {
	Convey("Test of Signature", t, func() {
		Convey("Encode and decode", func() {
			//ac, _ := account.NewAccount(nil)
			//sig, err := Sign(Secp256k1, Sha3([]byte("hello")), ac.Seckey)
			//So(err, ShouldBeNil)
			//So(bytes.Equal(sig.Pubkey, ac.Pubkey), ShouldBeTrue)
			info := common.Sha3([]byte("hello"))
			seckey := common.Sha3([]byte("seckey"))
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
		Convey("Sha3", func() {
			sha := "f420b28b56ce97e52adf4778a72b622c3e91115445026cf6e641459ec478dae8"
			So(common.ToHex(common.Sha3(privkey)), ShouldEqual, sha)
		})

		Convey("Calculate public key", func() {
			pub := "0314bf901a6640033ea07b39c6b3acb675fc0af6a6ab526f378216085a93e5c7a2"
			pubkey = Secp256k1.GetPubkey(privkey)
			So(common.ToHex(pubkey), ShouldEqual, pub)
		})

		Convey("SignInSecp256k1 and verify", func() {
			info := common.Sha3([]byte{1, 2, 3, 4})
			sig := Secp256k1.Sign(info, privkey)
			So(Secp256k1.Verify(info, pubkey, sig), ShouldBeTrue)
			So(Secp256k1.Verify(info, pubkey, []byte{5, 6, 7, 8}), ShouldBeFalse)
		})
	})
}

func TestSignature_Platform(t *testing.T) {
	algo := NewAlgorithm("ed25519")

	seckey := common.Base58Decode("1rANSfcRzr4HkhbUFZ7L1Zp69JZZHiDDq5v7dNSbbEqeU4jxy3fszV4HGiaLQEyqVpS1dKT9g7zCVRxBVzuiUzB")
	t.Log(fmt.Sprintf("seckey > %x", seckey))
	info := common.Sha3([]byte("hello"))
	t.Log(fmt.Sprintf("info   > %x", info))
	pubkey := algo.GetPubkey(seckey)
	t.Log(fmt.Sprintf("pubkey > %x", pubkey))
	t.Log("pubkey in base64 >", base64.StdEncoding.EncodeToString(pubkey))

	sig := algo.Sign(info, seckey)
	t.Log(fmt.Sprintf("sig    > %x", sig))
	t.Log("sig in base64 >", base64.StdEncoding.EncodeToString(sig))

	t.Log("secp256k1-----------------")

	secp := NewAlgorithm("secp256k1")
	sec := common.Base58Decode("EhNiaU4DzUmjCrvynV3gaUeuj2VjB1v2DCmbGD5U2nSE")
	t.Log(fmt.Sprintf("seckey > %x", sec))

	pubkey2 := secp.GetPubkey(sec)
	t.Log(fmt.Sprintf("pubkey    > %x", pubkey2))
	t.Log("pubkey in base64 >", base64.StdEncoding.EncodeToString(pubkey2))

	sig2 := secp.Sign(info, sec)
	t.Log(fmt.Sprintf("sig    > %x", sig2))
	t.Log("sig in base64 >", base64.StdEncoding.EncodeToString(sig2))

}
