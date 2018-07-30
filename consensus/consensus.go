package consensus

import (
	"sync"

	"github.com/iost-official/Go-IOS-Protocol/account"
	"github.com/iost-official/Go-IOS-Protocol/consensus/pob"
	"github.com/iost-official/Go-IOS-Protocol/core/block"
	"github.com/iost-official/Go-IOS-Protocol/core/blockcache"
	"github.com/iost-official/Go-IOS-Protocol/core/state"
)

type Consensus interface {
	Run()
	Stop()

	BlockChain() block.Chain
	CachedBlockChain() block.Chain
	BlockCache() blockcache.BlockCache
	StatePool() state.Pool
	CachedStatePool() state.Pool
}

const (
	CONSENSUS_POB = "pob"
)

var Cons Consensus

var once sync.Once

func ConsensusFactory(consensusType string, acc account.Account, bc block.Chain, pool state.Pool, witnessList []string) (Consensus, error) {

	if consensusType == "" {
		consensusType = CONSENSUS_POB
	}

	var err error

	switch consensusType {
	case CONSENSUS_POB:
		if Cons == nil {
			once.Do(func() {
				Cons, err = pob.NewPoB(acc, bc, pool, witnessList)
			})
		}
	}
	return Cons, err
}
