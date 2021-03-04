package main

import (
	"bytes"
	"fmt"

	"github.com/iost-official/go-iost/v3/common"
	"github.com/iost-official/go-iost/v3/core/block"
	"github.com/iost-official/go-iost/v3/core/blockcache"
)

const VerifierNum = 17

type Verifier struct {
	CurrentProducer []string
	EpochProducer   map[int64][]string
}

func (v *Verifier) init(blk *block.Block) error {
	// Here we believe this block as truth
	if blk.Head.Number%common.VoteInterval != 0 {
		return fmt.Errorf("invalid spv start block %v", blk.Head.Number)
	}
	w, err := blockcache.GetWitnessStatusFromBlock(blk)
	if err != nil {
		return err
	}
	if len(w.PendingList) != VerifierNum {
		return fmt.Errorf("invalid pending list %v at block %v", w.PendingList, blk.Head.Number)
	}
	v.CurrentProducer = w.PendingList
	v.EpochProducer = make(map[int64][]string)
	v.EpochProducer[blk.Head.Number] = w.PendingList
	return nil
}

func (v *Verifier) checkWitness(blk *block.Block, witnessBlocks []*block.Block) error {
	if blk.VerifySelf() != nil {
		return fmt.Errorf("invalid block signature")
	}
	for _, b := range witnessBlocks {
		if b.VerifySelf() != nil {
			return fmt.Errorf("invalid block signature")
		}
	}
	blockNumber := blk.Head.Number
	// we should check this blk is verified by more than 2/3 of current validators
	var currentEpochStartBlock int64
	if blockNumber%common.VoteInterval == 0 {
		currentEpochStartBlock = blockNumber - common.VoteInterval
	} else {
		currentEpochStartBlock = blockNumber / common.VoteInterval * common.VoteInterval
	}
	currentProducer, succ := v.EpochProducer[currentEpochStartBlock]
	if !succ {
		return fmt.Errorf("cannot update producer list at block %v: cannot find producer info of previous epoch", blockNumber)
	}
	var validWitness = make(map[string]bool)
	var validWitnessCount = 0
	parentHash := blk.HeadHash()
	parentBlockNumber := blockNumber
	for _, b := range witnessBlocks {
		if !bytes.Equal(b.Head.ParentHash, parentHash) {
			return fmt.Errorf("invalid block hash at block %v", b.Head.Number)
		}
		if b.Head.Number != parentBlockNumber+1 {
			return fmt.Errorf("invalid block number at block %v", b.Head.Number)
		}
		// Now we checked this `b` is a child of previous block
		_, succ := validWitness[b.Head.Witness]
		if !succ {
			for _, elem := range currentProducer {
				if b.Head.Witness == elem {
					validWitness[elem] = true
					validWitnessCount++
					break
				}
			}
		}
		parentBlockNumber = b.Head.Number
		parentHash = b.HeadHash()
	}
	// we need 12 (2/3 * 17) witness to confirm a block
	if validWitnessCount < 12 {
		return fmt.Errorf("valid witness not enough %v", validWitness)
	}
	return nil
}

func (v *Verifier) updateEpoch(blk *block.Block, witnessBlocks []*block.Block) error {
	voteBlockNumber := blk.Head.Number
	if voteBlockNumber%common.VoteInterval != 0 {
		return fmt.Errorf("invalid spv start block %v", voteBlockNumber)
	}
	w, err := blockcache.GetWitnessStatusFromBlock(blk)
	if err != nil {
		return err
	}
	if len(w.PendingList) != VerifierNum {
		return fmt.Errorf("invalid pending list %v at block %v", w.PendingList, voteBlockNumber)
	}
	err = v.checkWitness(blk, witnessBlocks)
	if err != nil {
		return err
	}
	v.EpochProducer[voteBlockNumber] = w.PendingList
	return nil
}

func (v *Verifier) checkBlock(blk *block.Block, witnessBlocks []*block.Block) error {
	err := v.checkWitness(blk, witnessBlocks)
	if err != nil {
		return err
	}
	fmt.Printf("check block %v done\n", blk.Head.Number)
	return nil
}
