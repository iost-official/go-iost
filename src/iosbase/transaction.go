package iosbase

import (
)

type Transaction interface {
	Verify(pool *StatePool) (bool, error)
	Transact(pool *StatePool)
	Bytes() []byte
}

