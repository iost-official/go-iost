package protocol

import "IOS/src/iosbase"

type Replica struct {
	blockChain iosbase.BlockChain
	statePool iosbase.StatePool
}

func (c *Consensus) replicaInit (bc iosbase.BlockChain, sp iosbase.StatePool) error {
	c.blockChain = bc
	c.statePool = sp
	return nil
}

func (c *Consensus) verifyTx (tx iosbase.Tx) error {
	return nil
}