package itest

import (
	"encoding/json"
	"os"

	log "github.com/sirupsen/logrus"

	"github.com/iost-official/go-iost/account"
	"github.com/iost-official/go-iost/common"
	"github.com/iost-official/go-iost/crypto"
)

const (
	DefaultKeys = `
[
{"seckey":"XJXxYmjY3q89J8UaXbWx96MCNQEf2dV49j3adv4Qhjtb3948ukn5o4XLE2iyFFn3wJf73bmbf1SSGxXFSfgCtL2","algorithm":"ed25519"},
{"seckey":"474oNHzDAz3njim6M83DWq7vmre4dHT89hpoBc31Y3aemkWr6YdChzKFTZD671ZcMyVzKFsQ9898RU7yr9y5NNZ2","algorithm":"ed25519"},
{"seckey":"124iRRFSKS2NUU1QCqnXzNyE6roMxqiC66vtf84ZN6mKG2RjdkZrLJK8WRc26Sm82wAkB2LFb7qXhz3shJSLsT4U","algorithm":"ed25519"},
{"seckey":"5KR3weGjMX1S74U9jjbc9n2zsU5tKXA1SYbk5P72vpiBYNLHmQ6sLYnhQpScEuKRUDCqUgMTdgjG2qnw61v1TAik","algorithm":"ed25519"},
{"seckey":"4PUdaMimqbYiPh3eeCkBY4ZaDwYTT64YT59tUGxF8eXVyEGBgavQtJnkgnrMStKptf1YJdr2rAXkU6a8YzVE5Maa","algorithm":"ed25519"}
]
`
)

type Key struct {
	*account.KeyPair
}

type KeyJSON struct {
	Seckey    string `json:"seckey"`
	Algorithm string `json:"algorithm"`
}

func LoadKeys(file string) ([]*Key, error) {
	data := []byte{}
	if file == "" {
		data = []byte(DefaultKeys)
	} else {
		f, err := os.Open(file)
		if err != nil {
			return nil, err
		}

		if _, err := f.Read(data); err != nil {
			return nil, err
		}
	}

	keys := []*Key{}
	if err := json.Unmarshal(data, &keys); err != nil {
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

	b, err := json.Marshal(&keys)
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
	k.KeyPair, err = account.NewKeyPair(
		common.Base58Decode(aux.Seckey),
		crypto.NewAlgorithm(aux.Algorithm),
	)
	if err != nil {
		return err
	}
	return nil
}

func (k *Key) MarshalJSON() ([]byte, error) {
	aux := &KeyJSON{
		Seckey:    common.Base58Encode(k.KeyPair.Seckey),
		Algorithm: k.KeyPair.Algorithm.String(),
	}
	return json.Marshal(aux)
}
