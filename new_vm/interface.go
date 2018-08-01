package new_vm

import (
	"github.com/iost-official/Go-IOS-Protocol/core/new_blockcache"
	"github.com/iost-official/Go-IOS-Protocol/core/new_tx"
	"github.com/iost-official/Go-IOS-Protocol/db"
)

const (
	defaultCacheLength = 1000
)

type Engine struct {
	monitor *Monitor
}

func (e *Engine) Init(cb *db.MVCCDB) {
	e.monitor = NewMonitor(cb, defaultCacheLength)
}
func (e *Engine) Exec(bcn *blockcache.BlockCacheNode, tx0 tx.Tx) (tx.TxReceipt, error) {
	// prepare ctx
}
func (e *Engine) GC() {

}
