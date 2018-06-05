package trie

/*
import (
	"bytes"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestHexCompact(t *testing.T) {
	Convey("Test HexCompact", t, func() {
		tests := []struct{ hex, compact []byte }{
			{hex: []byte{}, compact: []byte{0x00}},
			{hex: []byte{16}, compact: []byte{0x20}},
			{hex: []byte{1, 2, 3, 4, 5}, compact: []byte{0x11, 0x23, 0x45}},
			{hex: []byte{0, 1, 2, 3, 4, 5}, compact: []byte{0x00, 0x01, 0x23, 0x45}},
			{hex: []byte{15, 1, 12, 11, 8, 16}, compact: []byte{0x3f, 0x1c, 0xb8}},
			{hex: []byte{0, 15, 1, 12, 11, 8, 16}, compact: []byte{0x20, 0x0f, 0x1c, 0xb8}},
		}
		for _, test := range tests {
			So(true, ShouldEqual, bytes.Equal(hexToCompact(test.hex), test.compact))
			So(true, ShouldEqual, bytes.Equal(compactToHex(test.compact), test.hex))
		}
	})
}

func TestHexKeybytes(t *testing.T) {
	Convey("Test HexKeybytes", t, func() {
		tests := []struct{ key, hexIn, hexOut []byte }{
			{key: []byte{}, hexIn: []byte{16}, hexOut: []byte{16}},
			{key: []byte{}, hexIn: []byte{}, hexOut: []byte{16}},
			{
				key:    []byte{0x12, 0x34, 0x56},
				hexIn:  []byte{1, 2, 3, 4, 5, 6, 16},
				hexOut: []byte{1, 2, 3, 4, 5, 6, 16},
			},
			{
				key:    []byte{0x12, 0x34, 0x5},
				hexIn:  []byte{1, 2, 3, 4, 0, 5, 16},
				hexOut: []byte{1, 2, 3, 4, 0, 5, 16},
			},
			{
				key:    []byte{0x12, 0x34, 0x56},
				hexIn:  []byte{1, 2, 3, 4, 5, 6},
				hexOut: []byte{1, 2, 3, 4, 5, 6, 16},
			},
		}
		for _, test := range tests {
			So(true, ShouldEqual, bytes.Equal(keybytesToHex(test.key), test.hexOut))
			So(true, ShouldEqual, bytes.Equal(hexToKeybytes(test.hexIn), test.key))
		}
	})
}
*/
