package account

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"math/rand"
	"time"

	"github.com/iost-official/prototype/common"
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
	m.ID = GetIdByPubkey(m.Pubkey)
	return m, nil
}

func (member *Account) GetId() string {
	return member.ID
}

func randomSeckey() []byte {
	rand.Seed(time.Now().UnixNano())
	bin := new(bytes.Buffer)
	for i := 0; i < 4; i++ {
		b := make([]byte, 8)
		binary.BigEndian.PutUint64(b, rand.Uint64())
		bin.Write(b)
	}
	seckey := bin.Bytes()
	return seckey
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
