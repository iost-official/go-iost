package all_hash

import (
	//"fmt"
	"encoding/base32"
	"encoding/base64"
	"encoding/hex"
)

const (
	HashLength     = 32
	LongHashLength = 64
)

type Hash [HashLength]byte
type LongHash [LongHashLength]byte

func NewHash(bs []byte) Hash {
	bLen := len(bs)
	var h Hash

	if bLen < HashLength {
		for i, b := range bs {
			h[HashLength-bLen+i] = b
		}
	} else {
		for i, _ := range h {
			h[i] = bs[i]
		}
	}

	return h
}

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

func (h Hash) ToHex() string { return hex.EncodeToString(h[:]) }

func (h Hash) ToBase32Std() string { return base32.StdEncoding.EncodeToString(h[:]) }

func (h Hash) ToBase32Hex() string { return base32.HexEncoding.EncodeToString(h[:]) }

func (h Hash) ToBase64Std() string { return base64.StdEncoding.EncodeToString(h[:]) }

func (h Hash) ToBase64URL() string { return base64.URLEncoding.EncodeToString(h[:]) }

func (h LongHash) ToHex() string { return hex.EncodeToString(h[:]) }

func (h LongHash) ToBase32Std() string { return base32.StdEncoding.EncodeToString(h[:]) }

func (h LongHash) ToBase32Hex() string { return base32.HexEncoding.EncodeToString(h[:]) }

func (h LongHash) ToBase64Std() string { return base64.StdEncoding.EncodeToString(h[:]) }

func (h LongHash) ToBase64URL() string { return base64.URLEncoding.EncodeToString(h[:]) }
