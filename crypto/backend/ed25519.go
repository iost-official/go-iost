package backend

import (
	"crypto/rand"

	"github.com/iost-official/Go-IOS-Protocol/ilog"
	"golang.org/x/crypto/ed25519"
)

type Ed25519 struct{}

func (b *Ed25519) Sign(message []byte, seckey []byte) []byte {
	return ed25519.Sign(seckey, message)
}

func (b *Ed25519) Verify(message []byte, pubkey []byte, sig []byte) bool {
	return ed25519.Verify(pubkey, message, sig)
}

func (b *Ed25519) GetPubkey(seckey []byte) []byte {
	pubkey, ok := (ed25519.PrivateKey(seckey).Public()).([]byte)
	if !ok {
		ilog.Errorf("Failed to assert ed25519.PublicKey to []byte")
		return nil
	}
	return pubkey
}

func (b *Ed25519) GenSeckey() []byte {
	seed := make([]byte, 32)
	_, err := rand.Read(seed)
	if err != nil {
		ilog.Errorf("Failed to random seed, %v", err)
		return nil
	}
	return ed25519.NewKeyFromSeed(seed)
}
