package account

import (
	"fmt"
	"github.com/iost-official/go-iost/common"
	"crypto/rand"
)

var (
	MainAccount    Account
	GenesisAccount = map[string]float64{
		"2BibFrAhc57FAd3sDJFbPqjwskBJb5zPDtecPWVRJ1jxT": 3400000000,
		"tUFikMypfNGxuJcNbfreh8LM893kAQVNTktVQRsFYuEU":  3200000000,
		"s1oUQNTcRKL7uqJ1aRqUMzkAkgqJdsBB7uW9xrTd85qB":  3100000000,
		"22zr9ows3qndmAjnkiPFex26taATEaEfjGkatVCr5akSU": 3000000000,
		"wSKjLjqWbhH2LcJFwTW9Nfq9XPdhb4pw9KCM7QGtemZG":  2900000000,
		"oh7VBi17aQvG647cTfhhoRGby3tH55o3Qv7YHWD5q8XU":  2800000000,
		"28mKnLHaVvc1YRKc9CWpZxCpo2gLVCY3RL5nC9WbARRym": 2600000000,
	}
	// local net
	//GenesisAccount = map[string]float64{
	//	"iWgLQj3VTPN4dZnomuJMMCggv22LFw4nAkA6bmrVsmCo":  13400000000,
	//	"281pWKbjMYGWKf2QHXUKDy4rVULbF61WGCZoi4PiKhbEk": 13200000000,
	//	"bj38rN9xdqBa4eiMi1vPjcUwdMyZmQhvYbVA6cnHyQCH":  13100000000,
	//}
)

type Account struct {
	ID     string
	Pubkey []byte
	Seckey []byte
}

func NewAccount(seckey []byte) (Account, error) {
	var m Account
	var err error
	if seckey == nil {
		seckey, err = randomSeckey()
		if err != nil {
			return Account{}, err
		}
	}
	if len(seckey) != 32 {
		return Account{}, fmt.Errorf("seckey length error")
	}

	m.Seckey = seckey
	m.Pubkey = makePubkey(seckey)
	m.ID = GetIdByPubkey(m.Pubkey)
	return m, nil
}

func (member *Account) GetId() string {
	return member.ID
}

func randomSeckey() ([]byte, error) {
	seckey := make([]byte, 32)
	_, err := rand.Read(seckey)
	if err != nil {
		return nil, err
	}
	return seckey, nil
}

func makePubkey(seckey []byte) []byte {
	return common.CalcPubkeyInSecp256k1(seckey)
}

func GetIdByPubkey(pubkey []byte) string {
	return common.Base58Encode(pubkey)
}

func GetPubkeyByID(ID string) []byte {
	return common.Base58Decode(ID)
}
