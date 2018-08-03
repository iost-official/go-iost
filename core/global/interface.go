package global

import (
	"github.com/iost-official/Go-IOS-Protocol/common"
	"github.com/iost-official/Go-IOS-Protocol/core/block"
	"github.com/iost-official/Go-IOS-Protocol/core/state"
	"github.com/iost-official/Go-IOS-Protocol/core/tx"
)

type Global interface {
	TxDB() tx.TxPool

	StdPool() state.Pool

	Config() *common.Config
	BlockChain() block.Chain
	Mode() Mode
	SetMode(mode Mode) bool
}
