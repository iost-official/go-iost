package protocol

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

type SignedBlock struct {
	Sig         []byte
	Pubkey      []byte
	Blk         []byte
	BlkHeadHash []byte
}

func (d *SignedBlock) Size() (s uint64) {

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
	{
		l := uint64(len(d.Blk))

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
		l := uint64(len(d.BlkHeadHash))

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
	return
}
func (d *SignedBlock) Marshal(buf []byte) ([]byte, error) {
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
		l := uint64(len(d.Sig))

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
		copy(buf[i+0:], d.Sig)
		i += l
	}
	{
		l := uint64(len(d.Pubkey))

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
		copy(buf[i+0:], d.Pubkey)
		i += l
	}
	{
		l := uint64(len(d.Blk))

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
		copy(buf[i+0:], d.Blk)
		i += l
	}
	{
		l := uint64(len(d.BlkHeadHash))

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
		copy(buf[i+0:], d.BlkHeadHash)
		i += l
	}
	return buf[:i+0], nil
}

func (d *SignedBlock) Unmarshal(buf []byte) (uint64, error) {
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
		if uint64(cap(d.Sig)) >= l {
			d.Sig = d.Sig[:l]
		} else {
			d.Sig = make([]byte, l)
		}
		copy(d.Sig, buf[i+0:])
		i += l
	}
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
		if uint64(cap(d.Pubkey)) >= l {
			d.Pubkey = d.Pubkey[:l]
		} else {
			d.Pubkey = make([]byte, l)
		}
		copy(d.Pubkey, buf[i+0:])
		i += l
	}
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
		if uint64(cap(d.Blk)) >= l {
			d.Blk = d.Blk[:l]
		} else {
			d.Blk = make([]byte, l)
		}
		copy(d.Blk, buf[i+0:])
		i += l
	}
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
		if uint64(cap(d.BlkHeadHash)) >= l {
			d.BlkHeadHash = d.BlkHeadHash[:l]
		} else {
			d.BlkHeadHash = make([]byte, l)
		}
		copy(d.BlkHeadHash, buf[i+0:])
		i += l
	}
	return i + 0, nil
}

type Prepare struct {
	Sig      []byte
	Pubkey   []byte
	Rand     []byte
	IsAccept bool
}

func (d *Prepare) Size() (s uint64) {

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
	{
		l := uint64(len(d.Rand))

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
func (d *Prepare) Marshal(buf []byte) ([]byte, error) {
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
		l := uint64(len(d.Sig))

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
		copy(buf[i+0:], d.Sig)
		i += l
	}
	{
		l := uint64(len(d.Pubkey))

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
		copy(buf[i+0:], d.Pubkey)
		i += l
	}
	{
		l := uint64(len(d.Rand))

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
		copy(buf[i+0:], d.Rand)
		i += l
	}
	{
		if d.IsAccept {
			buf[i+0] = 1
		} else {
			buf[i+0] = 0
		}
	}
	return buf[:i+1], nil
}

func (d *Prepare) Unmarshal(buf []byte) (uint64, error) {
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
		if uint64(cap(d.Sig)) >= l {
			d.Sig = d.Sig[:l]
		} else {
			d.Sig = make([]byte, l)
		}
		copy(d.Sig, buf[i+0:])
		i += l
	}
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
		if uint64(cap(d.Pubkey)) >= l {
			d.Pubkey = d.Pubkey[:l]
		} else {
			d.Pubkey = make([]byte, l)
		}
		copy(d.Pubkey, buf[i+0:])
		i += l
	}
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
		if uint64(cap(d.Rand)) >= l {
			d.Rand = d.Rand[:l]
		} else {
			d.Rand = make([]byte, l)
		}
		copy(d.Rand, buf[i+0:])
		i += l
	}
	{
		d.IsAccept = buf[i+0] == 1
	}
	return i + 1, nil
}

type Commit struct {
	Sig         []byte
	Pubkey      []byte
	BlkHeadHash []byte
}

func (d *Commit) Size() (s uint64) {

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
	{
		l := uint64(len(d.BlkHeadHash))

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
	return
}
func (d *Commit) Marshal(buf []byte) ([]byte, error) {
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
		l := uint64(len(d.Sig))

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
		copy(buf[i+0:], d.Sig)
		i += l
	}
	{
		l := uint64(len(d.Pubkey))

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
		copy(buf[i+0:], d.Pubkey)
		i += l
	}
	{
		l := uint64(len(d.BlkHeadHash))

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
		copy(buf[i+0:], d.BlkHeadHash)
		i += l
	}
	return buf[:i+0], nil
}

func (d *Commit) Unmarshal(buf []byte) (uint64, error) {
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
		if uint64(cap(d.Sig)) >= l {
			d.Sig = d.Sig[:l]
		} else {
			d.Sig = make([]byte, l)
		}
		copy(d.Sig, buf[i+0:])
		i += l
	}
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
		if uint64(cap(d.Pubkey)) >= l {
			d.Pubkey = d.Pubkey[:l]
		} else {
			d.Pubkey = make([]byte, l)
		}
		copy(d.Pubkey, buf[i+0:])
		i += l
	}
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
		if uint64(cap(d.BlkHeadHash)) >= l {
			d.BlkHeadHash = d.BlkHeadHash[:l]
		} else {
			d.BlkHeadHash = make([]byte, l)
		}
		copy(d.BlkHeadHash, buf[i+0:])
		i += l
	}
	return i + 0, nil
}
