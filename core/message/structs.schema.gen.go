package message

import (
	"io"
	"time"
	"unsafe"
)

var (
	_ = unsafe.Sizeof(0)
	_ = io.ReadFull
	_ = time.Now()
)

type Message struct {
	Time    int64
	From    string
	To      string
	ReqType int32
	TTL     int8
	Body    []byte
}

func (d *Message) Size() (s uint64) {

	{
		l := uint64(len(d.From))

		{

			t := l
			for t >= 0x80 {
				t >>= 7
				s++
			}
			s++

		}
		s += l
	}
	{
		l := uint64(len(d.To))

		{

			t := l
			for t >= 0x80 {
				t >>= 7
				s++
			}
			s++

		}
		s += l
	}
	{
		l := uint64(len(d.Body))

		{

			t := l
			for t >= 0x80 {
				t >>= 7
				s++
			}
			s++

		}
		s += l
	}
	s += 13
	return
}
func (d *Message) Marshal(buf []byte) ([]byte, error) {
	size := d.Size()
	{
		if uint64(cap(buf)) >= size {
			buf = buf[:size]
		} else {
			buf = make([]byte, size)
		}
	}
	i := uint64(0)

	{

		buf[0+0] = byte(d.Time >> 0)

		buf[1+0] = byte(d.Time >> 8)

		buf[2+0] = byte(d.Time >> 16)

		buf[3+0] = byte(d.Time >> 24)

		buf[4+0] = byte(d.Time >> 32)

		buf[5+0] = byte(d.Time >> 40)

		buf[6+0] = byte(d.Time >> 48)

		buf[7+0] = byte(d.Time >> 56)

	}
	{
		l := uint64(len(d.From))

		{

			t := uint64(l)

			for t >= 0x80 {
				buf[i+8] = byte(t) | 0x80
				t >>= 7
				i++
			}
			buf[i+8] = byte(t)
			i++

		}
		copy(buf[i+8:], d.From)
		i += l
	}
	{
		l := uint64(len(d.To))

		{

			t := uint64(l)

			for t >= 0x80 {
				buf[i+8] = byte(t) | 0x80
				t >>= 7
				i++
			}
			buf[i+8] = byte(t)
			i++

		}
		copy(buf[i+8:], d.To)
		i += l
	}
	{

		buf[i+0+8] = byte(d.ReqType >> 0)

		buf[i+1+8] = byte(d.ReqType >> 8)

		buf[i+2+8] = byte(d.ReqType >> 16)

		buf[i+3+8] = byte(d.ReqType >> 24)

	}
	{

		buf[i+0+12] = byte(d.TTL >> 0)

	}
	{
		l := uint64(len(d.Body))

		{

			t := uint64(l)

			for t >= 0x80 {
				buf[i+13] = byte(t) | 0x80
				t >>= 7
				i++
			}
			buf[i+13] = byte(t)
			i++

		}
		copy(buf[i+13:], d.Body)
		i += l
	}
	return buf[:i+13], nil
}

func (d *Message) Unmarshal(buf []byte) (uint64, error) {
	i := uint64(0)

	{

		d.Time = 0 | (int64(buf[i+0+0]) << 0) | (int64(buf[i+1+0]) << 8) | (int64(buf[i+2+0]) << 16) | (int64(buf[i+3+0]) << 24) | (int64(buf[i+4+0]) << 32) | (int64(buf[i+5+0]) << 40) | (int64(buf[i+6+0]) << 48) | (int64(buf[i+7+0]) << 56)

	}
	{
		l := uint64(0)

		{

			bs := uint8(7)
			t := uint64(buf[i+8] & 0x7F)
			for buf[i+8]&0x80 == 0x80 {
				i++
				t |= uint64(buf[i+8]&0x7F) << bs
				bs += 7
			}
			i++

			l = t

		}
		d.From = string(buf[i+8 : i+8+l])
		i += l
	}
	{
		l := uint64(0)

		{

			bs := uint8(7)
			t := uint64(buf[i+8] & 0x7F)
			for buf[i+8]&0x80 == 0x80 {
				i++
				t |= uint64(buf[i+8]&0x7F) << bs
				bs += 7
			}
			i++

			l = t

		}
		d.To = string(buf[i+8 : i+8+l])
		i += l
	}
	{

		d.ReqType = 0 | (int32(buf[i+0+8]) << 0) | (int32(buf[i+1+8]) << 8) | (int32(buf[i+2+8]) << 16) | (int32(buf[i+3+8]) << 24)

	}
	{

		d.TTL = 0 | (int8(buf[i+0+12]) << 0)

	}
	{
		l := uint64(0)

		{

			bs := uint8(7)
			t := uint64(buf[i+13] & 0x7F)
			for buf[i+13]&0x80 == 0x80 {
				i++
				t |= uint64(buf[i+13]&0x7F) << bs
				bs += 7
			}
			i++

			l = t

		}
		if uint64(cap(d.Body)) >= l {
			d.Body = d.Body[:l]
		} else {
			d.Body = make([]byte, l)
		}
		copy(d.Body, buf[i+13:])
		i += l
	}
	return i + 13, nil
}

