package common

import (
	"bytes"
)

// SimpleEncoder is a simple encoder used to convert struct to bytes.
type SimpleEncoder struct {
	buf bytes.Buffer
}

// NewSimpleEncoder returns a new SimpleEncoder instance.
func NewSimpleEncoder() *SimpleEncoder {
	return &SimpleEncoder{}
}

// WriteByte writes a byte to buffer.
func (se *SimpleEncoder) WriteByte(b byte) { // nolint
	se.WriteBytes([]byte{b})
}

// WriteBytes writes a byte slice to buffer.
func (se *SimpleEncoder) WriteBytes(bs []byte) {
	se.WriteInt32(int32(len(bs)))
	se.buf.Write(bs)
}

// WriteString writes a string to buffer.
func (se *SimpleEncoder) WriteString(s string) {
	se.WriteBytes([]byte(s))
}

// WriteInt64 writes a int64 to buffer.
func (se *SimpleEncoder) WriteInt64(i int64) {
	se.buf.Write(Int64ToBytes(i))
}

// WriteInt32 writes a int32 to buffer.
func (se *SimpleEncoder) WriteInt32(i int32) {
	se.buf.Write(Int32ToBytes(i))
}

// WriteFloat64 writes a float64 to buffer.
func (se *SimpleEncoder) WriteFloat64(f float64) {
	se.buf.Write(Float64ToBytes(f))
}

// WriteBytesSlice writes a bytes slice to buffer.
func (se *SimpleEncoder) WriteBytesSlice(p [][]byte) {
	se.WriteInt32(int32(len(p)))
	for _, bs := range p {
		se.WriteBytes(bs)
	}
}

// WriteStringSlice writes a string slice to buffer.
func (se *SimpleEncoder) WriteStringSlice(p []string) {
	se.WriteInt32(int32(len(p)))
	for _, bs := range p {
		se.WriteString(bs)
	}
}

// WriteMapStringToI64 writes a map[string]int64 to buffer.
func (se *SimpleEncoder) WriteMapStringToI64(m map[string]int64) {
	key := make([]string, 0, len(m))
	for k := range m {
		key = append(key, k)
	}
	for i := 1; i < len(m); i++ {
		for j := 0; j < len(m)-i; j++ {
			if key[j] > key[j+1] {
				key[j], key[j+1] = key[j+1], key[j]
			}
		}
	}
	se.WriteInt32(int32(len(m)))
	for _, k := range key {
		se.WriteString(k)
		se.WriteInt64(m[k])
	}
}

// Bytes returns the result bytes of buffer.
func (se *SimpleEncoder) Bytes() []byte {
	return se.buf.Bytes()
}

// Reset resets the buffer.
func (se *SimpleEncoder) Reset() {
	se.buf.Reset()
}
