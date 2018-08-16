package new_consensus

import (
	"sync"

	"github.com/iost-official/Go-IOS-Protocol/consensus/pob"
	"github.com/iost-official/Go-IOS-Protocol/core/global"
	"github.com/iost-official/Go-IOS-Protocol/core/new_blockcache"
	"github.com/iost-official/Go-IOS-Protocol/p2p"
	"github.com/iost-official/Go-IOS-Protocol/account"
	"github.com/iost-official/Go-IOS-Protocol/consensus/common"
	"github.com/iost-official/Go-IOS-Protocol/core/new_txpool"
)

type Consensus interface {
	Run()
	Stop()
}

const (
	CONSENSUS_POB = "pob"
)

var Cons Consensus

var once sync.Once

func ConsensusFactory(consensusType string, acc account.Account, txPool txpool.TxPool, baseVariable global.BaseVariable, blkcache blockcache.BlockCache, p2pserv p2p.Service, sy consensus_common.Synchronizer, witnessList []string) (Consensus, error) {

	if consensusType == "" {
		consensusType = CONSENSUS_POB
	}

	var err error

	switch consensusType {
	case CONSENSUS_POB:
		if Cons == nil {
			once.Do(func() {
				Cons, err = pob.NewPoB(acc, baseVariable, blkcache, txPool, p2pserv, sy, witnessList)
			})
		}
	}
	return Cons, err
}