type RequestHeight struct {
	LocalBlockHeight uint64
	NeedBlockHeight  uint64
}

func (d *RequestHeight) Size() (s uint64) {

	s += 16
	return
}
func (d *RequestHeight) Marshal(buf []byte) ([]byte, error) {
	size := d.Size()
	{
		if uint64(cap(buf)) >= size {
			buf = buf[:size]
		} else {
			buf = make([]byte, size)
		}
	}
	i := uint64(0)

	{

		buf[0+0] = byte(d.LocalBlockHeight >> 0)

		buf[1+0] = byte(d.LocalBlockHeight >> 8)

		buf[2+0] = byte(d.LocalBlockHeight >> 16)

		buf[3+0] = byte(d.LocalBlockHeight >> 24)

		buf[4+0] = byte(d.LocalBlockHeight >> 32)

		buf[5+0] = byte(d.LocalBlockHeight >> 40)

		buf[6+0] = byte(d.LocalBlockHeight >> 48)

		buf[7+0] = byte(d.LocalBlockHeight >> 56)

	}
	{

		buf[0+8] = byte(d.NeedBlockHeight >> 0)

		buf[1+8] = byte(d.NeedBlockHeight >> 8)

		buf[2+8] = byte(d.NeedBlockHeight >> 16)

		buf[3+8] = byte(d.NeedBlockHeight >> 24)

		buf[4+8] = byte(d.NeedBlockHeight >> 32)

		buf[5+8] = byte(d.NeedBlockHeight >> 40)

		buf[6+8] = byte(d.NeedBlockHeight >> 48)

		buf[7+8] = byte(d.NeedBlockHeight >> 56)

	}
	return buf[:i+16], nil
}

func (d *RequestHeight) Unmarshal(buf []byte) (uint64, error) {
	i := uint64(0)

	{

		d.LocalBlockHeight = 0 | (uint64(buf[0+0]) << 0) | (uint64(buf[1+0]) << 8) | (uint64(buf[2+0]) << 16) | (uint64(buf[3+0]) << 24) | (uint64(buf[4+0]) << 32) | (uint64(buf[5+0]) << 40) | (uint64(buf[6+0]) << 48) | (uint64(buf[7+0]) << 56)

	}
	{

		d.NeedBlockHeight = 0 | (uint64(buf[0+8]) << 0) | (uint64(buf[1+8]) << 8) | (uint64(buf[2+8]) << 16) | (uint64(buf[3+8]) << 24) | (uint64(buf[4+8]) << 32) | (uint64(buf[5+8]) << 40) | (uint64(buf[6+8]) << 48) | (uint64(buf[7+8]) << 56)

	}
	return i + 16, nil
}

type ResponseHeight struct {
	BlockHeight uint64
}

func (d *ResponseHeight) Size() (s uint64) {

	s += 8
	return
}
func (d *ResponseHeight) Marshal(buf []byte) ([]byte, error) {
	size := d.Size()
	{
		if uint64(cap(buf)) >= size {
			buf = buf[:size]
		} else {
			buf = make([]byte, size)
		}
	}
	i := uint64(0)

	{

		buf[0+0] = byte(d.BlockHeight >> 0)

		buf[1+0] = byte(d.BlockHeight >> 8)

		buf[2+0] = byte(d.BlockHeight >> 16)

		buf[3+0] = byte(d.BlockHeight >> 24)

		buf[4+0] = byte(d.BlockHeight >> 32)

		buf[5+0] = byte(d.BlockHeight >> 40)

		buf[6+0] = byte(d.BlockHeight >> 48)

		buf[7+0] = byte(d.BlockHeight >> 56)

	}
	return buf[:i+8], nil
}

func (d *ResponseHeight) Unmarshal(buf []byte) (uint64, error) {
	i := uint64(0)

	{

		d.BlockHeight = 0 | (uint64(buf[0+0]) << 0) | (uint64(buf[1+0]) << 8) | (uint64(buf[2+0]) << 16) | (uint64(buf[3+0]) << 24) | (uint64(buf[4+0]) << 32) | (uint64(buf[5+0]) << 40) | (uint64(buf[6+0]) << 48) | (uint64(buf[7+0]) << 56)

	}
	return i + 8, nil
}

type RequestBlock struct {
	BlockNumber uint64
	BlockHash   []byte
}

