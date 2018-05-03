package core

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

type UTXO struct {
	BirthTxHash []byte
	Value       int64
	Script      string
}

func (d *UTXO) Size() (s uint64) {

	{
		l := uint64(len(d.BirthTxHash))

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
		l := uint64(len(d.Script))

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
func (d *UTXO) Marshal(buf []byte) ([]byte, error) {
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
		l := uint64(len(d.BirthTxHash))

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
		copy(buf[i+0:], d.BirthTxHash)
		i += l
	}
	{

		buf[i+0+0] = byte(d.Value >> 0)

		buf[i+1+0] = byte(d.Value >> 8)

		buf[i+2+0] = byte(d.Value >> 16)

		buf[i+3+0] = byte(d.Value >> 24)

		buf[i+4+0] = byte(d.Value >> 32)

		buf[i+5+0] = byte(d.Value >> 40)

		buf[i+6+0] = byte(d.Value >> 48)

		buf[i+7+0] = byte(d.Value >> 56)

	}
	{
		l := uint64(len(d.Script))

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
		copy(buf[i+8:], d.Script)
		i += l
	}
	return buf[:i+8], nil
}

func (d *UTXO) Unmarshal(buf []byte) (uint64, error) {
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
		if uint64(cap(d.BirthTxHash)) >= l {
			d.BirthTxHash = d.BirthTxHash[:l]
		} else {
			d.BirthTxHash = make([]byte, l)
		}
		copy(d.BirthTxHash, buf[i+0:])
		i += l
	}
	{

		d.Value = 0 | (int64(buf[i+0+0]) << 0) | (int64(buf[i+1+0]) << 8) | (int64(buf[i+2+0]) << 16) | (int64(buf[i+3+0]) << 24) | (int64(buf[i+4+0]) << 32) | (int64(buf[i+5+0]) << 40) | (int64(buf[i+6+0]) << 48) | (int64(buf[i+7+0]) << 56)

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
		d.Script = string(buf[i+8 : i+8+l])
		i += l
	}
	return i + 8, nil
}

type TxInput struct {
	TxHash       []byte
	UnlockScript string
	UTXOHash     []byte
}

func (d *TxInput) Size() (s uint64) {

	{
		l := uint64(len(d.TxHash))

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
		l := uint64(len(d.UnlockScript))

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
		l := uint64(len(d.UTXOHash))

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
func (d *TxInput) Marshal(buf []byte) ([]byte, error) {
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
		l := uint64(len(d.TxHash))

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
		copy(buf[i+0:], d.TxHash)
		i += l
	}
	{
		l := uint64(len(d.UnlockScript))

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
		copy(buf[i+0:], d.UnlockScript)
		i += l
	}
	{
		l := uint64(len(d.UTXOHash))

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
		copy(buf[i+0:], d.UTXOHash)
		i += l
	}
	return buf[:i+0], nil
}

func (d *TxInput) Unmarshal(buf []byte) (uint64, error) {
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
		if uint64(cap(d.TxHash)) >= l {
			d.TxHash = d.TxHash[:l]
		} else {
			d.TxHash = make([]byte, l)
		}
		copy(d.TxHash, buf[i+0:])
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
		d.UnlockScript = string(buf[i+0 : i+0+l])
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
		if uint64(cap(d.UTXOHash)) >= l {
			d.UTXOHash = d.UTXOHash[:l]
		} else {
			d.UTXOHash = make([]byte, l)
		}
		copy(d.UTXOHash, buf[i+0:])
		i += l
	}
	return i + 0, nil
}

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

type TxPoolRaw struct {
	Txs    []TxRaw
	TxHash [][]byte
}

