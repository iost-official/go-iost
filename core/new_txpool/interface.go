package new_txpool

import (
	"github.com/iost-official/Go-IOS-Protocol/core/message"
	"github.com/iost-official/Go-IOS-Protocol/core/new_blockcache"
	"github.com/iost-official/Go-IOS-Protocol/core/tx"
)

type TxPool interface {
	Start()
	Stop()
	AddLinkedNode(linkedNode *blockcache.BlockCacheNode, headNode *blockcache.BlockCacheNode) error
	AddTx(tx message.Message) error
	PendingTxs(maxCnt int) (tx.TransactionsList, error)
	ExistTxs(hash []byte, chainNode *blockcache.BlockCacheNode) (FRet, error)
}
