package tx

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

type TxBaseRaw struct {
	Time     int64
	Nonce    int64
	Contract []byte
}

func (d *TxBaseRaw) Size() (s uint64) {

	{
		l := uint64(len(d.Contract))

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
	s += 16
	return
}
func (d *TxBaseRaw) Marshal(buf []byte) ([]byte, error) {
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

		buf[0+8] = byte(d.Nonce >> 0)

		buf[1+8] = byte(d.Nonce >> 8)

		buf[2+8] = byte(d.Nonce >> 16)

		buf[3+8] = byte(d.Nonce >> 24)

		buf[4+8] = byte(d.Nonce >> 32)

		buf[5+8] = byte(d.Nonce >> 40)

		buf[6+8] = byte(d.Nonce >> 48)

		buf[7+8] = byte(d.Nonce >> 56)

	}
	{
		l := uint64(len(d.Contract))

		{

			t := uint64(l)

			for t >= 0x80 {
				buf[i+16] = byte(t) | 0x80
				t >>= 7
				i++
			}
			buf[i+16] = byte(t)
			i++

		}
		copy(buf[i+16:], d.Contract)
		i += l
	}
	return buf[:i+16], nil
}

func (d *TxBaseRaw) Unmarshal(buf []byte) (uint64, error) {
	i := uint64(0)

	{

		d.Time = 0 | (int64(buf[i+0+0]) << 0) | (int64(buf[i+1+0]) << 8) | (int64(buf[i+2+0]) << 16) | (int64(buf[i+3+0]) << 24) | (int64(buf[i+4+0]) << 32) | (int64(buf[i+5+0]) << 40) | (int64(buf[i+6+0]) << 48) | (int64(buf[i+7+0]) << 56)

	}
	{

		d.Nonce = 0 | (int64(buf[i+0+8]) << 0) | (int64(buf[i+1+8]) << 8) | (int64(buf[i+2+8]) << 16) | (int64(buf[i+3+8]) << 24) | (int64(buf[i+4+8]) << 32) | (int64(buf[i+5+8]) << 40) | (int64(buf[i+6+8]) << 48) | (int64(buf[i+7+8]) << 56)

	}
	{
		l := uint64(0)

		{

			bs := uint8(7)
			t := uint64(buf[i+16] & 0x7F)
			for buf[i+16]&0x80 == 0x80 {
				i++
				t |= uint64(buf[i+16]&0x7F) << bs
				bs += 7
			}
			i++

			l = t

		}
		if uint64(cap(d.Contract)) >= l {
			d.Contract = d.Contract[:l]
		} else {
			d.Contract = make([]byte, l)
		}
		copy(d.Contract, buf[i+16:])
		i += l
	}
	return i + 16, nil
}

type TxPublishRaw struct {
	Time     int64
	Nonce    int64
	Contract []byte
	Signs    [][]byte
}

func (d *TxPublishRaw) Size() (s uint64) {

	{
		l := uint64(len(d.Contract))

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
		l := uint64(len(d.Signs))

		{

			t := l
			for t >= 0x80 {
				t >>= 7
				s++
			}
			s++

		}

		for k0 := range d.Signs {

			{
				l := uint64(len(d.Signs[k0]))

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

		}

	}
	s += 16
	return
}
func (d *TxPublishRaw) Marshal(buf []byte) ([]byte, error) {
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

		buf[0+8] = byte(d.Nonce >> 0)

		buf[1+8] = byte(d.Nonce >> 8)

		buf[2+8] = byte(d.Nonce >> 16)

		buf[3+8] = byte(d.Nonce >> 24)

		buf[4+8] = byte(d.Nonce >> 32)

		buf[5+8] = byte(d.Nonce >> 40)

		buf[6+8] = byte(d.Nonce >> 48)

		buf[7+8] = byte(d.Nonce >> 56)

	}
	{
		l := uint64(len(d.Contract))

		{

			t := uint64(l)

			for t >= 0x80 {
				buf[i+16] = byte(t) | 0x80
				t >>= 7
				i++
			}
			buf[i+16] = byte(t)
			i++

		}
		copy(buf[i+16:], d.Contract)
		i += l
	}
	{
		l := uint64(len(d.Signs))

		{

			t := uint64(l)

			for t >= 0x80 {
				buf[i+16] = byte(t) | 0x80
				t >>= 7
				i++
			}
			buf[i+16] = byte(t)
			i++

		}
		for k0 := range d.Signs {

			{
				l := uint64(len(d.Signs[k0]))

				{

					t := uint64(l)

					for t >= 0x80 {
						buf[i+16] = byte(t) | 0x80
						t >>= 7
						i++
					}
					buf[i+16] = byte(t)
					i++

				}
				copy(buf[i+16:], d.Signs[k0])
				i += l
			}

		}
	}
	return buf[:i+16], nil
}

