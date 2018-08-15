package txpool

import (
	"github.com/iost-official/Go-IOS-Protocol/core/new_block"
	"github.com/iost-official/Go-IOS-Protocol/core/new_blockcache"
	"github.com/iost-official/Go-IOS-Protocol/core/new_tx"
)

type TxPool interface {
	Start()
	Stop()
	AddLinkedNode(linkedNode *blockcache.BlockCacheNode, headNode *blockcache.BlockCacheNode) error
	AddTx(tx *tx.Tx) TAddTx
	PendingTxs(maxCnt int) (TxsList, error)
	ExistTxs(hash []byte, chainBlock *block.Block) (FRet, error)
}
