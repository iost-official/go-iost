package common

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

type SignatureRaw struct {
	Algorithm int8
	Sig       []byte
	Pubkey    []byte
}

func (d *SignatureRaw) Size() (s uint64) {

	{
		l := uint64(len(d.Sig))

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
		l := uint64(len(d.Pubkey))

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
	s += 1
	return
}
func (d *SignatureRaw) Marshal(buf []byte) ([]byte, error) {
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

		buf[0+0] = byte(d.Algorithm >> 0)

	}
	{
		l := uint64(len(d.Sig))

		{

			t := uint64(l)

			for t >= 0x80 {
				buf[i+1] = byte(t) | 0x80
				t >>= 7
				i++
			}
			buf[i+1] = byte(t)
			i++

		}
		copy(buf[i+1:], d.Sig)
		i += l
	}
	{
		l := uint64(len(d.Pubkey))

		{

			t := uint64(l)

			for t >= 0x80 {
				buf[i+1] = byte(t) | 0x80
				t >>= 7
				i++
			}
			buf[i+1] = byte(t)
			i++

		}
		copy(buf[i+1:], d.Pubkey)
		i += l
	}
	return buf[:i+1], nil
}

func (d *SignatureRaw) Unmarshal(buf []byte) (uint64, error) {
	i := uint64(0)

	{

		d.Algorithm = 0 | (int8(buf[i+0+0]) << 0)

	}
	{
		l := uint64(0)

		{

			bs := uint8(7)
			t := uint64(buf[i+1] & 0x7F)
			for buf[i+1]&0x80 == 0x80 {
				i++
				t |= uint64(buf[i+1]&0x7F) << bs
				bs += 7
			}
			i++

			l = t

		}
		if uint64(cap(d.Sig)) >= l {
			d.Sig = d.Sig[:l]
		} else {
			d.Sig = make([]byte, l)
		}
		copy(d.Sig, buf[i+1:])
		i += l
	}
	{
		l := uint64(0)

		{

			bs := uint8(7)
			t := uint64(buf[i+1] & 0x7F)
			for buf[i+1]&0x80 == 0x80 {
				i++
				t |= uint64(buf[i+1]&0x7F) << bs
				bs += 7
			}
			i++

			l = t

		}
		if uint64(cap(d.Pubkey)) >= l {
			d.Pubkey = d.Pubkey[:l]
		} else {
			d.Pubkey = make([]byte, l)
		}
		copy(d.Pubkey, buf[i+1:])
		i += l
	}
	return i + 1, nil
}
