package global

import (
	"github.com/iost-official/go-iost/chainbase"
	"github.com/iost-official/go-iost/common"
	"github.com/iost-official/go-iost/core/block"
	"github.com/iost-official/go-iost/db"
)

// BaseVariableImpl is the implementation of BaseVariable
type BaseVariableImpl struct {
	blockChain block.Chain
	stateDB    db.MVCCDB
	config     *common.Config
}

// New return a BaseVariable instance
func New(chainBase *chainbase.ChainBase, conf *common.Config) *BaseVariableImpl {
	return &BaseVariableImpl{
		blockChain: chainBase.BlockChain(),
		stateDB:    chainBase.StateDB(),
		config:     conf,
	}
}

// StateDB return the state database
func (g *BaseVariableImpl) StateDB() db.MVCCDB {
	return g.stateDB
}

// BlockChain return the block chain
func (g *BaseVariableImpl) BlockChain() block.Chain {
	return g.blockChain
}

// Config return the config
func (g *BaseVariableImpl) Config() *common.Config {
	return g.config
}
