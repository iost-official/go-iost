package itest

import (
	rpcpb "github.com/iost-official/go-iost/rpc/pb"
)

// Block is the block object
type Block struct {
	Hash                string
	Version             int64
	ParentHash          string
	TxMerkleHash        string
	TxReceiptMerkleHash string
	Number              int64
	Witness             string
	Time                int64
}

// NewBlockFromPb returns a new Block instance from protobuffer struct.
func NewBlockFromPb(b *rpcpb.Block) *Block {
	ret := &Block{
		Hash:                b.Hash,
		Version:             b.Version,
		ParentHash:          b.ParentHash,
		TxMerkleHash:        b.TxMerkleHash,
		TxReceiptMerkleHash: b.TxReceiptMerkleHash,
		Number:              b.Number,
		Witness:             b.Witness,
		Time:                b.Time,
	}
	return ret
}
