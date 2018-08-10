package global

import (
	"github.com/iost-official/Go-IOS-Protocol/common"
	"github.com/iost-official/Go-IOS-Protocol/core/new_block"
	"github.com/iost-official/Go-IOS-Protocol/core/state"
	"github.com/iost-official/Go-IOS-Protocol/core/tx"
)

//go:generate mockgen -destination ../mocks/mock_global.go -package core_mock github.com/iost-official/Go-IOS-Protocol/core/global Global

type Global interface {
	TxDB() tx.TxPool

	StdPool() state.Pool

	Config() *common.Config
	BlockChain() block.Chain
	Mode() Mode
	SetMode(mode Mode) bool
}
