package core

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"github.com/iost-official/prototype/common"
	"math/rand"
	"time"
)

type Member struct {
	ID     string
	Pubkey []byte
	Seckey []byte
}

func NewMember(seckey []byte) (Member, error) {
	var m Member
	if seckey == nil {
		seckey = randomSeckey()
	}
	if len(seckey) != 32 {
		return Member{}, fmt.Errorf("seckey length error")
	}

	m.Seckey = seckey
	m.Pubkey = makePubkey(seckey)
	m.ID = GetIdByPubkey(m.Pubkey)
	return m, nil
}

func (member *Member)GetId() string {
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
