package backend

import (
	"crypto/rand"

	"github.com/ethereum/go-ethereum/crypto/secp256k1"
	"github.com/iost-official/Go-IOS-Protocol/ilog"
)

type Secp256k1 struct{}

func (b *Secp256k1) Sign(message []byte, seckey []byte) []byte {
	sig, err := secp256k1.Sign(message, seckey)
	if err != nil {
		ilog.Errorf("Failed to sign, %v", err)
		return nil
	}
	return sig[:64]
}

func (b *Secp256k1) Verify(message []byte, pubkey []byte, sig []byte) bool {
	return secp256k1.VerifySignature(pubkey, message, sig)
}

func (b *Secp256k1) GetPubkey(seckey []byte) []byte {
	x, y := secp256k1.S256().ScalarBaseMult(seckey)
	return secp256k1.CompressPubkey(x, y)
}

func (b *Secp256k1) GenSeckey() []byte {
	seckey := make([]byte, 32)
	_, err := rand.Read(seckey)
	if err != nil {
		ilog.Errorf("Failed to random seckey, %v", err)
		return nil
	}
	return seckey
}
