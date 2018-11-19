package block

import (
	"errors"

	"github.com/iost-official/go-iost/common"
	blockpb "github.com/iost-official/go-iost/core/block/pb"
	"github.com/iost-official/go-iost/core/merkletree"
	"github.com/iost-official/go-iost/core/tx"
	"github.com/iost-official/go-iost/crypto"
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
	sn := common.NewSimpleNotation()
	sn.WriteInt64(b.Version, true)
	sn.WriteBytes(b.ParentHash, false)
	sn.WriteBytes(b.TxMerkleHash, false)
	sn.WriteBytes(b.TxReceiptMerkleHash, false)
	sn.WriteBytes(b.Info, true)
	sn.WriteInt64(b.Number, true)
	sn.WriteString(b.Witness, true)
	sn.WriteInt64(b.Time, true)
	return sn.Bytes()
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
	return common.Sha3(b.ToBytes()), nil
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
func (b *Block) CalculateGasUsage() {
	for _, txr := range b.Receipts {
		b.Head.GasUsage += txr.GasUsage
	}
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

	br.Sign = b.Sign.ToPb()
	brByte, err := br.Marshal()
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
	sig := &crypto.Signature{}
	b.Sign = sig.FromPb(br.Sign)
	if err != nil {
		return errors.New("fail to decode signature")
	}
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
	b.CalculateGasUsage()
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
	br.Sign = b.Sign.ToPb()
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