func (d *TxPoolRaw) Size() (s uint64) {

	{
		l := uint64(len(d.Txs))

		{

			t := l
			for t >= 0x80 {
				t >>= 7
				s++
			}
			s++

		}

		for k0 := range d.Txs {

			{
				s += d.Txs[k0].Size()
			}

		}

	}
	{
		l := uint64(len(d.TxHash))

		{

			t := l
			for t >= 0x80 {
				t >>= 7
				s++
			}
			s++

		}

		for k0 := range d.TxHash {

			{
				l := uint64(len(d.TxHash[k0]))

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
	return
}
func (d *TxPoolRaw) Marshal(buf []byte) ([]byte, error) {
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
		l := uint64(len(d.Txs))

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
		for k0 := range d.Txs {

			{
				nbuf, err := d.Txs[k0].Marshal(buf[i+0:])
				if err != nil {
					return nil, err
				}
				i += uint64(len(nbuf))
			}

		}
	}
	{
		l := uint64(len(d.TxHash))

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
		for k0 := range d.TxHash {

			{
				l := uint64(len(d.TxHash[k0]))

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
				copy(buf[i+0:], d.TxHash[k0])
				i += l
			}

		}
	}
	return buf[:i+0], nil
}

func (d *TxPoolRaw) Unmarshal(buf []byte) (uint64, error) {
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
		if uint64(cap(d.Txs)) >= l {
			d.Txs = d.Txs[:l]
		} else {
			d.Txs = make([]TxRaw, l)
		}
		for k0 := range d.Txs {

			{
				ni, err := d.Txs[k0].Unmarshal(buf[i+0:])
				if err != nil {
					return 0, err
				}
				i += ni
			}

		}
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
		if uint64(cap(d.TxHash)) >= l {
			d.TxHash = d.TxHash[:l]
		} else {
			d.TxHash = make([][]byte, l)
		}
		for k0 := range d.TxHash {

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
				if uint64(cap(d.TxHash[k0])) >= l {
					d.TxHash[k0] = d.TxHash[k0][:l]
				} else {
					d.TxHash[k0] = make([]byte, l)
				}
				copy(d.TxHash[k0], buf[i+0:])
				i += l
			}

		}
	}
	return i + 0, nil
}

type BlockHead struct {
	Version    int32
	ParentHash []byte
	TreeHash   []byte
	Info       []byte
	Time       int64
}

