package block

import (
	"errors"

	"github.com/gogo/protobuf/proto"
	"github.com/iost-official/go-iost/common"
	blockpb "github.com/iost-official/go-iost/core/block/pb"
	"github.com/iost-official/go-iost/core/merkletree"
	"github.com/iost-official/go-iost/core/tx"
	"github.com/iost-official/go-iost/crypto"
)

// BlockHead is the struct of block head.
type BlockHead struct { // nolint
	Version    int64
	ParentHash []byte
	TxsHash    []byte
	MerkleHash []byte
	Info       []byte
	Number     int64
	Witness    string
	Time       int64
}

// ToPb convert BlockHead to proto buf data structure.
func (b *BlockHead) ToPb() *blockpb.BlockHead {
	return &blockpb.BlockHead{
		Version:    b.Version,
		ParentHash: b.ParentHash,
		TxsHash:    b.TxsHash,
		MerkleHash: b.MerkleHash,
		Info:       b.Info,
		Number:     b.Number,
		Witness:    b.Witness,
		Time:       b.Time,
	}
}

// FromPb convert BlockHead from proto buf data structure.
func (b *BlockHead) FromPb(bh *blockpb.BlockHead) {
	b.Version = bh.Version
	b.ParentHash = bh.ParentHash
	b.TxsHash = bh.TxsHash
	b.MerkleHash = bh.MerkleHash
	b.Info = bh.Info
	b.Number = bh.Number
	b.Witness = bh.Witness
	b.Time = bh.Time
}

// Encode is marshal
func (b *BlockHead) Encode() ([]byte, error) {
	bhByte, err := b.ToPb().Marshal()
	if err != nil {
		return nil, errors.New("fail to encode blockhead")
	}
	return bhByte, nil
}

// Decode is unmarshal
func (b *BlockHead) Decode(bhByte []byte) error {
	bh := &blockpb.BlockHead{}
	err := bh.Unmarshal(bhByte)
	if err != nil {
		return errors.New("fail to decode blockhead")
	}
	b.FromPb(bh)
	return nil
}

// Hash return hash
func (b *BlockHead) Hash() ([]byte, error) {
	bhByte, err := b.Encode()
	if err != nil {
		return nil, err
	}
	return common.Sha3(bhByte), nil
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

// CalculateTxsHash calculate the hash of the transaction
func (b *Block) CalculateTxsHash() []byte {
	hash := make([]byte, 0)
	for _, tx := range b.Txs {
		for _, sig := range tx.PublishSigns {
			hash = append(hash, sig.Sig...)
		}
	}
	return common.Sha3(hash)
}

// CalculateMerkleHash calculate the hash of the MerkleTree
func (b *Block) CalculateMerkleHash() []byte {
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
	br.Sign = &crypto.SignatureRaw{
		Algorithm: int32(b.Sign.Algorithm),
		Sig:       b.Sign.Sig,
		PubKey:    b.Sign.Pubkey,
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
	err := br.Unmarshal(blockByte)
	if err != nil {
		return errors.New("fail to decode blockraw")
	}
	h := &BlockHead{}
	h.FromPb(br.Head)
	b.Head = h

	b.TxHashes = nil
	b.Sign = &crypto.Signature{
		Algorithm: crypto.Algorithm(br.Sign.Algorithm),
		Sig:       br.Sign.Sig,
		Pubkey:    br.Sign.PubKey,
	}
	if err != nil {
		return errors.New("fail to decode signature")
	}
	switch br.BlockType {
	case blockpb.BlockType_NORMAL:
		for _, t := range br.Txs {
			var tt tx.Tx
			tt.FromPb(t)
			b.Txs = append(b.Txs, &tt)
		}
		for _, r := range br.Receipts {
			var rcpt tx.TxReceipt
			rcpt.FromPb(r)
			b.Receipts = append(b.Receipts, &rcpt)
		}
	case blockpb.BlockType_ONLYHASH:
		b.TxHashes = br.TxHashes
		b.ReceiptHashes = br.ReceiptHashes
	}
	return b.CalculateHeadHash()
}

// CalculateHeadHash calculate the hash of the head
func (b *Block) CalculateHeadHash() error {
	var err error
	b.hash, err = b.Head.Hash()
	return err
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
	br.Sign = &crypto.SignatureRaw{
		Algorithm: int32(b.Sign.Algorithm),
		Sig:       b.Sign.Sig,
		PubKey:    b.Sign.Pubkey,
	}
	for _, t := range b.Txs {
		br.TxHashes = append(br.TxHashes, t.Hash())
	}
	for _, r := range b.Receipts {
		br.ReceiptHashes = append(br.ReceiptHashes, r.Hash())
	}
	brByte, err := br.Marshal()
	if err != nil {
		return nil, errors.New("fail to encode blockraw")
	}
	return brByte, nil
}
