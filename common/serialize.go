package common

import (
	"bytes"
	"fmt"
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
	se.buf.WriteByte(b)
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

// SimpleDecoder is a simple decoder used to convert bytes to other types. Not used now!!!
type SimpleDecoder struct {
	input []byte
}

// NewSimpleDecoder returns a new SimpleDecoder instance.
func NewSimpleDecoder(input []byte) *SimpleDecoder {
	return &SimpleDecoder{input}
}

// ParseByte parse input, return first byte
func (sd *SimpleDecoder) ParseByte() (byte, error) {
	if len(sd.input) < 1 {
		return 0, fmt.Errorf("parse byte fail: invalid len %v", sd.input)
	}
	result := sd.input[0]
	sd.input = sd.input[1:]
	return result, nil
}

// ParseInt32 parse input, return first int32
func (sd *SimpleDecoder) ParseInt32() (int32, error) {
	if len(sd.input) < 4 {
		return 0, fmt.Errorf("parse int32 fail: invalid len %v", sd.input)
	}
	result := BytesToInt32(sd.input[:4])
	sd.input = sd.input[4:]
	return result, nil
}

// ParseBytes parse input, return first byte array
func (sd *SimpleDecoder) ParseBytes() ([]byte, error) {
	length, err := sd.ParseInt32()
	if err != nil {
		return nil, err
	}
	if len(sd.input) < int(length) {
		return nil, fmt.Errorf("bytes length too large: %v > %v", length, len(sd.input))
	}
	result := sd.input[:length]
	sd.input = sd.input[length:]
	return result, nil
}
