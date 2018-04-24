package state

import "github.com/iost-official/prototype/core"

type Patch struct {

}

type Pool interface {
	Copy() Pool
	GetPatch() Patch
}

func NewPool(bc core.BlockChain) Pool {
	return nil
}

type PoolImpl struct {

}