package common

import (
	"github.com/stretchr/testify/assert"
	"testing"
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
