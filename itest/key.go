package itest

import (
	"encoding/json"

	log "github.com/sirupsen/logrus"

	"github.com/iost-official/go-iost/account"
	"github.com/iost-official/go-iost/crypto"
)

type Key struct {
	*account.KeyPair
}

type KeyJSON struct {
	Seckey    []byte           `json:"seckey"`
	Algorithm crypto.Algorithm `json:"algorithm"`
}

func NewKey(algo crypto.Algorithm) *Key {
	keypair, err := account.NewKeyPair(nil, algo)
	if err != nil {
		log.Fatalf("Create key pair failed: %v", err)
	}
	return &Key{
		KeyPair: keypair,
	}
}

func (k *Key) UnmarshalJSON(b []byte) error {
	aux := &KeyJSON{}
	err := json.Unmarshal(b, aux)
	if err != nil {
		return err
	}
	k.KeyPair, err = account.NewKeyPair(aux.Seckey, aux.Algorithm)
	if err != nil {
		return err
	}
	return nil
}

func (k *Key) MarshalJSON() ([]byte, error) {
	aux := &KeyJSON{
		Seckey:    k.KeyPair.Seckey,
		Algorithm: k.KeyPair.Algorithm,
	}
	return json.Marshal(aux)
}
