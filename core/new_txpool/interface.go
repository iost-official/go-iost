package new_txpool

import (
	"github.com/iost-official/Go-IOS-Protocol/core/message"
	"github.com/iost-official/Go-IOS-Protocol/core/new_block"
	"github.com/iost-official/Go-IOS-Protocol/core/new_blockcache"
)

type TxPool interface {
	Start()
	Stop()
	AddLinkedNode(linkedNode *blockcache.BlockCacheNode, headNode *blockcache.BlockCacheNode) error
	AddTx(tx message.Message) error
	PendingTxs(maxCnt int) (TxsList, error)
	ExistTxs(hash []byte, chainBlock *block.Block) (FRet, error)
}
