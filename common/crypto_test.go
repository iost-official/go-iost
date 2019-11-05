package common

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSha3(t *testing.T) {
	input := []byte("abc")
	expected := []byte{0x3a, 0x98, 0x5d, 0xa7, 0x4f, 0xe2, 0x25, 0xb2, 0x4, 0x5c, 0x17, 0x2d, 0x6b, 0xd3, 0x90, 0xbd, 0x85, 0x5f, 0x8, 0x6e, 0x3e, 0x9d, 0x52, 0x5b, 0x46, 0xbf, 0xe2, 0x45, 0x11, 0x43, 0x15, 0x32}

	assert.Equal(t, expected, Sha3(input), "SHA3-256")
}

func TestRipemd160(t *testing.T) {
	input := ParseHex("823b54d3aabaf8e3122800ca5238afb2ccef071ce83b8d5654a597a5dd06347e")
	expected := ParseHex("3dbb2167cbfc2186343356125fff4163e6ebcce7")
	assert.Equal(t, expected, Ripemd160(input), "Ripemd160")
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
