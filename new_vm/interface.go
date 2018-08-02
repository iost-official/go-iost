package new_vm

import (
	"github.com/iost-official/Go-IOS-Protocol/core/block"
	"github.com/iost-official/Go-IOS-Protocol/core/new_tx"
	"github.com/iost-official/Go-IOS-Protocol/db"
)

const (
	defaultCacheLength = 1000
)

type Engine interface {
	Init(cb *db.MVCCDB)
	SetEnv(bh *block.BlockHead, stateVersion string) Engine
	Exec(tx0 tx.Tx) (tx.TxReceipt, error)
	GC()
}

type EngineImpl struct {
	monitor *Monitor
}

func (e *EngineImpl) Init(cb *db.MVCCDB) {
	e.monitor = NewMonitor(cb, defaultCacheLength)
}
func (e *EngineImpl) SetEnv(bh *block.BlockHead, stateVersion string) Engine {

}
func (e *EngineImpl) Exec(tx0 tx.Tx) (tx.TxReceipt, error) {
	// prepare ctx
}
func (e *EngineImpl) GC() {

}
