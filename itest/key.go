package itest

import (
	"encoding/json"
	"os"

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

func LoadKeys(file string) ([]*Key, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}

	data := []byte{}
	if _, err := f.Read(data); err != nil {
		return nil, err
	}

	keys := []*Key{}
	if err := json.Unmarshal(data, keys); err != nil {
		return nil, err
	}

	return keys, nil
}

func DumpKeys(keys []*Key, file string) error {
	f, err := os.Create(file)
	if err != nil {
		return err
	}
	defer f.Close()

	b, err := json.Marshal(keys)
	if err != nil {
		return err
	}
	if _, err := f.Write(b); err != nil {
		return err
	}

	return nil
}

func NewKey(seckey []byte, algo crypto.Algorithm) *Key {
	keypair, err := account.NewKeyPair(seckey, algo)
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
