package global

import (
	"fmt"
	"github.com/iost-official/Go-IOS-Protocol/common"
	"github.com/iost-official/Go-IOS-Protocol/core/block"
	"github.com/iost-official/Go-IOS-Protocol/core/state"
	"github.com/iost-official/Go-IOS-Protocol/core/tx"
	"github.com/pkg/errors"
)

type Mode uint

const (
	ModeNormal Mode = iota
	ModeSync
)

type GlobalImpl struct {
	txDB tx.TxPool

	stdPool    state.Pool
	blockChain block.Chain

	config *common.Config

	mode Mode
}

func New(conf *common.Config) (Global, error) {

	txDb := tx.TxDbInstance()
	if txDb == nil {
		return nil, errors.New("TxDbInstance failed, stop the program!")
	}

	err := state.PoolInstance()
	if err != nil {
		return nil, fmt.Errorf("PoolInstance failed, stop the program! err:%v", err)
	}

	blockChain, err := block.Instance()
	if err != nil {
		return nil, fmt.Errorf("NewBlockChain failed, stop the program! err:%v", err)
	}

	n := &GlobalImpl{txDB: txDb, config: conf, stdPool: state.StdPool, blockChain: blockChain, mode: ModeNormal}

	return n, nil
}

func (g *GlobalImpl) TxDB() tx.TxPool {
	return g.txDB
}

func (g *GlobalImpl) StdPool() state.Pool {
	return g.stdPool
}

func (g *GlobalImpl) BlockChain() block.Chain {
	return g.blockChain
}

func (g *GlobalImpl) Config() *common.Config {
	return g.config
}

func (g *GlobalImpl) Mode() Mode {
	return g.mode
}

func (g *GlobalImpl) SetMode(mode Mode) bool {

	g.mode = mode

	return true
}
