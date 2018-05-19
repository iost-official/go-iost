package consensus

import (
	"github.com/iost-official/prototype/core/block"
	"github.com/iost-official/prototype/core/state"
	"github.com/iost-official/prototype/consensus/dpos"
	"github.com/iost-official/prototype/account"
)

type TxStatus int

const (
	ACCEPT  TxStatus = iota
	CACHED
	POOL
	REJECT
	EXPIRED
)

type Consensus interface {
	Run()
	Stop()

	BlockChain() block.Chain
	CachedBlockChain() block.Chain
	StatePool() state.Pool
	CachedStatePool() state.Pool
}

const (
	CONSENSUS_DPOS = "dpos"
)

var Cons Consensus
func ConsensusFactory(consensusType string,acc account.Account, bc block.Chain, pool state.Pool, witnessList []string) (Consensus, error) {

	if consensusType == ""{
		consensusType = CONSENSUS_DPOS
	}

	var err error

	switch consensusType {
	case CONSENSUS_DPOS:
		Cons, err = dpos.NewDPoS(acc, bc, pool, witnessList)
	}
	return Cons,err
}
