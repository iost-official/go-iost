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
	s += 12
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
		l := uint64(len(d.Body))

		{

			t := uint64(l)

			for t >= 0x80 {
				buf[i+12] = byte(t) | 0x80
				t >>= 7
				i++
			}
			buf[i+12] = byte(t)
			i++

		}
		copy(buf[i+12:], d.Body)
		i += l
	}
	return buf[:i+12], nil
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
		l := uint64(0)

		{

			bs := uint8(7)
			t := uint64(buf[i+12] & 0x7F)
			for buf[i+12]&0x80 == 0x80 {
				i++
				t |= uint64(buf[i+12]&0x7F) << bs
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
		copy(d.Body, buf[i+12:])
		i += l
	}
	return i + 12, nil
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
}

func (d *RequestBlock) Size() (s uint64) {

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
	return buf[:i+8], nil
}

func (d *RequestBlock) Unmarshal(buf []byte) (uint64, error) {
	i := uint64(0)

	{

		d.BlockNumber = 0 | (uint64(buf[0+0]) << 0) | (uint64(buf[1+0]) << 8) | (uint64(buf[2+0]) << 16) | (uint64(buf[3+0]) << 24) | (uint64(buf[4+0]) << 32) | (uint64(buf[5+0]) << 40) | (uint64(buf[6+0]) << 48) | (uint64(buf[7+0]) << 56)

	}
	return i + 8, nil
}
