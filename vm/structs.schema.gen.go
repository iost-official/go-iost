package vm

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

type ContractInfo struct {
	Name     string
	Language string
	Version  int8
	GasLimit int64
	Price    float64
	Signers  [][]byte
	ApiList  []string
}

func (d *ContractInfo) Size() (s uint64) {

	{
		l := uint64(len(d.Name))

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
		l := uint64(len(d.Language))

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
		l := uint64(len(d.Signers))

		{

			t := l
			for t >= 0x80 {
				t >>= 7
				s++
			}
			s++

		}

		for k0 := range d.Signers {

			{
				l := uint64(len(d.Signers[k0]))

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
		l := uint64(len(d.ApiList))

		{

			t := l
			for t >= 0x80 {
				t >>= 7
				s++
			}
			s++

		}

		for k0 := range d.ApiList {

			{
				l := uint64(len(d.ApiList[k0]))

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
	s += 17
	return
}
func (d *ContractInfo) Marshal(buf []byte) ([]byte, error) {
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
		l := uint64(len(d.Name))

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
		copy(buf[i+0:], d.Name)
		i += l
	}
	{
		l := uint64(len(d.Language))

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
		copy(buf[i+0:], d.Language)
		i += l
	}
	{

		buf[i+0+0] = byte(d.Version >> 0)

	}
	{

		buf[i+0+1] = byte(d.GasLimit >> 0)

		buf[i+1+1] = byte(d.GasLimit >> 8)

		buf[i+2+1] = byte(d.GasLimit >> 16)

		buf[i+3+1] = byte(d.GasLimit >> 24)

		buf[i+4+1] = byte(d.GasLimit >> 32)

		buf[i+5+1] = byte(d.GasLimit >> 40)

		buf[i+6+1] = byte(d.GasLimit >> 48)

		buf[i+7+1] = byte(d.GasLimit >> 56)

	}
	{

		v := *(*uint64)(unsafe.Pointer(&(d.Price)))

		buf[i+0+9] = byte(v >> 0)

		buf[i+1+9] = byte(v >> 8)

		buf[i+2+9] = byte(v >> 16)

		buf[i+3+9] = byte(v >> 24)

		buf[i+4+9] = byte(v >> 32)

		buf[i+5+9] = byte(v >> 40)

		buf[i+6+9] = byte(v >> 48)

		buf[i+7+9] = byte(v >> 56)

	}
	{
		l := uint64(len(d.Signers))

		{

			t := uint64(l)

			for t >= 0x80 {
				buf[i+17] = byte(t) | 0x80
				t >>= 7
				i++
			}
			buf[i+17] = byte(t)
			i++

		}
		for k0 := range d.Signers {

			{
				l := uint64(len(d.Signers[k0]))

				{

					t := uint64(l)

					for t >= 0x80 {
						buf[i+17] = byte(t) | 0x80
						t >>= 7
						i++
					}
					buf[i+17] = byte(t)
					i++

				}
				copy(buf[i+17:], d.Signers[k0])
				i += l
			}

		}
	}
	{
		l := uint64(len(d.ApiList))

		{

			t := uint64(l)

			for t >= 0x80 {
				buf[i+17] = byte(t) | 0x80
				t >>= 7
				i++
			}
			buf[i+17] = byte(t)
			i++

		}
		for k0 := range d.ApiList {

			{
				l := uint64(len(d.ApiList[k0]))

				{

					t := uint64(l)

					for t >= 0x80 {
						buf[i+17] = byte(t) | 0x80
						t >>= 7
						i++
					}
					buf[i+17] = byte(t)
					i++

				}
				copy(buf[i+17:], d.ApiList[k0])
				i += l
			}

		}
	}
	return buf[:i+17], nil
}

func (d *ContractInfo) Unmarshal(buf []byte) (uint64, error) {
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
		d.Name = string(buf[i+0 : i+0+l])
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
		d.Language = string(buf[i+0 : i+0+l])
		i += l
	}
	{

		d.Version = 0 | (int8(buf[i+0+0]) << 0)

	}
	{

		d.GasLimit = 0 | (int64(buf[i+0+1]) << 0) | (int64(buf[i+1+1]) << 8) | (int64(buf[i+2+1]) << 16) | (int64(buf[i+3+1]) << 24) | (int64(buf[i+4+1]) << 32) | (int64(buf[i+5+1]) << 40) | (int64(buf[i+6+1]) << 48) | (int64(buf[i+7+1]) << 56)

	}
	{

		v := 0 | (uint64(buf[i+0+9]) << 0) | (uint64(buf[i+1+9]) << 8) | (uint64(buf[i+2+9]) << 16) | (uint64(buf[i+3+9]) << 24) | (uint64(buf[i+4+9]) << 32) | (uint64(buf[i+5+9]) << 40) | (uint64(buf[i+6+9]) << 48) | (uint64(buf[i+7+9]) << 56)
		d.Price = *(*float64)(unsafe.Pointer(&v))

	}
	{
		l := uint64(0)

		{

			bs := uint8(7)
			t := uint64(buf[i+17] & 0x7F)
			for buf[i+17]&0x80 == 0x80 {
				i++
				t |= uint64(buf[i+17]&0x7F) << bs
				bs += 7
			}
			i++

			l = t

		}
		if uint64(cap(d.Signers)) >= l {
			d.Signers = d.Signers[:l]
		} else {
			d.Signers = make([][]byte, l)
		}
		for k0 := range d.Signers {

			{
				l := uint64(0)

				{

					bs := uint8(7)
					t := uint64(buf[i+17] & 0x7F)
					for buf[i+17]&0x80 == 0x80 {
						i++
						t |= uint64(buf[i+17]&0x7F) << bs
						bs += 7
					}
					i++

					l = t

				}
				if uint64(cap(d.Signers[k0])) >= l {
					d.Signers[k0] = d.Signers[k0][:l]
				} else {
					d.Signers[k0] = make([]byte, l)
				}
				copy(d.Signers[k0], buf[i+17:])
				i += l
			}

		}
	}
	{
		l := uint64(0)

		{

			bs := uint8(7)
			t := uint64(buf[i+17] & 0x7F)
			for buf[i+17]&0x80 == 0x80 {
				i++
				t |= uint64(buf[i+17]&0x7F) << bs
				bs += 7
			}
			i++

			l = t

		}
		if uint64(cap(d.ApiList)) >= l {
			d.ApiList = d.ApiList[:l]
		} else {
			d.ApiList = make([]string, l)
		}
		for k0 := range d.ApiList {

			{
				l := uint64(0)

				{

					bs := uint8(7)
					t := uint64(buf[i+17] & 0x7F)
					for buf[i+17]&0x80 == 0x80 {
						i++
						t |= uint64(buf[i+17]&0x7F) << bs
						bs += 7
					}
					i++

					l = t

				}
				d.ApiList[k0] = string(buf[i+17 : i+17+l])
				i += l
			}

		}
	}
	return i + 17, nil
}

type ContractRaw struct {
	info ContractInfo
	code []byte
}

func (d *ContractRaw) Size() (s uint64) {

	{
		s += d.info.Size()
	}
	{
		l := uint64(len(d.code))

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
func (d *ContractRaw) Marshal(buf []byte) ([]byte, error) {
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
		nbuf, err := d.info.Marshal(buf[0:])
		if err != nil {
			return nil, err
		}
		i += uint64(len(nbuf))
	}
	{
		l := uint64(len(d.code))

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
		copy(buf[i+0:], d.code)
		i += l
	}
	return buf[:i+0], nil
}

func (d *ContractRaw) Unmarshal(buf []byte) (uint64, error) {
	i := uint64(0)

	{
		ni, err := d.info.Unmarshal(buf[i+0:])
		if err != nil {
			return 0, err
		}
		i += ni
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
		if uint64(cap(d.code)) >= l {
			d.code = d.code[:l]
		} else {
			d.code = make([]byte, l)
		}
		copy(d.code, buf[i+0:])
		i += l
	}
	return i + 0, nil
}
