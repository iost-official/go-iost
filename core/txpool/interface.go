package txpool

import (
	"github.com/iost-official/Go-IOS-Protocol/core/block"
	"github.com/iost-official/Go-IOS-Protocol/core/blockcache"
	"github.com/iost-official/Go-IOS-Protocol/core/tx"
)

//go:generate mockgen -destination mock/mock_txpool.go -package txpool_mock github.com/iost-official/Go-IOS-Protocol/core/txpool TxPool

// TxPool defines all the API of txpool package.
type TxPool interface {
	Start() error
	Stop()
	AddLinkedNode(linkedNode *blockcache.BlockCacheNode, headNode *blockcache.BlockCacheNode) error
	AddTx(tx *tx.Tx) TAddTx
	DelTx(hash []byte) error
	TxIterator() (*Iterator, *blockcache.BlockCacheNode)
	PendingTxs(maxCnt int) (TxsList, *blockcache.BlockCacheNode, error)
	ExistTxs(hash []byte, chainBlock *block.Block) (FRet, error)
	CheckTxs(txs []*tx.Tx, chainBlock *block.Block) (*tx.Tx, error)
	Lock()
	Release()
	TxTimeOut(tx *tx.Tx) bool
}
