package ilog

import (
	"bytes"
	"sync"
)

// BufPool wraps sync.Pool to make it handy.
type BufPool struct {
	pool *sync.Pool
}

// NewBufPool returns a new instance of BufPool.
func NewBufPool() *BufPool {
	p := &sync.Pool{
		New: func() any {
			return new(bytes.Buffer)
		},
	}
	return &BufPool{pool: p}
}

// Get returns a bytes.Buffer from bufPool.
func (bp *BufPool) Get() *bytes.Buffer {
	return bp.pool.Get().(*bytes.Buffer)
}

// Release puts bytes.Buffer back to bufPool.
func (bp *BufPool) Release(buf *bytes.Buffer) {
	buf.Reset()
	bp.pool.Put(buf)
}
