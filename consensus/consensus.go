package consensus

import (
	"sync"

	"github.com/iost-official/go-iost/account"
	"github.com/iost-official/go-iost/consensus/pob"
	"github.com/iost-official/go-iost/core/blockcache"
	"github.com/iost-official/go-iost/core/global"
	"github.com/iost-official/go-iost/core/txpool"
	"github.com/iost-official/go-iost/p2p"
)

// Consensus handles the different consensus strategy.
type Consensus interface {
	Start() error
	Stop()
}

var cons Consensus

var once sync.Once

// Factory handles the different consensus strategy.
func Factory(consensusType string, account *account.Account, baseVariable global.BaseVariable, blkcache blockcache.BlockCache, txPool txpool.TxPool, service p2p.Service) (Consensus, error) {
	if consensusType == "" {
		consensusType = "pob"
	}

	var err error

	switch consensusType {
	case "pob":
		if cons == nil {
			once.Do(func() {
				cons = pob.NewPoB(account, baseVariable, blkcache, txPool, service)
			})
		}
	}
	return cons, err
}
