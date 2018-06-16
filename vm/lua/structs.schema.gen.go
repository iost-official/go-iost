package lua

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

type contractRaw struct {
	info    []byte
	code    []byte
	methods []methodRaw
}

func (d *contractRaw) Size() (s uint64) {

	{
		l := uint64(len(d.info))

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
	{
		l := uint64(len(d.methods))

		{

			t := l
			for t >= 0x80 {
				t >>= 7
				s++
			}
			s++

		}

		for k0 := range d.methods {

			{
				s += d.methods[k0].Size()
			}

		}

	}
	return
}
func (d *contractRaw) Marshal(buf []byte) ([]byte, error) {
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
		l := uint64(len(d.info))

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
		copy(buf[i+0:], d.info)
		i += l
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
	{
		l := uint64(len(d.methods))

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
		for k0 := range d.methods {

			{
				nbuf, err := d.methods[k0].Marshal(buf[i+0:])
				if err != nil {
					return nil, err
				}
				i += uint64(len(nbuf))
			}

		}
	}
	return buf[:i+0], nil
}

func (d *contractRaw) Unmarshal(buf []byte) (uint64, error) {
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
		if uint64(cap(d.info)) >= l {
			d.info = d.info[:l]
		} else {
			d.info = make([]byte, l)
		}
		copy(d.info, buf[i+0:])
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
		if uint64(cap(d.code)) >= l {
			d.code = d.code[:l]
		} else {
			d.code = make([]byte, l)
		}
		copy(d.code, buf[i+0:])
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
		if uint64(cap(d.methods)) >= l {
			d.methods = d.methods[:l]
		} else {
			d.methods = make([]methodRaw, l)
		}
		for k0 := range d.methods {

			{
				ni, err := d.methods[k0].Unmarshal(buf[i+0:])
				if err != nil {
					return 0, err
				}
				i += ni
			}

		}
	}
	return i + 0, nil
}

type methodRaw struct {
	name string
	ic   int32
	oc   int32
}

func (d *methodRaw) Size() (s uint64) {

	{
		l := uint64(len(d.name))

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
func (d *methodRaw) Marshal(buf []byte) ([]byte, error) {
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
		l := uint64(len(d.name))

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
		copy(buf[i+0:], d.name)
		i += l
	}
	{

		buf[i+0+0] = byte(d.ic >> 0)

		buf[i+1+0] = byte(d.ic >> 8)

		buf[i+2+0] = byte(d.ic >> 16)

		buf[i+3+0] = byte(d.ic >> 24)

	}
	{

		buf[i+0+4] = byte(d.oc >> 0)

		buf[i+1+4] = byte(d.oc >> 8)

		buf[i+2+4] = byte(d.oc >> 16)

		buf[i+3+4] = byte(d.oc >> 24)

	}
	return buf[:i+8], nil
}

func (d *methodRaw) Unmarshal(buf []byte) (uint64, error) {
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
		d.name = string(buf[i+0 : i+0+l])
		i += l
	}
	{

		d.ic = 0 | (int32(buf[i+0+0]) << 0) | (int32(buf[i+1+0]) << 8) | (int32(buf[i+2+0]) << 16) | (int32(buf[i+3+0]) << 24)

	}
	{

		d.oc = 0 | (int32(buf[i+0+4]) << 0) | (int32(buf[i+1+4]) << 8) | (int32(buf[i+2+4]) << 16) | (int32(buf[i+3+4]) << 24)

	}
	return i + 8, nil
}
