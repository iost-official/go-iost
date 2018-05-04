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

type contractInfoRaw struct {
	Language string
	Version  int8
	GasLimit int64
	Price    float64
}

func (d *contractInfoRaw) Size() (s uint64) {

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
	s += 17
	return
}
func (d *contractInfoRaw) Marshal(buf []byte) ([]byte, error) {
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
	return buf[:i+17], nil
}

func (d *contractInfoRaw) Unmarshal(buf []byte) (uint64, error) {
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
	return i + 17, nil
}
