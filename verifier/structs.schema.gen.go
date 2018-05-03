package verifier

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

type VerifyLogRaw struct {
	List [][]int32
}

func (d *VerifyLogRaw) Size() (s uint64) {

	{
		l := uint64(len(d.List))

		{

			t := l
			for t >= 0x80 {
				t >>= 7
				s++
			}
			s++

		}

		for k0 := range d.List {

			{
				l := uint64(len(d.List[k0]))

				{

					t := l
					for t >= 0x80 {
						t >>= 7
						s++
					}
					s++

				}

				s += 4 * l

			}

		}

	}
	return
}
func (d *VerifyLogRaw) Marshal(buf []byte) ([]byte, error) {
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
		l := uint64(len(d.List))

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
		for k0 := range d.List {

			{
				l := uint64(len(d.List[k0]))

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
				for k1 := range d.List[k0] {

					{

						buf[i+0+0] = byte(d.List[k0][k1] >> 0)

						buf[i+1+0] = byte(d.List[k0][k1] >> 8)

						buf[i+2+0] = byte(d.List[k0][k1] >> 16)

						buf[i+3+0] = byte(d.List[k0][k1] >> 24)

					}

					i += 4

				}
			}

		}
	}
	return buf[:i+0], nil
}

func (d *VerifyLogRaw) Unmarshal(buf []byte) (uint64, error) {
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
		if uint64(cap(d.List)) >= l {
			d.List = d.List[:l]
		} else {
			d.List = make([][]int32, l)
		}
		for k0 := range d.List {

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
				if uint64(cap(d.List[k0])) >= l {
					d.List[k0] = d.List[k0][:l]
				} else {
					d.List[k0] = make([]int32, l)
				}
				for k1 := range d.List[k0] {

					{

						d.List[k0][k1] = 0 | (int32(buf[i+0+0]) << 0) | (int32(buf[i+1+0]) << 8) | (int32(buf[i+2+0]) << 16) | (int32(buf[i+3+0]) << 24)

					}

					i += 4

				}
			}

		}
	}
	return i + 0, nil
}
