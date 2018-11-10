package common

import (
	"encoding/binary"
	"encoding/hex"
)

// FromHex string to bytes
func FromHex(s string) []byte {
	if len(s) > 1 {
		if s[0:2] == "0x" || s[0:2] == "0X" {
			s = s[2:]
		}
	}
	if len(s)%2 == 1 {
		s = "0" + s
	}
	return Hex2Bytes(s)
}

// Hex2Bytes hex string to bytes
func Hex2Bytes(str string) []byte {
	h, _ := hex.DecodeString(str)
	return h
}

// CopyBytes ...
func CopyBytes(b []byte) (copiedBytes []byte) {
	if b == nil {
		return nil
	}
	copiedBytes = make([]byte, len(b))
	copy(copiedBytes, b)

	return
}

// Int64ToBytes converts int64 to byte slice.
func Int64ToBytes(n int64) []byte {
	b := make([]byte, 8)
	binary.BigEndian.PutUint64(b, uint64(n))
	return b
}

// BytesToInt64 converts byte slice to int64.
func BytesToInt64(b []byte) int64 {
	return int64(binary.BigEndian.Uint64(b))
}

// Int32ToBytes converts int32 to byte slice.
func Int32ToBytes(n int32) []byte {
	b := make([]byte, 4)
	binary.BigEndian.PutUint32(b, uint32(n))
	return b
}

// BytesToInt32 converts byte slice to int32.
func BytesToInt32(b []byte) int32 {
	return int32(binary.BigEndian.Uint32(b))
}
