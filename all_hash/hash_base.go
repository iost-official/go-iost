package all_hash

import (
  //"fmt"
  "encoding/hex"
  "encoding/base32"
  "encoding/base64"
)

const (
  HashLength = 32
  LongHashLength = 64
)

type Hash [HashLength]byte
type LongHash [LongHashLength]byte

func NewHash(bs []byte) Hash {
  b_len := len(bs)
  var h Hash

  if b_len < HashLength {
    for i, b := range bs {
      h[HashLength - b_len + i] = b
    }
  } else {
    for i, _ := range h {
      h[i] = bs[i]
    }
  }

  return h
}

func NewLongHash(bs []byte) LongHash {
  b_len := len(bs)
  var h LongHash

  if b_len < LongHashLength {
    for i, b := range bs {
      h[LongHashLength - b_len + i] = b
    }
  } else {
    for i, _ := range h {
      h[i] = bs[i]
    }
  }

  return h
}

func (h Hash) ToHex() string {return hex.EncodeToString(h[:])}

func (h Hash) ToBase32Std() string {return base32.StdEncoding.EncodeToString(h[:])}

func (h Hash) ToBase32Hex() string {return base32.HexEncoding.EncodeToString(h[:])}

func (h Hash) ToBase64Std() string {return base64.StdEncoding.EncodeToString(h[:])}

func (h Hash) ToBase64URL() string {return base64.URLEncoding.EncodeToString(h[:])}


func (h LongHash) ToHex() string {return hex.EncodeToString(h[:])}

func (h LongHash) ToBase32Std() string {return base32.StdEncoding.EncodeToString(h[:])}

func (h LongHash) ToBase32Hex() string {return base32.HexEncoding.EncodeToString(h[:])}

func (h LongHash) ToBase64Std() string {return base64.StdEncoding.EncodeToString(h[:])}

func (h LongHash) ToBase64URL() string {return base64.URLEncoding.EncodeToString(h[:])}
