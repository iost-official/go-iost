package block

import (
	"errors"
	"fmt"

	"github.com/golang/protobuf/proto"
	"github.com/iost-official/go-iost/account"
	"github.com/iost-official/go-iost/common"
	blockpb "github.com/iost-official/go-iost/core/block/pb"
	"github.com/iost-official/go-iost/core/merkletree"
	"github.com/iost-official/go-iost/core/tx"
	"github.com/iost-official/go-iost/core/version"
	"github.com/iost-official/go-iost/crypto"
)

// BlockHead versions
const (
	V0 int64 = iota
	V1
)

// BlockHead is the struct of block head.
type BlockHead struct { // nolint
	Version             int64
	ParentHash          []byte
	TxMerkleHash        []byte
	TxReceiptMerkleHash []byte
	Info                []byte
	Number              int64
	Witness             string
	Time                int64
	GasUsage            int64
}

// ToPb convert BlockHead to proto buf data structure.
func (b *BlockHead) ToPb() *blockpb.BlockHead {
	return &blockpb.BlockHead{
		Version:             b.Version,
		ParentHash:          b.ParentHash,
		TxMerkleHash:        b.TxMerkleHash,
		TxReceiptMerkleHash: b.TxReceiptMerkleHash,
		Info:                b.Info,
		Number:              b.Number,
		Witness:             b.Witness,
		Time:                b.Time,
	}
}

// ToBytes converts BlockHead to a specific byte slice.
func (b *BlockHead) ToBytes() []byte {
	se := common.NewSimpleEncoder()
	se.WriteInt64(b.Version)
	se.WriteBytes(b.ParentHash)
	se.WriteBytes(b.TxMerkleHash)
	se.WriteBytes(b.TxReceiptMerkleHash)
	se.WriteBytes(b.Info)
	se.WriteInt64(b.Number)
	se.WriteString(b.Witness)
	se.WriteInt64(b.Time)
	return se.Bytes()
}

// FromPb convert BlockHead from proto buf data structure.
func (b *BlockHead) FromPb(bh *blockpb.BlockHead) *BlockHead {
	b.Version = bh.Version
	b.ParentHash = bh.ParentHash
	b.TxMerkleHash = bh.TxMerkleHash
	b.TxReceiptMerkleHash = bh.TxReceiptMerkleHash
	b.Info = bh.Info
	b.Number = bh.Number
	b.Witness = bh.Witness
	b.Time = bh.Time
	return b
}

// Encode is marshal
func (b *BlockHead) Encode() ([]byte, error) {
	bhByte, err := proto.Marshal(b.ToPb())
	if err != nil {
		return nil, errors.New("fail to encode blockhead")
	}
	return bhByte, nil
}

// Decode is unmarshal
func (b *BlockHead) Decode(bhByte []byte) error {
	bh := &blockpb.BlockHead{}
	err := proto.Unmarshal(bhByte, bh)
	if err != nil {
		return errors.New("fail to decode blockhead")
	}
	b.FromPb(bh)
	return nil
}

func (b *BlockHead) hash() []byte {
	return common.Sha3(b.ToBytes())
}

// Rules create new rules for this block
func (b *BlockHead) Rules() *version.Rules {
	return version.NewRules(b.Number)
}

// Block is the implementation of block
type Block struct {
	hash          []byte
	Head          *BlockHead
	Sign          *crypto.Signature
	Txs           []*tx.Tx
	Receipts      []*tx.TxReceipt
	TxHashes      [][]byte
	ReceiptHashes [][]byte
}

// CalculateGasUsage calculates the block's gas usage.
func (b *Block) CalculateGasUsage() int64 {
	if b.Head.GasUsage == 0 {
		for _, txr := range b.Receipts {
			b.Head.GasUsage += txr.GasUsage
		}
	}
	return b.Head.GasUsage
}

