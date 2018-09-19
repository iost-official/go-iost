package hash

import (
	//"fmt"
	"encoding/base32"
	"encoding/base64"
	"encoding/hex"
)

const (
	// HashLength ...
	HashLength = 32
	// LongHashLength ...
	LongHashLength = 64
)

// Hash ...
type Hash [HashLength]byte

// LongHash ...
type LongHash [LongHashLength]byte

// NewHash ...
func NewHash(bs []byte) Hash {
	bLen := len(bs)
	var h Hash

	if bLen < HashLength {
		for i, b := range bs {
			h[HashLength-bLen+i] = b
		}
	} else {
		for i := range h {
			h[i] = bs[i]
		}
	}

	return h
}

// NewLongHash ...
func NewLongHash(bs []byte) LongHash {
	bLen := len(bs)
	var h LongHash

	if bLen < LongHashLength {
		for i, b := range bs {
			h[LongHashLength-bLen+i] = b
		}
	} else {
		for i := range h {
			h[i] = bs[i]
		}
	}

	return h
}

// ToHex ...
func (h Hash) ToHex() string { return hex.EncodeToString(h[:]) }

// ToBase32Std ...
func (h Hash) ToBase32Std() string { return base32.StdEncoding.EncodeToString(h[:]) }

// ToBase32Hex ...
func (h Hash) ToBase32Hex() string { return base32.HexEncoding.EncodeToString(h[:]) }

// ToBase64Std ...
func (h Hash) ToBase64Std() string { return base64.StdEncoding.EncodeToString(h[:]) }

// ToBase64URL ...
func (h Hash) ToBase64URL() string { return base64.URLEncoding.EncodeToString(h[:]) }

// ToHex ...
func (h LongHash) ToHex() string { return hex.EncodeToString(h[:]) }

// ToBase32Std ...
func (h LongHash) ToBase32Std() string { return base32.StdEncoding.EncodeToString(h[:]) }

// ToBase32Hex ...
func (h LongHash) ToBase32Hex() string { return base32.HexEncoding.EncodeToString(h[:]) }

// ToBase64Std ...
func (h LongHash) ToBase64Std() string { return base64.StdEncoding.EncodeToString(h[:]) }

// ToBase64URL ...
func (h LongHash) ToBase64URL() string { return base64.URLEncoding.EncodeToString(h[:]) }
