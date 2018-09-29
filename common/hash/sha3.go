package hash

import (
	"golang.org/x/crypto/sha3"
)

// Sha3_256 ...
func Sha3_256(raw []byte) Hash {
	return Hash(sha3.Sum256(raw))
}

// Sha3_512 ...
func Sha3_512(raw []byte) LongHash {
	return LongHash(sha3.Sum512(raw))
}
