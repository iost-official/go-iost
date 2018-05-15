package hash

import (
	"crypto/sha256"
	"crypto/sha512"
)

func Sha256(raw []byte) Hash { return Hash(sha256.Sum256(raw)) }

func Sha512_256(raw []byte) Hash { return Hash(sha512.Sum512_256(raw)) }

func Sha512(raw []byte) LongHash { return LongHash(sha512.Sum512(raw)) }
