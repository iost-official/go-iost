package iosbase

import (
	"math/big"
)

type State interface {
	GetValue() *big.Int
	GetScript() string
	GetHash() []byte
	Bytes() []byte
}

