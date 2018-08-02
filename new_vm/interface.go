package new_vm

import (
	"context"

	"sync"

	"github.com/iost-official/Go-IOS-Protocol/core/block"
	"github.com/iost-official/Go-IOS-Protocol/core/new_tx"
	"github.com/iost-official/Go-IOS-Protocol/new_vm/database"
)

const (
	defaultCacheLength = 1000
)

type Engine interface {
	Init(cb database.IMultiValue)
	SetEnv(bh *block.BlockHead, stateVersion string) Engine
	Exec(tx0 tx.Tx) (tx.TxReceipt, error)
	GC()
}

var staticMonitor *Monitor

var once sync.Once

type EngineImpl struct {
	ctx context.Context
}

func (e *EngineImpl) Init(cb database.IMultiValue) {
	if staticMonitor == nil {
		once.Do(func() {
			staticMonitor = NewMonitor(cb, defaultCacheLength)
		})
	}
}
func (e *EngineImpl) SetEnv(bh *block.BlockHead, commit string) Engine {
	return nil
}
func (e *EngineImpl) Exec(tx0 tx.Tx) (tx.TxReceipt, error) {
	// prepare ctx
	return tx.TxReceipt{}, nil
}
func (e *EngineImpl) GC() {

}
