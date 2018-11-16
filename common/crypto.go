package common

import (
	"encoding/hex"

	"encoding/binary"
	"hash/crc32"

	"github.com/btcsuite/btcutil/base58"
	"golang.org/x/crypto/sha3"
)

// Sha3 ...
func Sha3(raw []byte) []byte {
	data := sha3.Sum256(raw)
	return data[:]
}

// Base58Encode ...
func Base58Encode(raw []byte) string {
	return base58.Encode(raw)
}

// Base58Decode ...
func Base58Decode(s string) []byte {
	return base58.Decode(s)
}

// Parity ...
func Parity(bit []byte) []byte {
	crc32q := crc32.MakeTable(crc32.Koopman)
	crc := crc32.Checksum(bit, crc32q)
	bs := make([]byte, 4)
	binary.LittleEndian.PutUint32(bs, crc)
	return bs
}

// ToHex ...
func ToHex(data []byte) string {
	return hex.EncodeToString(data)
}

// ParseHex ...
func ParseHex(s string) []byte {
	d, err := hex.DecodeString(s)
	if err != nil {
		println(err)
		return nil
	}
	return d
}
