package batch

import "github.com/iost-official/Go-IOS-Protocol/core/tx"

type Batch struct {
	Txs      []*tx.Tx
	Receipts []*tx.TxReceipt
}

type TxSender interface {
	Tx() *tx.Tx
}