func (d *BlockHead) Size() (s uint64) {

	{
		l := uint64(len(d.ParentHash))

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
		l := uint64(len(d.TreeHash))

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
		l := uint64(len(d.Info))

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
func (d *BlockHead) Marshal(buf []byte) ([]byte, error) {
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

		buf[0+0] = byte(d.Version >> 0)

		buf[1+0] = byte(d.Version >> 8)

		buf[2+0] = byte(d.Version >> 16)

		buf[3+0] = byte(d.Version >> 24)

	}
	{
		l := uint64(len(d.ParentHash))

		{

			t := uint64(l)

			for t >= 0x80 {
				buf[i+4] = byte(t) | 0x80
				t >>= 7
				i++
			}
			buf[i+4] = byte(t)
			i++

		}
		copy(buf[i+4:], d.ParentHash)
		i += l
	}
	{
		l := uint64(len(d.TreeHash))

		{

			t := uint64(l)

			for t >= 0x80 {
				buf[i+4] = byte(t) | 0x80
				t >>= 7
				i++
			}
			buf[i+4] = byte(t)
			i++

		}
		copy(buf[i+4:], d.TreeHash)
		i += l
	}
	{
		l := uint64(len(d.Info))

		{

			t := uint64(l)

			for t >= 0x80 {
				buf[i+4] = byte(t) | 0x80
				t >>= 7
				i++
			}
			buf[i+4] = byte(t)
			i++

		}
		copy(buf[i+4:], d.Info)
		i += l
	}
	{

		buf[i+0+4] = byte(d.Time >> 0)

		buf[i+1+4] = byte(d.Time >> 8)

		buf[i+2+4] = byte(d.Time >> 16)

		buf[i+3+4] = byte(d.Time >> 24)

		buf[i+4+4] = byte(d.Time >> 32)

		buf[i+5+4] = byte(d.Time >> 40)

		buf[i+6+4] = byte(d.Time >> 48)

		buf[i+7+4] = byte(d.Time >> 56)

	}
	return buf[:i+12], nil
}

func (d *BlockHead) Unmarshal(buf []byte) (uint64, error) {
	i := uint64(0)

	{

		d.Version = 0 | (int32(buf[i+0+0]) << 0) | (int32(buf[i+1+0]) << 8) | (int32(buf[i+2+0]) << 16) | (int32(buf[i+3+0]) << 24)

	}
	{
		l := uint64(0)

		{

			bs := uint8(7)
			t := uint64(buf[i+4] & 0x7F)
			for buf[i+4]&0x80 == 0x80 {
				i++
				t |= uint64(buf[i+4]&0x7F) << bs
				bs += 7
			}
			i++

			l = t

		}
		if uint64(cap(d.ParentHash)) >= l {
			d.ParentHash = d.ParentHash[:l]
		} else {
			d.ParentHash = make([]byte, l)
		}
		copy(d.ParentHash, buf[i+4:])
		i += l
	}
	{
		l := uint64(0)

		{

			bs := uint8(7)
			t := uint64(buf[i+4] & 0x7F)
			for buf[i+4]&0x80 == 0x80 {
				i++
				t |= uint64(buf[i+4]&0x7F) << bs
				bs += 7
			}
			i++

			l = t

		}
		if uint64(cap(d.TreeHash)) >= l {
			d.TreeHash = d.TreeHash[:l]
		} else {
			d.TreeHash = make([]byte, l)
		}
		copy(d.TreeHash, buf[i+4:])
		i += l
	}
	{
		l := uint64(0)

		{

			bs := uint8(7)
			t := uint64(buf[i+4] & 0x7F)
			for buf[i+4]&0x80 == 0x80 {
				i++
				t |= uint64(buf[i+4]&0x7F) << bs
				bs += 7
			}
			i++

			l = t

		}
		if uint64(cap(d.Info)) >= l {
			d.Info = d.Info[:l]
		} else {
			d.Info = make([]byte, l)
		}
		copy(d.Info, buf[i+4:])
		i += l
	}
	{

		d.Time = 0 | (int64(buf[i+0+4]) << 0) | (int64(buf[i+1+4]) << 8) | (int64(buf[i+2+4]) << 16) | (int64(buf[i+3+4]) << 24) | (int64(buf[i+4+4]) << 32) | (int64(buf[i+5+4]) << 40) | (int64(buf[i+6+4]) << 48) | (int64(buf[i+7+4]) << 56)

	}
	return i + 12, nil
}

type Block struct {
	Version int32
	Head    BlockHead
	Content []byte
}

func (d *Block) Size() (s uint64) {

	{
		s += d.Head.Size()
	}
	{
		l := uint64(len(d.Content))

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
	s += 4
	return
}
func (d *Block) Marshal(buf []byte) ([]byte, error) {
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

		buf[0+0] = byte(d.Version >> 0)

		buf[1+0] = byte(d.Version >> 8)

		buf[2+0] = byte(d.Version >> 16)

		buf[3+0] = byte(d.Version >> 24)

	}
	{
		nbuf, err := d.Head.Marshal(buf[4:])
		if err != nil {
			return nil, err
		}
		i += uint64(len(nbuf))
	}
	{
		l := uint64(len(d.Content))

		{

			t := uint64(l)

			for t >= 0x80 {
				buf[i+4] = byte(t) | 0x80
				t >>= 7
				i++
			}
			buf[i+4] = byte(t)
			i++

		}
		copy(buf[i+4:], d.Content)
		i += l
	}
	return buf[:i+4], nil
}

func (d *Block) Unmarshal(buf []byte) (uint64, error) {
	i := uint64(0)

	{

		d.Version = 0 | (int32(buf[i+0+0]) << 0) | (int32(buf[i+1+0]) << 8) | (int32(buf[i+2+0]) << 16) | (int32(buf[i+3+0]) << 24)

	}
	{
		ni, err := d.Head.Unmarshal(buf[i+4:])
		if err != nil {
			return 0, err
		}
		i += ni
	}
	{
		l := uint64(0)

		{

			bs := uint8(7)
			t := uint64(buf[i+4] & 0x7F)
			for buf[i+4]&0x80 == 0x80 {
				i++
				t |= uint64(buf[i+4]&0x7F) << bs
				bs += 7
			}
			i++

			l = t

		}
		if uint64(cap(d.Content)) >= l {
			d.Content = d.Content[:l]
		} else {
			d.Content = make([]byte, l)
		}
		copy(d.Content, buf[i+4:])
		i += l
	}
	return i + 4, nil
}

type Request struct {
	Time    int64  // 发送时的时间戳
	From    string // From To是钱包地址的base58编码字符串（就是Member.ID，下同）
	To      string
	ReqType int32 // 此request的类型码，通过类型可以确定body的格式以方便解码body
	Body    []byte
}

func (d *Request) Size() (s uint64) {

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
func (d *Request) Marshal(buf []byte) ([]byte, error) {
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

func (d *Request) Unmarshal(buf []byte) (uint64, error) {
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
