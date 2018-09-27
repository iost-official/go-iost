package crypto

import (
	"github.com/iost-official/go-iost/crypto/backend"
)

// AlgorithmBackend is the interface of algorithm backend
type AlgorithmBackend interface {
	Sign(message []byte, seckey []byte) []byte
	Verify(message []byte, pubkey []byte, sig []byte) bool
	GetPubkey(seckey []byte) []byte
	GenSeckey() []byte
}

// Algorithm is the crypto algorithm of signature
type Algorithm uint8

// Algorithm list
const (
	_ Algorithm = iota
	Secp256k1
	Ed25519
)

func (a Algorithm) getBackend() AlgorithmBackend {
	switch a {
	case Secp256k1:
		return &backend.Secp256k1{}
	case Ed25519:
		return &backend.Ed25519{}
	default:
		return &backend.Secp256k1{}
	}
}

// Sign will signature the message with seckey
func (a Algorithm) Sign(message []byte, seckey []byte) []byte {
	return a.getBackend().Sign(message, seckey)
}

// Verify will verify the message with pubkey and sig
func (a Algorithm) Verify(message []byte, pubkey []byte, sig []byte) bool {
	return a.getBackend().Verify(message, pubkey, sig)
}

// GetPubkey will get the public key of the secret key
func (a Algorithm) GetPubkey(seckey []byte) []byte {
	return a.getBackend().GetPubkey(seckey)
}

// GenSeckey will generate the secret key
func (a Algorithm) GenSeckey() []byte {
	return a.getBackend().GenSeckey()
}
