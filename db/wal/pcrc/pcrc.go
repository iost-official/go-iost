package pcrc

import (
	"hash/crc64"
	"hash"
)

// The size of a CRC-64 checksum in bytes.
const Size = 8

type digest struct {
	crc uint64
	tab *crc64.Table
}

func New(prev uint64, table *crc64.Table) hash.Hash64 {
	return &digest{prev, table}
}

func (d *digest) Size() int { return Size }

func (d *digest) BlockSize() int { return 1 }

func (d *digest) Reset() { d.crc = 0 }

func (d *digest) Write(p []byte) (n int, err error) {
	d.crc = crc64.Update(d.crc, d.tab, p)
	return len(p), nil
}

func (d *digest) Sum64() uint64 { return d.crc }

func (d *digest) Sum(in []byte) []byte {
	s := d.Sum64()
	return append(in, byte(s>>24), byte(s>>16), byte(s>>8), byte(s))
}
