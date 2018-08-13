package ilog

import (
	"bytes"
	"sync"
)

type BufPool struct {
	pool *sync.Pool
}

func NewBufPool() *BufPool {
	p := &sync.Pool{
		New: func() interface{} {
			return new(bytes.Buffer)
		},
	}
	return &BufPool{pool: p}
}

func (bp *BufPool) Get() *bytes.Buffer {
	return bp.pool.Get().(*bytes.Buffer)
}

func (bp *BufPool) Release(buf *bytes.Buffer) {
	buf.Reset()
	bp.pool.Put(buf)
}
