package new_txpool

import (
	"github.com/iost-official/Go-IOS-Protocol/core/new_blockcache"
	"github.com/iost-official/Go-IOS-Protocol/core/tx"
)

type TxPool interface {
	Start()
	Stop()
	AddLinkBlock(newNode *blockcache.BlockCacheNode, longestChainNode *blockcache.BlockCacheNode) error
	AddTx(tx *tx.Tx) error
	PendingTxs(maxCnt int) (tx.TransactionsList, error)
	PendingTxsNum() (int, error)
	ExistTxs(hash string, chainNode *blockcache.BlockCacheNode) (FRet, error)
}
