package global

import (
	"fmt"
	"os"
	"sync"

	"github.com/iost-official/go-iost/common"
	"github.com/iost-official/go-iost/consensus/snapshot"
	"github.com/iost-official/go-iost/core/block"
	"github.com/iost-official/go-iost/db"
	"github.com/iost-official/go-iost/ilog"
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
	blockChain    block.Chain
	stateDB       db.MVCCDB
	mode          TMode
	modeMutex     *sync.RWMutex
	continuousNum int
	config        *common.Config
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
		blockChain:    blockChain,
		stateDB:       stateDB,
		mode:          ModeInit,
		modeMutex:     new(sync.RWMutex),
		continuousNum: 6,
		config:        conf,
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

// Continuous return the number of continue blocks
func (g *BaseVariableImpl) Continuous() int {
	return g.continuousNum
}

// Mode return the mode
func (g *BaseVariableImpl) Mode() TMode {
	g.modeMutex.RLock()
	defer g.modeMutex.RUnlock()
	return g.mode
}

// SetMode is set the mode
func (g *BaseVariableImpl) SetMode(m TMode) {
	g.modeMutex.Lock()
	defer g.modeMutex.Unlock()
	g.mode = m
}
