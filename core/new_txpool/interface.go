package txpool

import (
	"github.com/iost-official/Go-IOS-Protocol/core/block"
	"github.com/iost-official/Go-IOS-Protocol/core/tx"
)

type TxPool interface {
	Start()
	Stop()
	AddConfirmBlock(block *block.Block, isLongestChain bool) error
	AddTx(tx *tx.Tx) error
	PendingTxs(maxCnt int) (tx.TransactionsList, error)
	PendingTxsNum() (int, error)
	ExistTxs(hash string, block *block.Block) (bool, error)
}
