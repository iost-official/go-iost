// +build cgo

package backend

import (
	"crypto/rand"
	"fmt"

	"github.com/ethereum/go-ethereum/crypto/secp256k1"
	"github.com/iost-official/go-iost/v3/ilog"
)

// Secp256k1 is the secp256k1 crypto algorithm
type Secp256k1 struct{}

// Sign will signature the message with seckey by secp256k1
func (b *Secp256k1) Sign(message []byte, seckey []byte) []byte {
	sig, err := secp256k1.Sign(message, seckey)
	if err != nil {
		ilog.Errorf("Failed to sign, %v", err)
		return nil
	}
	return sig[:64]
}

// Verify will verify the message with pubkey and sig by secp256k1
func (b *Secp256k1) Verify(message []byte, pubkey []byte, sig []byte) bool {
	return secp256k1.VerifySignature(pubkey, message, sig)
}

// GetPubkey will get the public key of the secret key by secp256k1
func (b *Secp256k1) GetPubkey(seckey []byte) []byte {
	x, y := secp256k1.S256().ScalarBaseMult(seckey)
	return secp256k1.CompressPubkey(x, y)
}

// GenSeckey will generate the secret key by secp256k1
func (b *Secp256k1) GenSeckey() []byte {
	seckey := make([]byte, 32)
	_, err := rand.Read(seckey)
	if err != nil {
		ilog.Errorf("Failed to random seckey, %v", err)
		return nil
	}
	return seckey
}

// CheckSeckey ...
func (b *Secp256k1) CheckSeckey(seckey []byte) error {
	if len(seckey) != 32 {
		return fmt.Errorf("seckey length error secp256k1 seckey length should not be %v", len(seckey))
	}
	return nil
}
