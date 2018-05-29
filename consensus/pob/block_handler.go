package pob

import (
	"time"

	"errors"

	"github.com/iost-official/prototype/common"
	"github.com/iost-official/prototype/consensus/common"
	"github.com/iost-official/prototype/core/block"
	"github.com/iost-official/prototype/core/tx"
)

const MaxTxInBlock = 1000

type Generator interface {
	NewBlock() block.Block
}

type GeneratorImpl struct {
	holder *Holder
	bc     consensus_common.BlockCache
	rec    *Recorder
}

func (g *GeneratorImpl) NewBlock() block.Block {
	base := g.bc.LongestChain().Top()
	content := make([]tx.Tx, 1000)
	(*g.rec).Close()
	for i := 0; i < MaxTxInBlock; i++ {
		txx := (*g.rec).Pop()
		content = append(content, txx)
	}
	(*g.rec).Listen()
	bh := block.BlockHead{
		Version:    1,
		ParentHash: base.HeadHash(),
		TreeHash:   nil,
		BlockHash:  nil,
		Time:       time.Now().Unix(),
		Witness:    g.holder.self.ID,
	}
	sig, err := common.Sign(common.Secp256k1, bh.Hash(), g.holder.self.Seckey)
	if err != nil {
		panic(err)
	}
	bh.Signature = sig.Encode()

	blk := block.Block{Head: bh, Content: content}
	return blk
}

type Checker interface {
	Check(block2 *block.Block) error
}

type CheckerImpl struct {
}

func (c CheckerImpl) Check(block2 *block.Block) error {
	bh := block2.Head
	var sig common.Signature
	sig.Decode(bh.Signature)
	bh.Signature = nil
	if !common.VerifySignature(bh.Hash(), sig) {
		return errors.New("sign error")
	}
	return nil
}
