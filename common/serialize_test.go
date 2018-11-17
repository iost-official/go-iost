package common

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestWriteByte(t *testing.T) {
	sn := NewSimpleNotation()
	sn.WriteByte('a', true)
	assert.Equal(t, []byte{'`', 'a'}, sn.Bytes())

	sn.WriteByte('\\', true)
	assert.Equal(t, []byte{'`', 'a', '`', '\\', '\\'}, sn.Bytes())

	sn.WriteByte('`', true)
	assert.Equal(t, []byte{'`', 'a', '`', '\\', '\\', '`', '\\', '`'}, sn.Bytes())

	sn.WriteByte('`', false)
	assert.Equal(t, []byte{'`', 'a', '`', '\\', '\\', '`', '\\', '`', '`', '`'}, sn.Bytes())
}

func TestWriteBytes(t *testing.T) {
	sn := NewSimpleNotation()
	sn.WriteBytes([]byte("aaa"), true)
	assert.Equal(t, []byte{'`', 'a', 'a', 'a'}, sn.Bytes())

	sn.WriteBytes([]byte("\\`"), true)
	assert.Equal(t, []byte{'`', 'a', 'a', 'a', '`', '\\', '\\', '\\', '`'}, sn.Bytes())

	sn.WriteBytes([]byte("\\`"), false)
	assert.Equal(t, []byte{'`', 'a', 'a', 'a', '`', '\\', '\\', '\\', '`', '`', '\\', '`'}, sn.Bytes())
}

func TestWriteInt64(t *testing.T) {
	sn := NewSimpleNotation()
	sn.WriteInt64(1023, true)
	assert.Equal(t, []byte{'`', 0, 0, 0, 0, 0, 0, 0x3, 0xff}, sn.Bytes())
}

func TestWriteInt32(t *testing.T) {
	sn := NewSimpleNotation()
	sn.WriteInt32(1023, true)
	assert.Equal(t, []byte{'`', 0, 0, 0x3, 0xff}, sn.Bytes())
}

func TestWriteBytesSlice(t *testing.T) {
	sn := NewSimpleNotation()
	sn.WriteBytesSlice([][]byte{[]byte("aa"), []byte("bb")}, true)
	assert.Equal(t, []byte{'`', '^', 'a', 'a', '^', 'b', 'b'}, sn.Bytes())

	sn.WriteBytesSlice([][]byte{[]byte("^`")}, true)
	assert.Equal(t, []byte{'`', '^', 'a', 'a', '^', 'b', 'b', '`', '^', '\\', '^', '\\', '`'}, sn.Bytes())
}

func TestWriteStringSlice(t *testing.T) {
	sn := NewSimpleNotation()
	sn.WriteStringSlice([]string{"aa", "bb"}, true)
	assert.Equal(t, []byte{'`', '^', 'a', 'a', '^', 'b', 'b'}, sn.Bytes())

	sn.WriteBytesSlice([][]byte{[]byte("^`")}, true)
	assert.Equal(t, []byte{'`', '^', 'a', 'a', '^', 'b', 'b', '`', '^', '\\', '^', '\\', '`'}, sn.Bytes())
}

func TestWriteMapStringToI64(t *testing.T) {
	sn := NewSimpleNotation()
	sn.WriteMapStringToI64(map[string]int64{"bb": 1024, "aa": 7}, true)
	assert.Equal(t, []byte{'`', '<', 'a', 'a', '/', 0, 0, 0, 0, 0, 0, 0, 0x7, '<', 'b', 'b', '/', 0, 0, 0, 0, 0, 0, 0x4, 0}, sn.Bytes())
}

func TestReset(t *testing.T) {
	sn := NewSimpleNotation()
	assert.Empty(t, sn.Bytes())
	sn.WriteInt32(1, true)
	assert.NotEqual(t, []byte(nil), sn.Bytes())
	sn.Reset()
	assert.Empty(t, sn.Bytes())
}