func (d *RequestBlock) Size() (s uint64) {

	{
		l := uint64(len(d.BlockHash))

		{

			t := l
			for t >= 0x80 {
				t >>= 7
				s++
			}
			s++

		}
		s += l
	}
	s += 8
	return
}
func (d *RequestBlock) Marshal(buf []byte) ([]byte, error) {
	size := d.Size()
	{
		if uint64(cap(buf)) >= size {
			buf = buf[:size]
		} else {
			buf = make([]byte, size)
		}
	}
	i := uint64(0)

	{

		buf[0+0] = byte(d.BlockNumber >> 0)

		buf[1+0] = byte(d.BlockNumber >> 8)

		buf[2+0] = byte(d.BlockNumber >> 16)

		buf[3+0] = byte(d.BlockNumber >> 24)

		buf[4+0] = byte(d.BlockNumber >> 32)

		buf[5+0] = byte(d.BlockNumber >> 40)

		buf[6+0] = byte(d.BlockNumber >> 48)

		buf[7+0] = byte(d.BlockNumber >> 56)

	}
	{
		l := uint64(len(d.BlockHash))

		{

			t := uint64(l)

			for t >= 0x80 {
				buf[i+8] = byte(t) | 0x80
				t >>= 7
				i++
			}
			buf[i+8] = byte(t)
			i++

		}
		copy(buf[i+8:], d.BlockHash)
		i += l
	}
	return buf[:i+8], nil
}

func (d *RequestBlock) Unmarshal(buf []byte) (uint64, error) {
	i := uint64(0)

	{

		d.BlockNumber = 0 | (uint64(buf[i+0+0]) << 0) | (uint64(buf[i+1+0]) << 8) | (uint64(buf[i+2+0]) << 16) | (uint64(buf[i+3+0]) << 24) | (uint64(buf[i+4+0]) << 32) | (uint64(buf[i+5+0]) << 40) | (uint64(buf[i+6+0]) << 48) | (uint64(buf[i+7+0]) << 56)

	}
	{
		l := uint64(0)

		{

			bs := uint8(7)
			t := uint64(buf[i+8] & 0x7F)
			for buf[i+8]&0x80 == 0x80 {
				i++
				t |= uint64(buf[i+8]&0x7F) << bs
				bs += 7
			}
			i++

			l = t

		}
		if uint64(cap(d.BlockHash)) >= l {
			d.BlockHash = d.BlockHash[:l]
		} else {
			d.BlockHash = make([]byte, l)
		}
		copy(d.BlockHash, buf[i+8:])
		i += l
	}
	return i + 8, nil
}

type BlockHashQuery struct {
	Start uint64
	End   uint64
}

func (d *BlockHashQuery) Size() (s uint64) {

	s += 16
	return
}
func (d *BlockHashQuery) Marshal(buf []byte) ([]byte, error) {
	size := d.Size()
	{
		if uint64(cap(buf)) >= size {
			buf = buf[:size]
		} else {
			buf = make([]byte, size)
		}
	}
	i := uint64(0)

	{

		buf[0+0] = byte(d.Start >> 0)

		buf[1+0] = byte(d.Start >> 8)

		buf[2+0] = byte(d.Start >> 16)

		buf[3+0] = byte(d.Start >> 24)

		buf[4+0] = byte(d.Start >> 32)

		buf[5+0] = byte(d.Start >> 40)

		buf[6+0] = byte(d.Start >> 48)

		buf[7+0] = byte(d.Start >> 56)

	}
	{

		buf[0+8] = byte(d.End >> 0)

		buf[1+8] = byte(d.End >> 8)

		buf[2+8] = byte(d.End >> 16)

		buf[3+8] = byte(d.End >> 24)

		buf[4+8] = byte(d.End >> 32)

		buf[5+8] = byte(d.End >> 40)

		buf[6+8] = byte(d.End >> 48)

		buf[7+8] = byte(d.End >> 56)

	}
	return buf[:i+16], nil
}

func (d *BlockHashQuery) Unmarshal(buf []byte) (uint64, error) {
	i := uint64(0)

	{

		d.Start = 0 | (uint64(buf[0+0]) << 0) | (uint64(buf[1+0]) << 8) | (uint64(buf[2+0]) << 16) | (uint64(buf[3+0]) << 24) | (uint64(buf[4+0]) << 32) | (uint64(buf[5+0]) << 40) | (uint64(buf[6+0]) << 48) | (uint64(buf[7+0]) << 56)

	}
	{

		d.End = 0 | (uint64(buf[0+8]) << 0) | (uint64(buf[1+8]) << 8) | (uint64(buf[2+8]) << 16) | (uint64(buf[3+8]) << 24) | (uint64(buf[4+8]) << 32) | (uint64(buf[5+8]) << 40) | (uint64(buf[6+8]) << 48) | (uint64(buf[7+8]) << 56)

	}
	return i + 16, nil
}

