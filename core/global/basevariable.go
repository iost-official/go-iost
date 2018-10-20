package global

import (
	"fmt"

	"github.com/iost-official/go-iost/common"
	"github.com/iost-official/go-iost/core/block"
	"github.com/iost-official/go-iost/db"
)

// TMode type of mode
type TMode uint

const (
	// ModeNormal is normal mode
	ModeNormal TMode = iota
	// ModeSync is sync mode
	ModeSync
	// ModeInit init mode
	ModeInit
)

// BuildTime build time
var BuildTime string

// GitHash git hash
var GitHash string

// String return string of mode
func (m TMode) String() string {
	switch m {
	case ModeNormal:
		return "ModeNormal"
	case ModeSync:
		return "ModeSync"
	case ModeInit:
		return "ModeInit"
	default:
		return ""
	}
}

// BaseVariableImpl is the implementation of BaseVariable
type BaseVariableImpl struct {
	blockChain block.Chain
	stateDB    db.MVCCDB
	mode       TMode
	config     *common.Config
}

// New return a BaseVariable instance
// nolint: gocyclo
func New(conf *common.Config) (*BaseVariableImpl, error) {
	blockChain, err := block.NewBlockChain(conf.DB.LdbPath + "BlockChainDB")
	if err != nil {
		return nil, fmt.Errorf("new blockchain failed, stop the program. err: %v", err)
	}

	stateDB, err := db.NewMVCCDB(conf.DB.LdbPath + "StateDB")
	if err != nil {
		return nil, fmt.Errorf("new statedb failed, stop the program. err: %v", err)
	}

	return &BaseVariableImpl{
		blockChain: blockChain,
		stateDB:    stateDB,
		mode:       ModeInit,
		config:     conf,
	}, nil
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

// Mode return the mode
func (g *BaseVariableImpl) Mode() TMode {
	return g.mode
}

// SetMode is set the mode
func (g *BaseVariableImpl) SetMode(m TMode) {
	g.mode = m
}
