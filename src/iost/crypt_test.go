package iost

import (
	"fmt"
	"testing"
)

func TestSign(t *testing.T) {
	testData := "c6e193266883a500c6e51a117e012d96ad113d5f21f42b28eb648be92a78f92f"
	fmt.Println("============test of Crypt")
	fmt.Printf("hex test：255\n")
	fmt.Printf("to Hex > %v\n", ToHex([]byte{255}))
	fmt.Printf("parse hex > %v\n", ParseHex("ff"))
	fmt.Printf("private key：%v\n", testData)
	privkey := ParseHex(testData)
	fmt.Printf("sha256 > %x\n", Sha256(privkey))
	pubkey := CalcPubkey(ParseHex(testData))
	fmt.Printf("pubkey > %v\n", ToHex(pubkey))
	fmt.Printf("hash160 > %v\n", ToHex(Hash160(pubkey)))
	//fmt.Printf("base > %v\n", Base58Encode(Hash160(pubkey)))
	fmt.Printf("base58decode > %v\n", ToHex(Base58Decode("16UwLL9Risc3QfPqBUvKofHmBQ7wMtjvM")))
	fmt.Printf("base58encode > %v\n", Base58Encode(ParseHex("00010966776006953D5567439E5E39F86A0D273BEED61967F6")))
	fmt.Printf("sign to 0x1234567890: \n")
	sig := Sign(Sha256(ParseHex("1234567890")), privkey)
	fmt.Printf("sign > %v\n", ToHex(sig))
	fmt.Printf("verify > %v\n", VerifySignature(Sha256(ParseHex("1234567890")), pubkey, sig))
}
