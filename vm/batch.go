package vm

import (
	"time"

	"github.com/iost-official/Go-IOS-Protocol/core/block"
	"github.com/iost-official/Go-IOS-Protocol/core/tx"
	"github.com/iost-official/Go-IOS-Protocol/vm/database"
)

type Batch struct {
	Txs      []*tx.Tx
	Receipts []*tx.TxReceipt
}

type TxSender interface {
	Tx() *tx.Tx
}

type Maker struct {
	sender TxSender
}

func (m *Maker) Batch(bh *block.BlockHead, db database.IMultiValue, limit time.Duration, thread int) *Batch {
	vis := make([]*database.Visitor, 0)
	mappers := make([]database.Mapper, 0)
	engines := make([]engineImpl, 0)
	bvr := database.NewBatchVisitorRoot(10000, db)
	for i := 0; i < thread; i++ {
		vi, mapper := database.NewBatchVisitor(bvr)
		vis = append(vis, vi)
		mappers = append(mappers, mapper)

		e := newEngine(bh, vi)

		// todo setup engine
		e.Exec()
	}

}
