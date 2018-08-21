package account

import (
	"fmt"

	"crypto/rand"

	"github.com/iost-official/Go-IOS-Protocol/common"
)

var (
	MainAccount    Account
	GenesisAccount = map[string]int64{
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
	if seckey == nil {
		seckey = randomSeckey()
	}
	if len(seckey) != 32 {
		return Account{}, fmt.Errorf("seckey length error")
	}

	m.Seckey = seckey
	m.Pubkey = makePubkey(seckey)
	m.ID = GetIDByPubkey(m.Pubkey)
	return m, nil
}

func randomSeckey() []byte {
	seckey := make([]byte, 32)
	_, err := rand.Read(seckey)
	if err != nil {
		return nil
	}
	return seckey
}

func makePubkey(seckey []byte) []byte {
	return common.CalcPubkeyInSecp256k1(seckey)
}

func GetIDByPubkey(pubkey []byte) string {
	if len(pubkey) != 33 {
		panic("illegal pubkey")
	}
	return "IOST" + common.Base58Encode(append(pubkey, common.Parity(pubkey)...))
}

func GetPubkeyByID(ID string) []byte {
	b := common.Base58Decode(ID[4:])
	return b[:len(b)-4]
}