// CalculateTxMerkleHash calculate the merkle hash of the transaction.
func (b *Block) CalculateTxMerkleHash() []byte {
	m := merkletree.MerkleTree{}
	hashes := make([][]byte, 0, len(b.Txs))
	for _, tx := range b.Txs {
		hashes = append(hashes, tx.Hash())
	}
	m.Build(hashes)
	return m.RootHash()
}

// CalculateTxReceiptMerkleHash calculate the merkle hash of the transaction receipt.
func (b *Block) CalculateTxReceiptMerkleHash() []byte {
	m := merkletree.TXRMerkleTree{}
	m.Build(b.Receipts)
	return m.RootHash()
}

// Encode is marshal
func (b *Block) Encode() ([]byte, error) {
	br := &blockpb.Block{
		Head:      b.Head.ToPb(),
		BlockType: blockpb.BlockType_NORMAL,
	}
	for _, t := range b.Txs {
		br.Txs = append(br.Txs, t.ToPb())
	}
	for _, r := range b.Receipts {
		br.Receipts = append(br.Receipts, r.ToPb())
	}

	if b.Sign != nil {
		br.Sign = b.Sign.ToPb()
	}
	brByte, err := proto.Marshal(br)
	if err != nil {
		return nil, errors.New("fail to encode blockraw")
	}
	return brByte, nil
}

// Decode is unmarshal
func (b *Block) Decode(blockByte []byte) error {
	br := &blockpb.Block{}
	err := proto.Unmarshal(blockByte, br)
	if err != nil {
		return errors.New("fail to decode blockraw")
	}
	h := &BlockHead{}
	h.FromPb(br.Head)
	b.Head = h

	b.TxHashes = nil
	sig := &crypto.Signature{}
	b.Sign = sig.FromPb(br.Sign)

	switch br.BlockType {
	case blockpb.BlockType_NORMAL:
		for _, t := range br.Txs {
			tt := &tx.Tx{}
			b.Txs = append(b.Txs, tt.FromPb(t))
		}
		for _, r := range br.Receipts {
			rcpt := &tx.TxReceipt{}
			b.Receipts = append(b.Receipts, rcpt.FromPb(r))
		}
	case blockpb.BlockType_ONLYHASH:
		b.TxHashes = br.TxHashes
		b.ReceiptHashes = br.ReceiptHashes
	}
	b.CalculateHeadHash()
	return nil
}

// CalculateHeadHash calculate the hash of the head
func (b *Block) CalculateHeadHash() {
	b.hash = b.Head.hash()
}

// HeadHash return block hash
func (b *Block) HeadHash() []byte {
	return b.hash
}

// LenTx return len of transaction
func (b *Block) LenTx() int {
	return len(b.Txs)
}

// EncodeM is marshal
func (b *Block) EncodeM() ([]byte, error) {
	br := &blockpb.Block{
		Head:      b.Head.ToPb(),
		BlockType: blockpb.BlockType_ONLYHASH,
	}
	br.Sign = b.Sign.ToPb()
	for _, t := range b.Txs {
		br.TxHashes = append(br.TxHashes, t.Hash())
	}
	for _, r := range b.Receipts {
		br.ReceiptHashes = append(br.ReceiptHashes, r.Hash())
	}
	brByte, err := proto.Marshal(br)
	if err != nil {
		return nil, errors.New("fail to encode blockraw")
	}
	return brByte, nil
}

// VerifySelf verify block's signature and some base fields.
func (b *Block) VerifySelf() error {
	signature := b.Sign
	signature.SetPubkey(account.DecodePubkey(b.Head.Witness))
	hash := b.HeadHash()
	if !signature.Verify(hash) {
		return fmt.Errorf("The signature of block %v is wrong", common.Base58Encode(hash))
	}
	if len(b.Txs) != len(b.Receipts) {
		return fmt.Errorf("Tx len %v unmatch receipt len %v", len(b.Txs), len(b.Receipts))
	}
	return nil

}
