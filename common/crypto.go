package common

import (
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"hash/crc32"

	"github.com/btcsuite/btcutil/base58"
	"github.com/iost-official/go-iost/ilog"
	"golang.org/x/crypto/ripemd160" // nolint:staticcheck
	"golang.org/x/crypto/sha3"
)

// Sha3 ...
func Sha3(raw []byte) []byte {
	defer func() {
		if e := recover(); e != nil {
			ilog.Warnf("sha3 panic. err=%v", e)
		}
	}()
	data := sha3.Sum256(raw)
	return data[:]
}

// Sha256 ...
func Sha256(raw []byte) []byte {
	defer func() {
		if e := recover(); e != nil {
			ilog.Warnf("sha256 panic. err=%v", e)
		}
	}()
	data := sha256.Sum256(raw)
	return data[:]
}

// Ripemd160 ...
func Ripemd160(raw []byte) []byte {
	h := ripemd160.New()
	h.Write(raw)
	return h.Sum(nil)
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
