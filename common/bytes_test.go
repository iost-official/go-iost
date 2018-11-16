package common

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFromHex(t *testing.T) {
	expected := []byte{1}

	assert := assert.New(t)

	assert.Equal(expected, FromHex("0x01"), "hex string starts with \"0x\"")
	assert.Equal(expected, FromHex("0X01"), "hex string starts with \"0X\"")
	assert.Equal(expected, FromHex("0x1"), "hex string with odd length")
}

func TestHex2Bytes(t *testing.T) {
	assert := assert.New(t)

	assert.Equal([]byte{1}, Hex2Bytes("01"), "vaild hex string")
	assert.Equal([]byte{}, Hex2Bytes("1"), "invalid hex string")
}

func TestCopyBytes(t *testing.T) {
	assert := assert.New(t)

	assert.Nil(CopyBytes(nil), "nil input")

	input := []byte{1, 1, 1}
	assert.Equal(input, CopyBytes(input), "normal input")
}

func TestInt64ToBytes(t *testing.T) {
	assert.Equal(t, []byte{0, 0, 0, 0, 0, 0, 0x3, 0xff}, Int64ToBytes(1023))
}

func TestBytesToInt64(t *testing.T) {
	assert.Equal(t, int64(1023), BytesToInt64([]byte{0, 0, 0, 0, 0, 0, 0x3, 0xff}))
}

func TestInt32ToBytes(t *testing.T) {
	assert.Equal(t, []byte{0, 0, 0x3, 0xff}, Int32ToBytes(1023))
}

func TestBytesToInt32(t *testing.T) {
	assert.Equal(t, int32(1023), BytesToInt32([]byte{0, 0, 0x3, 0xff}))
}
