package iosbase

import (
	"encoding/binary"
	"bytes"
)

type Serializable interface {
	Bytes() []byte
}

type Binary struct {
	bytes  []byte
	length int
}

func NewBinary() Binary {
	var b Binary
	b.bytes = []byte{}
	b.length = 0
	return b
}

func BinaryFromBase58(s string) Binary {
	bits := Base58Decode(s)
	return Binary{bits, len(bits)}
}

func (bin *Binary) ToBase58() string {
	return Base58Encode(bin.bytes)
}

func (bin *Binary) PutUInt(u uint32) *Binary {
	var bBuf = bytes.NewBuffer([]byte{})
	binary.Write(bBuf, binary.BigEndian, u)
	bin.bytes = append(bin.bytes, bBuf.Bytes()...)
	bin.length += 4
	return bin
}

func (bin *Binary) PutULong(u uint64) *Binary {
	var bBuf = bytes.NewBuffer([]byte{})
	binary.Write(bBuf, binary.BigEndian, u)
	bin.bytes = append(bin.bytes, bBuf.Bytes()...)
	bin.length += 8
	return bin
}

func (bin *Binary) PutInt(i int) *Binary {
	var bBuf = bytes.NewBuffer([]byte{})
	binary.Write(bBuf, binary.BigEndian, i)
	bin.bytes = append(bin.bytes, bBuf.Bytes()...)
	bin.length += 4
	return bin
}

func (bin *Binary) PutBytes(b []byte) *Binary {
	bin.bytes = append(bin.bytes, b...)
	bin.length += len(b)
	return bin
}

func (bin *Binary) Put(o Serializable) *Binary {
	bytes := o.Bytes()
	bin.bytes = append(bin.bytes, bytes...)
	bin.length += len(bytes)
	return bin
}

func GetInt(b []byte, start, end int) int {
	var ans int
	bBuf := bytes.NewBuffer(b[start:end])
	binary.Read(bBuf, binary.BigEndian, &ans)
	return ans
}


func (bin *Binary) Bytes() []byte {
	return bin.bytes
}

func (bin Binary) Hash160() Binary {
	bin.bytes = Hash160(bin.bytes)
	return bin
}

func Equal(a []byte, b []byte) bool {
	if len(a) != len(b) {
		return false
	}
	for i := 0; i < len(a); i++ {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
