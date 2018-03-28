package protocol

import "IOS/src/iosbase"

type RuntimeData struct {
	iosbase.Member

	character Character
	view      View
	phase     Phase
	isRunning bool

	blockChain iosbase.BlockChain
	statePool  iosbase.StatePool
}
