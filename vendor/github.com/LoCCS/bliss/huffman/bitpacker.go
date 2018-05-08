package huffman

import (
	"fmt"
)

const (
	MAX_PACKER_SIZE int = 8192
)

type BitPacker struct {
	data  [MAX_PACKER_SIZE]byte
	nbyte uint32
	nbit  uint32
}

type BitUnpacker struct {
	data  []byte
	pbyte uint32
	pbit  uint32
	size  uint32
}

func NewBitPacker() *BitPacker {
	return &BitPacker{
		[MAX_PACKER_SIZE]byte{}, 0, 0,
	}
}

func NewBitUnpacker(data []byte, nbit uint32) *BitUnpacker {
	if int(nbit) > len(data)*8 {
		return nil
	}
	return &BitUnpacker{
		data, 0, 0, nbit,
	}
}

func (packer *BitPacker) WriteBits(code uint64, nbit uint32) error {
	if int(packer.nbyte) >= MAX_PACKER_SIZE {
		return fmt.Errorf("Packer full!")
	}
	for nbit > 0 {
		if int(packer.nbyte) >= MAX_PACKER_SIZE {
			return fmt.Errorf("Packer full!")
		}
		left := 8 - packer.nbit
		if nbit < left {
			// packer.data[packer.nbyte] <<= nbit
			packer.data[packer.nbyte] |= byte(code&uint64(uint32(1<<nbit)-1)) << (left - nbit)
			packer.nbit += nbit
			return nil
		} else {
			// packer.data[packer.nbyte] <<= left
			packer.data[packer.nbyte] |= byte((code >> (nbit - left)) & uint64(uint32(1<<left)-1))
			packer.nbit = 0
			packer.nbyte += 1
			nbit -= left
			code &= uint64((1 << nbit) - 1)
		}
	}
	return nil
}

func (unpacker *BitUnpacker) Left() uint32 {
	return unpacker.size - (unpacker.pbyte*8 + unpacker.pbit)
}

func (unpacker *BitUnpacker) ReadBits(nbit uint32) (uint64, error) {
	if unpacker.Left() < nbit {
		return 0, fmt.Errorf("Not enough bits left!")
	}
	ret := uint64(0)
	for nbit > 0 {
		left := 8 - unpacker.pbit
		if nbit < left {
			ret <<= nbit
			ret |= uint64((unpacker.data[unpacker.pbyte] >> (left - nbit)) & byte((1<<nbit)-1))
			unpacker.pbit += nbit
			return ret, nil
		} else {
			ret <<= left
			ret |= uint64(unpacker.data[unpacker.pbyte] & byte((1<<left)-1))
			unpacker.pbyte += 1
			unpacker.pbit = 0
			nbit -= left
		}
	}
	return ret, nil
}

func (packer *BitPacker) Size() uint32 {
	return packer.nbyte*8 + packer.nbit
}

func (packer *BitPacker) Data() []byte {
	return packer.data[:(packer.Size()+7)/8]
}
