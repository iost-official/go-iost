package crypto

import (
	"github.com/iost-official/go-iost/crypto/backend"
	"github.com/iost-official/go-iost/ilog"
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

// NewAlgorithm returns a new algorithm by name
func NewAlgorithm(name string) Algorithm {
	switch name {
	case "secp256k1":
		return Secp256k1
	case "ed25519":
		return Ed25519
	default:
		return Ed25519
	}
}

// String return algorithm readable string
func (a Algorithm) String() string {
	switch a {
	case Secp256k1:
		return "secp256k1"
	case Ed25519:
		return "ed25519"
	default:
		return "secp256k1"
	}
}

// Sign will signature the message with seckey
func (a Algorithm) Sign(message []byte, seckey []byte) []byte {
	return a.getBackend().Sign(message, seckey)
}

// Verify will verify the message with pubkey and sig
func (a Algorithm) Verify(message []byte, pubkey []byte, sig []byte) (ret bool) {
	// catch ed25519.Verify panic
	defer func() {
		if e := recover(); e != nil {
			ilog.Warnf("verify panic. err=%v", e)
			ret = false
		}
	}()
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
