package consensus

import (
	"sync"

	"github.com/iost-official/Go-IOS-Protocol/account"
	"github.com/iost-official/Go-IOS-Protocol/consensus/pob"
	"github.com/iost-official/Go-IOS-Protocol/consensus/synchronizer"
	"github.com/iost-official/Go-IOS-Protocol/core/blockcache"
	"github.com/iost-official/Go-IOS-Protocol/core/global"
	"github.com/iost-official/Go-IOS-Protocol/core/new_txpool"
	"github.com/iost-official/Go-IOS-Protocol/p2p"
)

// Consensus handles the different consensus strategy.
type Consensus interface {
	Start() error
	Stop()
}

var cons Consensus

var once sync.Once

// Factory handles the different consensus strategy.
func Factory(consensusType string, account account.Account, baseVariable global.BaseVariable, blkcache blockcache.BlockCache, txPool txpool.TxPool, service p2p.Service, synchronizer synchronizer.Synchronizer, witnessList []string) (Consensus, error) {
	if consensusType == "" {
		consensusType = "pob"
	}

	var err error

	switch consensusType {
	case "pob":
		if cons == nil {
			once.Do(func() {
				cons = pob.NewPoB(account, baseVariable, blkcache, txPool, service, synchronizer, witnessList)
			})
		}
	}
	return cons, err
}
