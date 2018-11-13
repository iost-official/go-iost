package common

import (
	"bytes"
)

// SimpleNotation is a simple serialize notation used to convert struct to bytes.
type SimpleNotation struct {
	buf bytes.Buffer

	sliceSep byte // to separate item of slice in one field
	fieldSep byte // to separate fields
	mapKVSep byte // to separate one key-value
	mapSep   byte // to separate different key-values in a map
	slash    byte
}

// NewSimpleNotation returns a new SimpleNotation instance.
func NewSimpleNotation() *SimpleNotation {
	return &SimpleNotation{
		sliceSep: '^',
		fieldSep: '`',
		mapKVSep: '/',
		mapSep:   '<',
		slash:    '\\',
	}
}

func (sn *SimpleNotation) escape(bs []byte) []byte {
	n := bytes.Replace(bs, []byte{sn.slash}, []byte{sn.slash, sn.slash}, -1)
	n = bytes.Replace(n, []byte{sn.sliceSep}, []byte{sn.slash, sn.sliceSep}, -1)
	n = bytes.Replace(n, []byte{sn.fieldSep}, []byte{sn.slash, sn.fieldSep}, -1)
	n = bytes.Replace(n, []byte{sn.mapKVSep}, []byte{sn.slash, sn.mapKVSep}, -1)
	n = bytes.Replace(n, []byte{sn.mapSep}, []byte{sn.slash, sn.mapSep}, -1)
	return n
}

// WriteByte writes a byte to buffer.
func (sn *SimpleNotation) WriteByte(b byte, escape bool) { // nolint
	sn.WriteBytes([]byte{b}, escape)
}

// WriteBytes writes a byte slice to buffer.
func (sn *SimpleNotation) WriteBytes(bs []byte, escape bool) {
	sn.buf.WriteByte(sn.fieldSep)
	if escape {
		bs = sn.escape(bs)
	}
	sn.buf.Write(bs)
}

// WriteString writes a string to buffer.
func (sn *SimpleNotation) WriteString(s string, escape bool) {
	sn.WriteBytes([]byte(s), escape)
}

// WriteInt64 writes a int64 to buffer.
func (sn *SimpleNotation) WriteInt64(i int64, escape bool) {
	sn.WriteBytes(Int64ToBytes(i), escape)
}

// WriteInt32 writes a int32 to buffer.
func (sn *SimpleNotation) WriteInt32(i int32, escape bool) {
	sn.WriteBytes(Int32ToBytes(i), escape)
}

// WriteBytesSlice writes a bytes slice to buffer.
func (sn *SimpleNotation) WriteBytesSlice(p [][]byte, escape bool) {
	var buf bytes.Buffer
	for _, bs := range p {
		buf.WriteByte(sn.sliceSep)
		if escape {
			bs = sn.escape(bs)
		}
		buf.Write(bs)
	}

	sn.WriteBytes(buf.Bytes(), false)
}

// WriteStringSlice writes a string slice to buffer.
func (sn *SimpleNotation) WriteStringSlice(p []string, escape bool) {
	var buf bytes.Buffer
	for _, s := range p {
		buf.WriteByte(sn.sliceSep)
		bs := []byte(s)
		if escape {
			bs = sn.escape(bs)
		}
		buf.Write(bs)
	}

	sn.WriteBytes(buf.Bytes(), false)
}

// WriteMapStringToI64 writes a map[string]int64 to buffer.
func (sn *SimpleNotation) WriteMapStringToI64(m map[string]int64, escape bool) {
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
	var buf bytes.Buffer
	for _, k := range key {
		buf.WriteByte(sn.mapSep)
		kb := []byte(k)
		vb := Int64ToBytes(m[k])
		if escape {
			kb = sn.escape(kb)
			vb = sn.escape(vb)
		}
		buf.Write(append(append(kb, sn.mapKVSep), vb...))
	}

	sn.WriteBytes(buf.Bytes(), false)
}

// Bytes returns the result bytes of buffer.
func (sn *SimpleNotation) Bytes() []byte {
	return sn.buf.Bytes()
}

// Reset resets the buffer.
func (sn *SimpleNotation) Reset() {
	sn.buf.Reset()
}
