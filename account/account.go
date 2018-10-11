package account

import (
	"fmt"

	"github.com/iost-official/go-iost/common"
	"github.com/iost-official/go-iost/crypto"
)

// Account account of the ios
type Account struct {
	ID        string
	Algorithm crypto.Algorithm
	Pubkey    []byte
	Seckey    []byte
}

// NewAccount create an account
func NewAccount(seckey []byte, algo crypto.Algorithm) (*Account, error) {
	if seckey == nil {
		seckey = algo.GenSeckey()
	}
	if (len(seckey) != 32 && algo == crypto.Secp256k1) ||
		(len(seckey) != 64 && algo == crypto.Ed25519) {
		return nil, fmt.Errorf("seckey length error")
	}
	pubkey := algo.GetPubkey(seckey)
	id := GetIDByPubkey(pubkey)

	account := &Account{
		ID:        id,
		Algorithm: algo,
		Pubkey:    pubkey,
		Seckey:    seckey,
	}
	return account, nil
}

// Sign sign a tx
func (a *Account) Sign(info []byte) *crypto.Signature {
	return crypto.NewSignature(a.Algorithm, info, a.Seckey)
}

// GetIDByPubkey ...
func GetIDByPubkey(pubkey []byte) string {
	return "IOST" + common.Base58Encode(append(pubkey, common.Parity(pubkey)...))
}

// GetPubkeyByID ...
func GetPubkeyByID(ID string) []byte {
	b := common.Base58Decode(ID[4:])
	return b[:len(b)-4]
}