func (d *TxPublishRaw) Unmarshal(buf []byte) (uint64, error) {
	i := uint64(0)

	{

		d.Time = 0 | (int64(buf[i+0+0]) << 0) | (int64(buf[i+1+0]) << 8) | (int64(buf[i+2+0]) << 16) | (int64(buf[i+3+0]) << 24) | (int64(buf[i+4+0]) << 32) | (int64(buf[i+5+0]) << 40) | (int64(buf[i+6+0]) << 48) | (int64(buf[i+7+0]) << 56)

	}
	{

		d.Nonce = 0 | (int64(buf[i+0+8]) << 0) | (int64(buf[i+1+8]) << 8) | (int64(buf[i+2+8]) << 16) | (int64(buf[i+3+8]) << 24) | (int64(buf[i+4+8]) << 32) | (int64(buf[i+5+8]) << 40) | (int64(buf[i+6+8]) << 48) | (int64(buf[i+7+8]) << 56)

	}
	{
		l := uint64(0)

		{

			bs := uint8(7)
			t := uint64(buf[i+16] & 0x7F)
			for buf[i+16]&0x80 == 0x80 {
				i++
				t |= uint64(buf[i+16]&0x7F) << bs
				bs += 7
			}
			i++

			l = t

		}
		if uint64(cap(d.Contract)) >= l {
			d.Contract = d.Contract[:l]
		} else {
			d.Contract = make([]byte, l)
		}
		copy(d.Contract, buf[i+16:])
		i += l
	}
	{
		l := uint64(0)

		{

			bs := uint8(7)
			t := uint64(buf[i+16] & 0x7F)
			for buf[i+16]&0x80 == 0x80 {
				i++
				t |= uint64(buf[i+16]&0x7F) << bs
				bs += 7
			}
			i++

			l = t

		}
		if uint64(cap(d.Signs)) >= l {
			d.Signs = d.Signs[:l]
		} else {
			d.Signs = make([][]byte, l)
		}
		for k0 := range d.Signs {

			{
				l := uint64(0)

				{

					bs := uint8(7)
					t := uint64(buf[i+16] & 0x7F)
					for buf[i+16]&0x80 == 0x80 {
						i++
						t |= uint64(buf[i+16]&0x7F) << bs
						bs += 7
					}
					i++

					l = t

				}
				if uint64(cap(d.Signs[k0])) >= l {
					d.Signs[k0] = d.Signs[k0][:l]
				} else {
					d.Signs[k0] = make([]byte, l)
				}
				copy(d.Signs[k0], buf[i+16:])
				i += l
			}

		}
	}
	return i + 16, nil
}

type TxRaw struct {
	Time      int64
	Nonce     int64
	Contract  []byte
	Signs     [][]byte
	Publisher []byte
}

