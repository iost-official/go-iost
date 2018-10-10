package global

import (
	"github.com/iost-official/go-iost/common"
	"github.com/iost-official/go-iost/core/block"
	"github.com/iost-official/go-iost/db"
)

//go:generate mockgen -destination ../mocks/mock_global.go -package core_mock github.com/iost-official/go-iost/core/global BaseVariable

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
