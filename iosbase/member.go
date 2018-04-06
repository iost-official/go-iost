package iosbase

import (
	"fmt"
	"math/rand"
	"time"
)

type Member struct {
	ID       string
	Username string
	Pubkey   []byte
	Seckey   []byte
}

func NewMember(seckey []byte, name string) (Member, error) {
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
	if len(name) > 15 || len(name) == 0 {
		return Member{}, fmt.Errorf("name length error")
	}
	for c := range name {
		if !IsValidNameChar(c) {
			return Member{}, fmt.Errorf("invalid char in name")
		}
	}
	m.Seckey = seckey
	m.Pubkey = CalcPubkey(seckey)
	m.ID = Base58Encode(m.Pubkey)
	m.Username = name
	return m, nil
}

func IsValidNameChar(c rune) bool {
	str_map := VALID_USERNAME
	return str_map[c]
}