func (d *TxRaw) Size() (s uint64) {

	{
		l := uint64(len(d.Contract))

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
		l := uint64(len(d.Signs))

		{

			t := l
			for t >= 0x80 {
				t >>= 7
				s++
			}
			s++

		}

		for k0 := range d.Signs {

			{
				l := uint64(len(d.Signs[k0]))

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

		}

	}
	{
		l := uint64(len(d.Publisher))

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
	s += 16
	return
}
func (d *TxRaw) Marshal(buf []byte) ([]byte, error) {
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

		buf[0+8] = byte(d.Nonce >> 0)

		buf[1+8] = byte(d.Nonce >> 8)

		buf[2+8] = byte(d.Nonce >> 16)

		buf[3+8] = byte(d.Nonce >> 24)

		buf[4+8] = byte(d.Nonce >> 32)

		buf[5+8] = byte(d.Nonce >> 40)

		buf[6+8] = byte(d.Nonce >> 48)

		buf[7+8] = byte(d.Nonce >> 56)

	}
	{
		l := uint64(len(d.Contract))

		{

			t := uint64(l)

			for t >= 0x80 {
				buf[i+16] = byte(t) | 0x80
				t >>= 7
				i++
			}
			buf[i+16] = byte(t)
			i++

		}
		copy(buf[i+16:], d.Contract)
		i += l
	}
	{
		l := uint64(len(d.Signs))

		{

			t := uint64(l)

			for t >= 0x80 {
				buf[i+16] = byte(t) | 0x80
				t >>= 7
				i++
			}
			buf[i+16] = byte(t)
			i++

		}
		for k0 := range d.Signs {

			{
				l := uint64(len(d.Signs[k0]))

				{

					t := uint64(l)

					for t >= 0x80 {
						buf[i+16] = byte(t) | 0x80
						t >>= 7
						i++
					}
					buf[i+16] = byte(t)
					i++

				}
				copy(buf[i+16:], d.Signs[k0])
				i += l
			}

		}
	}
	{
		l := uint64(len(d.Publisher))

		{

			t := uint64(l)

			for t >= 0x80 {
				buf[i+16] = byte(t) | 0x80
				t >>= 7
				i++
			}
			buf[i+16] = byte(t)
			i++

		}
		copy(buf[i+16:], d.Publisher)
		i += l
	}
	return buf[:i+16], nil
}

func (d *TxRaw) Unmarshal(buf []byte) (uint64, error) {
	i := uint64(0)

	{

		d.Time = 0 | (int64(buf[i+0+0]) << 0) | (int64(buf[i+1+0]) << 8) | (int64(buf[i+2+0]) << 16) | (int64(buf[i+3+0]) << 24) | (int64(buf[i+4+0]) << 32) | (int64(buf[i+5+0]) << 40) | (int64(buf[i+6+0]) << 48) | (int64(buf[i+7+0]) << 56)

	}
	{

		d.Nonce = 0 | (int64(buf[i+0+8]) << 0) | (int64(buf[i+1+8]) << 8) | (int64(buf[i+2+8]) << 16) | (int64(buf[i+3+8]) << 24) | (int64(buf[i+4+8]) << 32) | (int64(buf[i+5+8]) << 40) | (int64(buf[i+6+8]) << 48) | (int64(buf[i+7+8]) << 56)

	}
	{
		l := uint64(0)

		{

			bs := uint8(7)
			t := uint64(buf[i+16] & 0x7F)
			for buf[i+16]&0x80 == 0x80 {
				i++
				t |= uint64(buf[i+16]&0x7F) << bs
				bs += 7
			}
			i++

			l = t

		}
		if uint64(cap(d.Contract)) >= l {
			d.Contract = d.Contract[:l]
		} else {
			d.Contract = make([]byte, l)
		}
		copy(d.Contract, buf[i+16:])
		i += l
	}
	{
		l := uint64(0)

		{

			bs := uint8(7)
			t := uint64(buf[i+16] & 0x7F)
			for buf[i+16]&0x80 == 0x80 {
				i++
				t |= uint64(buf[i+16]&0x7F) << bs
				bs += 7
			}
			i++

			l = t

		}
		if uint64(cap(d.Signs)) >= l {
			d.Signs = d.Signs[:l]
		} else {
			d.Signs = make([][]byte, l)
		}
		for k0 := range d.Signs {

			{
				l := uint64(0)

				{

					bs := uint8(7)
					t := uint64(buf[i+16] & 0x7F)
					for buf[i+16]&0x80 == 0x80 {
						i++
						t |= uint64(buf[i+16]&0x7F) << bs
						bs += 7
					}
					i++

					l = t

				}
				if uint64(cap(d.Signs[k0])) >= l {
					d.Signs[k0] = d.Signs[k0][:l]
				} else {
					d.Signs[k0] = make([]byte, l)
				}
				copy(d.Signs[k0], buf[i+16:])
				i += l
			}

		}
	}
	{
		l := uint64(0)

		{

			bs := uint8(7)
			t := uint64(buf[i+16] & 0x7F)
			for buf[i+16]&0x80 == 0x80 {
				i++
				t |= uint64(buf[i+16]&0x7F) << bs
				bs += 7
			}
			i++

			l = t

		}
		if uint64(cap(d.Publisher)) >= l {
			d.Publisher = d.Publisher[:l]
		} else {
			d.Publisher = make([]byte, l)
		}
		copy(d.Publisher, buf[i+16:])
		i += l
	}
	return i + 16, nil
}
