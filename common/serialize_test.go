package common

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestWriteByte(t *testing.T) {
	se := NewSimpleEncoder()
	se.WriteByte('a')
	assert.Equal(t, []byte{0, 0, 0, 0x1, 'a'}, se.Bytes())

	se.WriteByte('b')
	assert.Equal(t, []byte{0, 0, 0, 0x1, 'a', 0, 0, 0, 0x1, 'b'}, se.Bytes())
}

func TestWriteBytes(t *testing.T) {
	se := NewSimpleEncoder()
	se.WriteBytes([]byte("aaa"))
	assert.Equal(t, []byte{0, 0, 0, 0x3, 'a', 'a', 'a'}, se.Bytes())

	se.WriteBytes([]byte("bb"))
	assert.Equal(t, []byte{0, 0, 0, 0x3, 'a', 'a', 'a', 0, 0, 0, 0x2, 'b', 'b'}, se.Bytes())
}

func TestWriteString(t *testing.T) {
	se := NewSimpleEncoder()
	se.WriteString("aaa")
	assert.Equal(t, []byte{0, 0, 0, 0x3, 'a', 'a', 'a'}, se.Bytes())

	se.WriteString("bb")
	assert.Equal(t, []byte{0, 0, 0, 0x3, 'a', 'a', 'a', 0, 0, 0, 0x2, 'b', 'b'}, se.Bytes())
}

func TestWriteInt64(t *testing.T) {
	se := NewSimpleEncoder()
	se.WriteInt64(1023)
	assert.Equal(t, []byte{0, 0, 0, 0, 0, 0, 0x3, 0xff}, se.Bytes())
}

func TestWriteInt32(t *testing.T) {
	se := NewSimpleEncoder()
	se.WriteInt32(1023)
	assert.Equal(t, []byte{0, 0, 0x3, 0xff}, se.Bytes())
}

func TestWriteBytesSlice(t *testing.T) {
	se := NewSimpleEncoder()
	se.WriteBytesSlice([][]byte{[]byte("aa"), []byte("bb")})
	assert.Equal(t, []byte{0, 0, 0, 0x2, 0, 0, 0, 0x2, 'a', 'a', 0, 0, 0, 0x2, 'b', 'b'}, se.Bytes())
}

func TestWriteStringSlice(t *testing.T) {
	se := NewSimpleEncoder()
	se.WriteStringSlice([]string{"aa", "bb"})
	assert.Equal(t, []byte{0, 0, 0, 0x2, 0, 0, 0, 0x2, 'a', 'a', 0, 0, 0, 0x2, 'b', 'b'}, se.Bytes())
}

func TestWriteMapStringToI64(t *testing.T) {
	se := NewSimpleEncoder()
	se.WriteMapStringToI64(map[string]int64{"bb": 1024, "aa": 7})
	assert.Equal(t, []byte{0, 0, 0, 0x2, 0, 0, 0, 0x2, 'a', 'a', 0, 0, 0, 0, 0, 0, 0, 0x7, 0, 0, 0, 0x2, 'b', 'b', 0, 0, 0, 0, 0, 0, 0x4, 0}, se.Bytes())
}

func TestReset(t *testing.T) {
	se := NewSimpleEncoder()
	assert.Empty(t, se.Bytes())
	se.WriteInt32(1)
	assert.NotEqual(t, []byte(nil), se.Bytes())
	se.Reset()
	assert.Empty(t, se.Bytes())
}
