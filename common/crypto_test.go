package common

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestSha256(t *testing.T) {
	input := []byte("abc")
	expected := []byte{0xba, 0x78, 0x16, 0xbf, 0x8f, 0x1, 0xcf, 0xea, 0x41, 0x41, 0x40, 0xde, 0x5d, 0xae, 0x22, 0x23, 0xb0, 0x3, 0x61, 0xa3, 0x96, 0x17, 0x7a, 0x9c, 0xb4, 0x10, 0xff, 0x61, 0xf2, 0x0, 0x15, 0xad}

	assert.Equal(t, expected, Sha256(input))
}

func TestSha3(t *testing.T) {
	input := []byte("abc")
	expected := []byte{0x3a, 0x98, 0x5d, 0xa7, 0x4f, 0xe2, 0x25, 0xb2, 0x4, 0x5c, 0x17, 0x2d, 0x6b, 0xd3, 0x90, 0xbd, 0x85, 0x5f, 0x8, 0x6e, 0x3e, 0x9d, 0x52, 0x5b, 0x46, 0xbf, 0xe2, 0x45, 0x11, 0x43, 0x15, 0x32}

	assert.Equal(t, expected, Sha3(input), "SHA3-256")
}

func TestHash160(t *testing.T) {
	input := []byte("abc")
	expected := []byte{0x9c, 0x11, 0x85, 0xa5, 0xc5, 0xe9, 0xfc, 0x54, 0x61, 0x28, 0x8, 0x97, 0x7e, 0xe8, 0xf5, 0x48, 0xb2, 0x25, 0x8d, 0x31}

	assert.Equal(t, expected, Hash160(input))
}

func TestBase58Encode(t *testing.T) {
	input := []byte{0x01, 0xB9, 0x7B}
	expected := "abc"

	assert.Equal(t, expected, Base58Encode(input))
}

func TestBase58Decode(t *testing.T) {
	input := "abc"
	expected := []byte{0x01, 0xB9, 0x7B}

	assert.Equal(t, expected, Base58Decode(input))
}

func TestParity(t *testing.T) {
	input := []byte("abc")
	expected := []byte{0xac, 0x22, 0x23, 0xba}

	assert.Equal(t, expected, Parity(input))
}

func TestToHex(t *testing.T) {
	input := []byte("The Times 03/Jan/2009 Chancellor on brink of second bailout for banks")
	expected := "5468652054696d65732030332f4a616e2f32303039204368616e63656c6c6f72206f6e206272696e6b206f66207365636f6e64206261696c6f757420666f722062616e6b73"

	assert.Equal(t, expected, ToHex(input))
}

func TestParseHex(t *testing.T) {
	input := "5468652054696d65732030332f4a616e2f32303039204368616e63656c6c6f72206f6e206272696e6b206f66207365636f6e64206261696c6f757420666f722062616e6b73"
	expected := []byte("The Times 03/Jan/2009 Chancellor on brink of second bailout for banks")

	assert.Equal(t, expected, ParseHex(input))
}
