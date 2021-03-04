// +build !cgo

package backend

import (
	"fmt"
	"math/big"

	"github.com/btcsuite/btcd/btcec"

	"github.com/iost-official/go-iost/v3/ilog"
)

// Secp256k1 is the secp256k1 crypto algorithm
type Secp256k1 struct{}

// Sign will signature the message with seckey by secp256k1
func (b *Secp256k1) Sign(hash []byte, seckey []byte) []byte {
	if len(hash) != 32 {
		ilog.Errorf("hash is required to be exactly 32 bytes (%d)", len(hash))
		return nil
	}
	privKey, _ := btcec.PrivKeyFromBytes(btcec.S256(), seckey)
	sig, err := btcec.SignCompact(btcec.S256(), privKey, hash, false)
	if err != nil {
		ilog.Errorf("Failed to sign, %v", err)
		return nil
	}
	return sig[1:]
}

// Verify will verify the message with pubkey and sig by secp256k1
func (b *Secp256k1) Verify(hash []byte, pubkey []byte, signature []byte) bool {
	if len(signature) != 64 {
		return false
	}
	pubKey, err := btcec.ParsePubKey(pubkey, btcec.S256())
	if err != nil {
		return false
	}
	sig := &btcec.Signature{R: new(big.Int).SetBytes(signature[:32]), S: new(big.Int).SetBytes(signature[32:])}
	return sig.Verify(hash, pubKey)
}

// GetPubkey will get the public key of the secret key by secp256k1
func (b *Secp256k1) GetPubkey(seckey []byte) []byte {
	_, pubKey := btcec.PrivKeyFromBytes(btcec.S256(), seckey)
	return pubKey.SerializeCompressed()
}

// GenSeckey will generate the secret key by secp256k1
func (b *Secp256k1) GenSeckey() []byte {
	seckey, err := btcec.NewPrivateKey(btcec.S256())
	if err != nil {
		ilog.Errorf("Failed to random seckey, %v", err)
		return nil
	}
	return seckey.Serialize()
}

// CheckSeckey ...
func (b *Secp256k1) CheckSeckey(seckey []byte) error {
	if len(seckey) != 32 {
		return fmt.Errorf("seckey length error secp256k1 seckey length should not be %v", len(seckey))
	}
	return nil
}
