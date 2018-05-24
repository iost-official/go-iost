package pob

import (
	"github.com/iost-official/prototype/consensus/common"
	"github.com/iost-official/prototype/core/block"
)

type Generator interface {
	NewBlock() block.Block
}

type GeneratorImpl struct {
	bc  consensus_common.BlockCache
	rec Recorder
}

func (g *GeneratorImpl) NewBlock() block.Block {
	base := g.bc.LongestChain().Top()
	bh := block.BlockHead{
		ParentHash: base.HeadHash(),
	}
	return block.Block{Head: bh}
}

type Checker interface {
}

type pobCheckerImpl struct {
}
