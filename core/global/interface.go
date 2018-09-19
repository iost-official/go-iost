package global

import (
	"github.com/iost-official/Go-IOS-Protocol/common"
	"github.com/iost-official/Go-IOS-Protocol/core/block"
	"github.com/iost-official/Go-IOS-Protocol/db"
)

//go:generate mockgen -destination ../mocks/mock_global.go -package core_mock github.com/iost-official/Go-IOS-Protocol/core/global BaseVariable

// BaseVariable defines BaseVariable's API.
type BaseVariable interface {
	TxDB() TxDB
	StateDB() db.MVCCDB
	Config() *common.Config
	BlockChain() block.Chain
	WitnessList() []string
	Mode() TMode
	SetMode(m TMode)
}
