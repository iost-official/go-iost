package account

import (
	"github.com/iost-official/go-iost/v3/common"
	"github.com/iost-official/go-iost/v3/crypto"
)

// KeyPair keypair of the iost account
type KeyPair struct {
	Algorithm crypto.Algorithm
	Pubkey    []byte
	Seckey    []byte
}

// NewKeyPair create a keypair
func NewKeyPair(seckey []byte, algo crypto.Algorithm) (*KeyPair, error) {
	if seckey == nil {
		seckey = algo.GenSeckey()
	}

	err := algo.CheckSeckey(seckey)
	if err != nil {
		return nil, err
	}

	pubkey := algo.GetPubkey(seckey)

	account := &KeyPair{
		Algorithm: algo,
		Pubkey:    pubkey,
		Seckey:    seckey,
	}
	return account, nil
}

// Sign sign a tx
func (a *KeyPair) Sign(info []byte) *crypto.Signature {
	return crypto.NewSignature(a.Algorithm, info, a.Seckey)
}

// ReadablePubkey ...
func (a *KeyPair) ReadablePubkey() string {
	return EncodePubkey(a.Pubkey)
}

// EncodePubkey ...
func EncodePubkey(pubkey []byte) string {
	return common.Base58Encode(pubkey)
}

// DecodePubkey ...
func DecodePubkey(readablePubKey string) []byte {
	return common.Base58Decode(readablePubKey)
}
