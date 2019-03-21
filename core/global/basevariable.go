package global

import (
	"fmt"
	"os"

	"github.com/iost-official/go-iost/common"
	"github.com/iost-official/go-iost/consensus/snapshot"
	"github.com/iost-official/go-iost/core/block"
	"github.com/iost-official/go-iost/db"
	"github.com/iost-official/go-iost/ilog"
)

// BaseVariableImpl is the implementation of BaseVariable
type BaseVariableImpl struct {
	blockChain block.Chain
	stateDB    db.MVCCDB
	config     *common.Config
}

// New return a BaseVariable instance
func New(conf *common.Config) (*BaseVariableImpl, error) {
	if conf.Snapshot.Enable {
		conf.Snapshot.Enable = false
		s, err := os.Stat(conf.DB.LdbPath + "BlockChainDB")
		if err == nil && s.IsDir() {
			ilog.Warnln("start iserver with the snapshot failed, blockchain db already has.")
		} else {
			s, err = os.Stat(conf.DB.LdbPath + "StateDB")
			if err == nil && s.IsDir() {
				ilog.Warnln("start iserver with the snapshot failed, state db already has.")
			} else {
				err = snapshot.FromSnapshot(conf)
				if err != nil {
					ilog.Fatalf("start iserver with the snapshot failed, err:%v", err)
				}
				conf.Snapshot.Enable = true
			}
		}
	}

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
