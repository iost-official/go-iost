package global

import (
	"github.com/iost-official/Go-IOS-Protocol/common"
	"github.com/iost-official/Go-IOS-Protocol/core/block"
	"github.com/iost-official/Go-IOS-Protocol/core/new_blockcache"
	"github.com/iost-official/Go-IOS-Protocol/core/state"
	"github.com/iost-official/Go-IOS-Protocol/core/tx"
)

type Global interface {
	TxPool() *txpool.TxPool

	TxDB() *tx.TxPoolDb

	StdPool() *state.Pool

	BlockChain() *block.Chain

	BlockCache() *blockcache.BlockCache

	Holder() *tx.Holder
	Config() *common.Config
}
