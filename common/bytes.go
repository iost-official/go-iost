package common

import (
	"encoding/binary"
	"math"
)

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

// Float64ToBytes converts float64 to byte slice.
func Float64ToBytes(f float64) []byte {
	bits := math.Float64bits(f)
	b := make([]byte, 8)
	binary.BigEndian.PutUint64(b, bits)
	return b
}
