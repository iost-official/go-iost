package global

import (
	"fmt"
	"time"

	"github.com/iost-official/Go-IOS-Protocol/common"
	"github.com/iost-official/Go-IOS-Protocol/core/new_block"
	"github.com/iost-official/Go-IOS-Protocol/core/new_tx"
	"github.com/iost-official/Go-IOS-Protocol/db"
)

type Mode struct {
	mode TMode
}

func (m *Mode) Mode() TMode {
	return m.mode
}

func (m *Mode) SetMode(i TMode) bool {
	m.mode = i
	return true
}

type TMode uint

const (
	ModeNormal TMode = iota
	ModeSync
	ModeProduce
)

type BaseVariableImpl struct {
	blockChain block.Chain
	stateDB    db.MVCCDB
	txDB       tx.TxDB

	mode   *Mode
	config *common.Config
}

func New(conf *common.Config) (*BaseVariableImpl, error) {
	block.LevelDBPath = conf.LdbPath
	blockChain, err := block.Instance()
	if err != nil {
		return nil, fmt.Errorf("new blockchain failed, stop the program. err: %v", err)
	}
	blk, err := blockChain.Top()
	if err != nil {
		t := time.Date(1970, 1, 1, 0, 0, 0, 0, time.UTC)
		blk = block.GenGenesis(common.GetTimestamp(t.Unix()).Slot)
	}
	if blk == nil {
		return nil, fmt.Errorf("new statedb failed, stop the program. err: %v", err)
	}

	stateDB, err := db.NewMVCCDB("StatePoolDB")
	if err != nil {
		return nil, fmt.Errorf("new statedb failed, stop the program. err: %v", err)
	}

	hash := stateDB.CurrentTag()
	if hash != nil {
	}

	tx.LdbPath = conf.LdbPath
	txDb := tx.TxDbInstance()
	if txDb == nil {
		return nil, fmt.Errorf("new txdb failed, stop the program.")
	}
	//TODO: check DB, state, txDB
	m := new(Mode)
	m.SetMode(ModeNormal)

	n := &BaseVariableImpl{txDB: txDb, config: conf, stateDB: stateDB, blockChain: blockChain, mode: m}
	return n, nil
}

func (g *BaseVariableImpl) TxDB() tx.TxDB {
	return g.txDB
}

func (g *BaseVariableImpl) StateDB() *db.MVCCDB {
	return g.stateDB
}

func (g *BaseVariableImpl) BlockChain() block.Chain {
	return g.blockChain
}

func (g *BaseVariableImpl) Config() *common.Config {
	return g.config
}

func (g *BaseVariableImpl) Mode() *Mode {
	return g.mode
}
