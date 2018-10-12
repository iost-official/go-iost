package common

import (
	"github.com/stretchr/testify/assert"
	"testing"

	"strconv"
)

func TestIntToBytes(t *testing.T) {
	var input int = -1
	var expected []byte

	if strconv.IntSize == 64 {
		expected = []byte{0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF}
	} else {
		expected = []byte{0xFF, 0xFF, 0xFF, 0xFF}
	}

	assert.Equal(t, expected, IntToBytes(input), "IntToBytes returns input in Big-Endian")
}

func TestBytesToInt(t *testing.T) {
	var input []byte
	var expected int

	if strconv.IntSize == 64 {
		input = []byte{0x10, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}
		expected = 0x1000000000000000
	} else {
		input = []byte{0x10, 0x00, 0x00, 0x00}
		expected = 0x10000000
	}

	assert.Equal(t, expected, BytesToInt(input), "BytesToInt interprets byte array in Big-Endian")
}

func TestInt64ToBytes(t *testing.T) {
	var input int64 = -1
	expected := []byte{0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF}

	assert.Equal(t, expected, Int64ToBytes(input), "Int64ToBytes returns input in Big-Endian")
}

func TestBytesToInt64(t *testing.T) {
	input := []byte{0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF}
	var expected int64 = -1

	assert.Equal(t, expected, BytesToInt64(input), "BytesToInt64 interprets byte array in Big-Endian")
}

func TestUint64ToBytes(t *testing.T) {
	var input uint64 = 0xFFFFFFFFFFFFFFFF
	expected := []byte{0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF}

	assert.Equal(t, expected, Uint64ToBytes(input), "Uint64ToBytes returns input in Big-Endian")
}

func TestBytesToUint64(t *testing.T) {
	input := []byte{0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF}
	var expected uint64 = 0xFFFFFFFFFFFFFFFF

	assert.Equal(t, expected, BytesToUint64(input), "BytesToUint64 interprets byte array in Big-Endian")
}
