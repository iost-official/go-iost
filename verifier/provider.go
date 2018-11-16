package verifier

import (
	"github.com/iost-official/go-iost/core/tx"
	"github.com/iost-official/go-iost/core/txpool"
)

// ProviderImpl impl of provider
type ProviderImpl struct {
	cache    []*tx.Tx
	pool     *txpool.SortedTxMap
	iter     *txpool.Iterator
	droplist map[*tx.Tx]error
}

// NewProvider ...
func NewProvider(pool *txpool.SortedTxMap) *ProviderImpl {
	return &ProviderImpl{
		cache:    make([]*tx.Tx, 0),
		droplist: make(map[*tx.Tx]error),
		pool:     pool,
		iter:     pool.Iter(),
	}
}

// Tx get next tx
func (p *ProviderImpl) Tx() *tx.Tx {
	if len(p.cache) != 0 {
		t := p.cache[len(p.cache)-1]
		p.cache = p.cache[:len(p.cache)-1]
		return t
	}
	t, ok := p.iter.Next()
	if !ok {
		return nil
	}
	return t
}

// Return send tx to pool
func (p *ProviderImpl) Return(t *tx.Tx) {
	p.cache = append(p.cache, t)
}

// Drop drop bad tx
func (p *ProviderImpl) Drop(t *tx.Tx, err error) {
	p.droplist[t] = err
}

// List list tx and errors of drop txs
func (p *ProviderImpl) List() (a []*tx.Tx, b []error) {
	a = make([]*tx.Tx, 0)
	b = make([]error, 0)
	for k, v := range p.droplist {
		a = append(a, k)
		b = append(b, v)
	}
	return
}

// Drop drop bad tx
func (p *ProviderImpl) Close() {
	for t := range p.droplist {
		p.pool.Del(t.Hash())
	}
}
