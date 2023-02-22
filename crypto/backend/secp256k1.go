package backend

import (
	"fmt"

	secp "github.com/decred/dcrd/dcrec/secp256k1/v4"
	secp_ecdsa "github.com/decred/dcrd/dcrec/secp256k1/v4/ecdsa"

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
	privKey := secp.PrivKeyFromBytes(seckey)
	sig := secp_ecdsa.SignCompact(privKey, hash, false)
	return sig[1:]
}

// Verify will verify the message with pubkey and sig by secp256k1
func (b *Secp256k1) Verify(hash []byte, pubkey []byte, signature []byte) bool {
	if len(signature) != 64 {
		return false
	}
	pubKey, err := secp.ParsePubKey(pubkey)
	if err != nil {
		return false
	}
	var r, s secp.ModNScalar
	if r.SetByteSlice(signature[:32]) {
		return false // overflow
	}
	if s.SetByteSlice(signature[32:]) {
		return false
	}
	if s.IsOverHalfOrder() {
		return false
	}
	sig := secp_ecdsa.NewSignature(&r, &s)
	return sig.Verify(hash, pubKey)
}

// GetPubkey will get the public key of the secret key by secp256k1
func (b *Secp256k1) GetPubkey(seckey []byte) []byte {
	privKey := secp.PrivKeyFromBytes(seckey)
	pubKey := privKey.PubKey()
	return pubKey.SerializeCompressed()
}

// GenSeckey will generate the secret key by secp256k1
func (b *Secp256k1) GenSeckey() []byte {
	seckey, err := secp.GeneratePrivateKey()
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
