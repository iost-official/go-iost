package global

import (
	"errors"
	"github.com/iost-official/Go-IOS-Protocol/common"
	"github.com/iost-official/Go-IOS-Protocol/core/state"
	"github.com/iost-official/Go-IOS-Protocol/core/tx"
	"github.com/iost-official/prototype/core/block"
	"github.com/spf13/viper"
)

type GlobalImpl struct {
	txDB *tx.TxPoolDb

	stdPool    *state.Pool
	blockChain *block.Chain

	config *common.Config
}

func New(viper *viper.Viper) (Global, error) {

	conf, err := common.NewConfig()
	if err != nil {
		return nil, errors.New("NewConfig error")
	}

	if err := conf.LocalConfig(viper); err != nil {
		return nil, errors.New("config LocalConfig error")
	}

	g := &GlobalImpl{config: conf}

	return g, nil
}

func (g *GlobalImpl) TxDB() *tx.TxPoolDb {
	return g.txDB
}

func (g *GlobalImpl) StdPool() *state.Pool {
	return g.stdPool
}

func (g *GlobalImpl) BlockChain() *block.Chain {
	return g.blockChain
}

func (g *GlobalImpl) Config() *common.Config {
	return g.config
}
