package iost

type IConsensus interface {
	Verify (blk IBlock, statePool []State) bool
}

