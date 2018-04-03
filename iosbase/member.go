package iosbase

import (
	"fmt"
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
		rand.Seed(time.Now().UnixNano())
		bin := NewBinary()
		for i := 0; i < 4; i++ {
			bin.PutULong(rand.Uint64())
		}
		seckey = bin.bytes
	}
	if len(seckey) != 32 {
		return Member{}, fmt.Errorf("seckey length error")
	}
	m.Seckey = seckey
	m.Pubkey = CalcPubkey(seckey)
	m.ID = Base58Encode(m.Pubkey)
	return m, nil
}
