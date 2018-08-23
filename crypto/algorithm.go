package crypto

import (
	"github.com/iost-official/Go-IOS-Protocol/crypto/backend"
)

type AlgorithmBackend interface {
	Sign(message []byte, seckey []byte) []byte
	Verify(message []byte, pubkey []byte, sig []byte) bool
	GetPubkey(seckey []byte) []byte
	GenSeckey() []byte
}

type Algorithm uint8

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

func (a Algorithm) Sign(message []byte, seckey []byte) []byte {
	return a.getBackend().Sign(message, seckey)
}

func (a Algorithm) Verify(message []byte, pubkey []byte, sig []byte) bool {
	return a.getBackend().Verify(message, pubkey, sig)
}

func (a Algorithm) GetPubkey(seckey []byte) []byte {
	return a.getBackend().GetPubkey(seckey)
}

func (a Algorithm) GenSeckey() []byte {
	return a.getBackend().GenSeckey()
}
