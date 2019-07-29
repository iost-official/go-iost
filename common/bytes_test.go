package common

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

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

func TestFloat64ToBytes(t *testing.T) {
	assert.Equal(t, []byte{0x3f, 0xf0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0}, Float64ToBytes(1))
}

func TestBytesToFloat64(t *testing.T) {
	assert.Equal(t, float64(1), BytesToFloat64([]byte{0x3f, 0xf0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0}))
}
