package common

import (
	"crypto/sha256"
	"encoding/hex"

	"github.com/btcsuite/btcutil/base58"
	"github.com/ethereum/go-ethereum/crypto/secp256k1"
	"golang.org/x/crypto/ripemd160"
)

func Sha256(raw []byte) []byte {
	var data = sha256.Sum256(raw)
	return data[:]
}

func Hash160(raw []byte) []byte {
	var data = sha256.Sum256(raw)
	return ripemd160.New().Sum(data[len(data):])
}

func Base58Encode(raw []byte) string {
	return base58.Encode(raw)
}

func Base58Decode(s string) []byte {
	return base58.Decode(s)
}

func ToHex(data []byte) string {
	return hex.EncodeToString(data)
}

func ParseHex(s string) []byte {
	d, err := hex.DecodeString(s)
	if err != nil {
		println(err)
		return nil
	}
	return d
}

func Sign(info, privkey []byte) []byte {
	sig, err := secp256k1.Sign(info, privkey)
	if err != nil {
		println(err)
		return nil
	}
	return sig[:64]
}

func VerifySignature(info, pubkey, sig []byte) bool {
	return secp256k1.VerifySignature(pubkey, info, sig)
}

func CalcPubkey(privkey []byte) []byte {
	myCurve := secp256k1.S256()
	x, y := myCurve.ScalarBaseMult(privkey)
	return secp256k1.CompressPubkey(x, y)
}