type BlockHash struct {
	Height uint64
	Hash   []byte
}

func (d *BlockHash) Size() (s uint64) {

	{
		l := uint64(len(d.Hash))

		{

			t := l
			for t >= 0x80 {
				t >>= 7
				s++
			}
			s++

		}
		s += l
	}
	s += 8
	return
}
func (d *BlockHash) Marshal(buf []byte) ([]byte, error) {
	size := d.Size()
	{
		if uint64(cap(buf)) >= size {
			buf = buf[:size]
		} else {
			buf = make([]byte, size)
		}
	}
	i := uint64(0)

	{

		buf[0+0] = byte(d.Height >> 0)

		buf[1+0] = byte(d.Height >> 8)

		buf[2+0] = byte(d.Height >> 16)

		buf[3+0] = byte(d.Height >> 24)

		buf[4+0] = byte(d.Height >> 32)

		buf[5+0] = byte(d.Height >> 40)

		buf[6+0] = byte(d.Height >> 48)

		buf[7+0] = byte(d.Height >> 56)

	}
	{
		l := uint64(len(d.Hash))

		{

			t := uint64(l)

			for t >= 0x80 {
				buf[i+8] = byte(t) | 0x80
				t >>= 7
				i++
			}
			buf[i+8] = byte(t)
			i++

		}
		copy(buf[i+8:], d.Hash)
		i += l
	}
	return buf[:i+8], nil
}

func (d *BlockHash) Unmarshal(buf []byte) (uint64, error) {
	i := uint64(0)

	{

		d.Height = 0 | (uint64(buf[i+0+0]) << 0) | (uint64(buf[i+1+0]) << 8) | (uint64(buf[i+2+0]) << 16) | (uint64(buf[i+3+0]) << 24) | (uint64(buf[i+4+0]) << 32) | (uint64(buf[i+5+0]) << 40) | (uint64(buf[i+6+0]) << 48) | (uint64(buf[i+7+0]) << 56)

	}
	{
		l := uint64(0)

		{

			bs := uint8(7)
			t := uint64(buf[i+8] & 0x7F)
			for buf[i+8]&0x80 == 0x80 {
				i++
				t |= uint64(buf[i+8]&0x7F) << bs
				bs += 7
			}
			i++

			l = t

		}
		if uint64(cap(d.Hash)) >= l {
			d.Hash = d.Hash[:l]
		} else {
			d.Hash = make([]byte, l)
		}
		copy(d.Hash, buf[i+8:])
		i += l
	}
	return i + 8, nil
}

type BlockHashResponse struct {
	BlockHashes []BlockHash
}

func (d *BlockHashResponse) Size() (s uint64) {

	{
		l := uint64(len(d.BlockHashes))

		{

			t := l
			for t >= 0x80 {
				t >>= 7
				s++
			}
			s++

		}

		for k0 := range d.BlockHashes {

			{
				s += d.BlockHashes[k0].Size()
			}

		}

	}
	return
}
func (d *BlockHashResponse) Marshal(buf []byte) ([]byte, error) {
	size := d.Size()
	{
		if uint64(cap(buf)) >= size {
			buf = buf[:size]
		} else {
			buf = make([]byte, size)
		}
	}
	i := uint64(0)

	{
		l := uint64(len(d.BlockHashes))

		{

			t := uint64(l)

			for t >= 0x80 {
				buf[i+0] = byte(t) | 0x80
				t >>= 7
				i++
			}
			buf[i+0] = byte(t)
			i++

		}
		for k0 := range d.BlockHashes {

			{
				nbuf, err := d.BlockHashes[k0].Marshal(buf[i+0:])
				if err != nil {
					return nil, err
				}
				i += uint64(len(nbuf))
			}

		}
	}
	return buf[:i+0], nil
}

func (d *BlockHashResponse) Unmarshal(buf []byte) (uint64, error) {
	i := uint64(0)

	{
		l := uint64(0)

		{

			bs := uint8(7)
			t := uint64(buf[i+0] & 0x7F)
			for buf[i+0]&0x80 == 0x80 {
				i++
				t |= uint64(buf[i+0]&0x7F) << bs
				bs += 7
			}
			i++

			l = t

		}
		if uint64(cap(d.BlockHashes)) >= l {
			d.BlockHashes = d.BlockHashes[:l]
		} else {
			d.BlockHashes = make([]BlockHash, l)
		}
		for k0 := range d.BlockHashes {

			{
				ni, err := d.BlockHashes[k0].Unmarshal(buf[i+0:])
				if err != nil {
					return 0, err
				}
				i += ni
			}

		}
	}
	return i + 0, nil
}
