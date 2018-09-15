package common

import (
	"bytes"
	"encoding/binary"
)

const (
	// HashLength ...
	HashLength = 32
	// AddressLength ...
	AddressLength = 20
	// VoteInterval the interval between the votes
	VoteInterval = 200
)

// Hash is hash of data
type Hash [HashLength]byte

// BytesToHash ...
func BytesToHash(b []byte) Hash {
	var h Hash
	h.SetBytes(b)
	return h
}

// HexToHash ...
func HexToHash(s string) Hash { return BytesToHash(FromHex(s)) }

// Bytes return hash
func (h Hash) Bytes() []byte { return h[:] }

// SetBytes ...
func (h *Hash) SetBytes(b []byte) {
	if len(b) > len(h) {
		b = b[len(b)-HashLength:]
	}

	copy(h[HashLength-len(b):], b)
}

// IntToBytes ...
func IntToBytes(n int) []byte {
	x := int32(n)
	bytesBuffer := bytes.NewBuffer([]byte{})
	binary.Write(bytesBuffer, binary.BigEndian, x)
	return bytesBuffer.Bytes()
}

// BytesToInt ...
func BytesToInt(b []byte) int {
	bytesBuffer := bytes.NewBuffer(b)
	var x int32
	binary.Read(bytesBuffer, binary.BigEndian, &x)
	return int(x)
}

// Int64ToBytes ...
func Int64ToBytes(i int64) []byte {
	bytesBuffer := bytes.NewBuffer([]byte{})
	binary.Write(bytesBuffer, binary.BigEndian, i)
	return bytesBuffer.Bytes()
}

// BytesToInt64 ...
func BytesToInt64(b []byte) int64 {
	bytesBuffer := bytes.NewBuffer(b)
	var x int64
	binary.Read(bytesBuffer, binary.BigEndian, &x)
	return x
}

// Uint64ToBytes ...
func Uint64ToBytes(i uint64) []byte {
	buf := make([]byte, 8)
	binary.LittleEndian.PutUint64(buf, i)
	return buf
}

// BytesToUint64 ...
func BytesToUint64(b []byte) uint64 {
	return binary.LittleEndian.Uint64(b)
}
