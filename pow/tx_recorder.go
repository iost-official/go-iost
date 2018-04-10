package pow

import (
	"github.com/iost-official/prototype/core"
	"fmt"
)

const (
	SinglePoolSize = 100
)

type Recorder struct {
	pools []core.TxPool
}

func NewRecorder() *Recorder {
	txp := core.NewTxPool()
	rec := Recorder{
		pools: []core.TxPool{txp},
	}
	return &rec
}

func (r *Recorder) Add(tx core.Tx) error {
	last := len(r.pools) - 1
	if r.pools[last].Size() > 100 {
		txp := core.NewTxPool()
		r.pools = append(r.pools, txp)
		last ++
	}
	r.pools[last].Add(tx)
	return nil
}

func (r *Recorder) Find(txHash []byte) (core.Tx, error) {
	for _, p := range r.pools {
		tx, err := p.Find(txHash)
		if err == nil {
			return tx, err
		}
	}
	return core.Tx{}, fmt.Errorf("not found")
}

func (r *Recorder) Pop() core.TxPool {
	txp := r.pools[0]
	if len(r.pools) <= 1 {
		r.pools = []core.TxPool{core.NewTxPool()}
	} else {
		r.pools = r.pools[1:]
	}
	return txp
}