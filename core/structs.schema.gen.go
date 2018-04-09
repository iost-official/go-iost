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
	StateHash    []byte
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
		l := uint64(len(d.StateHash))

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
		l := uint64(len(d.StateHash))

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
		copy(buf[i+0:], d.StateHash)
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
		if uint64(cap(d.StateHash)) >= l {
			d.StateHash = d.StateHash[:l]
		} else {
			d.StateHash = make([]byte, l)
		}
		copy(d.StateHash, buf[i+0:])
		i += l
	}
	return i + 0, nil
}

type Tx struct {
	Version  int32
	Recorder string
	Inputs   []TxInput
	Outputs  []UTXO
	Time     int64
}

func (d *Tx) Size() (s uint64) {

	{
		l := uint64(len(d.Recorder))

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
		l := uint64(len(d.Inputs))

		{

			t := l
			for t >= 0x80 {
				t >>= 7
				s++
			}
			s++

		}

		for k0 := range d.Inputs {

			{
				s += d.Inputs[k0].Size()
			}

		}

	}
	{
		l := uint64(len(d.Outputs))

		{

			t := l
			for t >= 0x80 {
				t >>= 7
				s++
			}
			s++

		}

		for k0 := range d.Outputs {

			{
				s += d.Outputs[k0].Size()
			}

		}

	}
	s += 12
	return
}
func (d *Tx) Marshal(buf []byte) ([]byte, error) {
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
		l := uint64(len(d.Recorder))

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
		copy(buf[i+4:], d.Recorder)
		i += l
	}
	{
		l := uint64(len(d.Inputs))

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
		for k0 := range d.Inputs {

			{
				nbuf, err := d.Inputs[k0].Marshal(buf[i+4:])
				if err != nil {
					return nil, err
				}
				i += uint64(len(nbuf))
			}

		}
	}
	{
		l := uint64(len(d.Outputs))

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
		for k0 := range d.Outputs {

			{
				nbuf, err := d.Outputs[k0].Marshal(buf[i+4:])
				if err != nil {
					return nil, err
				}
				i += uint64(len(nbuf))
			}

		}
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

func (d *Tx) Unmarshal(buf []byte) (uint64, error) {
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
		d.Recorder = string(buf[i+4 : i+4+l])
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
		if uint64(cap(d.Inputs)) >= l {
			d.Inputs = d.Inputs[:l]
		} else {
			d.Inputs = make([]TxInput, l)
		}
		for k0 := range d.Inputs {

			{
				ni, err := d.Inputs[k0].Unmarshal(buf[i+4:])
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
			t := uint64(buf[i+4] & 0x7F)
			for buf[i+4]&0x80 == 0x80 {
				i++
				t |= uint64(buf[i+4]&0x7F) << bs
				bs += 7
			}
			i++

			l = t

		}
		if uint64(cap(d.Outputs)) >= l {
			d.Outputs = d.Outputs[:l]
		} else {
			d.Outputs = make([]UTXO, l)
		}
		for k0 := range d.Outputs {

			{
				ni, err := d.Outputs[k0].Unmarshal(buf[i+4:])
				if err != nil {
					return 0, err
				}
				i += ni
			}

		}
	}
	{

		d.Time = 0 | (int64(buf[i+0+4]) << 0) | (int64(buf[i+1+4]) << 8) | (int64(buf[i+2+4]) << 16) | (int64(buf[i+3+4]) << 24) | (int64(buf[i+4+4]) << 32) | (int64(buf[i+5+4]) << 40) | (int64(buf[i+6+4]) << 48) | (int64(buf[i+7+4]) << 56)

	}
	return i + 12, nil
}

type TxPoolRaw struct {
	Txs    []Tx
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
			d.Txs = make([]Tx, l)
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
	Version   int8
	SuperHash []byte
	TreeHash  []byte
	Time      int64
}

func (d *BlockHead) Size() (s uint64) {

	{
		l := uint64(len(d.SuperHash))

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
	s += 9
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

	}
	{
		l := uint64(len(d.SuperHash))

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
		copy(buf[i+1:], d.SuperHash)
		i += l
	}
	{
		l := uint64(len(d.TreeHash))

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
		copy(buf[i+1:], d.TreeHash)
		i += l
	}
	{

		buf[i+0+1] = byte(d.Time >> 0)

		buf[i+1+1] = byte(d.Time >> 8)

		buf[i+2+1] = byte(d.Time >> 16)

		buf[i+3+1] = byte(d.Time >> 24)

		buf[i+4+1] = byte(d.Time >> 32)

		buf[i+5+1] = byte(d.Time >> 40)

		buf[i+6+1] = byte(d.Time >> 48)

		buf[i+7+1] = byte(d.Time >> 56)

	}
	return buf[:i+9], nil
}

func (d *BlockHead) Unmarshal(buf []byte) (uint64, error) {
	i := uint64(0)

	{

		d.Version = 0 | (int8(buf[i+0+0]) << 0)

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
		if uint64(cap(d.SuperHash)) >= l {
			d.SuperHash = d.SuperHash[:l]
		} else {
			d.SuperHash = make([]byte, l)
		}
		copy(d.SuperHash, buf[i+1:])
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
		if uint64(cap(d.TreeHash)) >= l {
			d.TreeHash = d.TreeHash[:l]
		} else {
			d.TreeHash = make([]byte, l)
		}
		copy(d.TreeHash, buf[i+1:])
		i += l
	}
	{

		d.Time = 0 | (int64(buf[i+0+1]) << 0) | (int64(buf[i+1+1]) << 8) | (int64(buf[i+2+1]) << 16) | (int64(buf[i+3+1]) << 24) | (int64(buf[i+4+1]) << 32) | (int64(buf[i+5+1]) << 40) | (int64(buf[i+6+1]) << 48) | (int64(buf[i+7+1]) << 56)

	}
	return i + 9, nil
}

type Block struct {
	Version   int32
	SuperHash []byte
	Head      BlockHead
	Content   []byte
}

func (d *Block) Size() (s uint64) {

	{
		l := uint64(len(d.SuperHash))

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
		l := uint64(len(d.SuperHash))

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
		copy(buf[i+4:], d.SuperHash)
		i += l
	}
	{
		nbuf, err := d.Head.Marshal(buf[i+4:])
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
		if uint64(cap(d.SuperHash)) >= l {
			d.SuperHash = d.SuperHash[:l]
		} else {
			d.SuperHash = make([]byte, l)
		}
		copy(d.SuperHash, buf[i+4:])
		i += l
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
