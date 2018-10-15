package block

import (
	"errors"

	"github.com/gogo/protobuf/proto"
	"github.com/iost-official/go-iost/common"
	"github.com/iost-official/go-iost/core/merkletree"
	"github.com/iost-official/go-iost/core/tx"
	"github.com/iost-official/go-iost/crypto"
)

// Block is the implementation of block
type Block struct {
	hash     []byte
	Head     *BlockHead
	Sign     *crypto.Signature
	Txs      []*tx.Tx
	Receipts []*tx.TxReceipt
}

// CalculateTxsHash calculate the hash of the transaction
func (b *Block) CalculateTxsHash() []byte {
	hash := make([]byte, 0)
	for _, tx := range b.Txs {
		hash = append(hash, tx.Publisher.Sig...)
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
	br := &BlockRaw{
		Head: b.Head,
	}
	for _, t := range b.Txs {
		br.Txs = append(br.Txs, t.ToTxRaw())
	}
	for _, r := range b.Receipts {
		br.Receipts = append(br.Receipts, r.ToTxReceiptRaw())
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
	br := &BlockRaw{}
	err := proto.Unmarshal(blockByte, br)
	if err != nil {
		return errors.New("fail to decode blockraw")
	}
	b.Head = br.Head
	b.Sign = &crypto.Signature{
		Algorithm: crypto.Algorithm(br.Sign.Algorithm),
		Sig:       br.Sign.Sig,
		Pubkey:    br.Sign.PubKey,
	}
	if err != nil {
		return errors.New("fail to decode signature")
	}
	for _, t := range br.Txs {
		var tt tx.Tx
		tt.FromTxRaw(t)
		b.Txs = append(b.Txs, &tt)
	}
	for _, r := range br.Receipts {
		var rcpt tx.TxReceipt
		rcpt.FromTxReceiptRaw(r)
		b.Receipts = append(b.Receipts, &rcpt)
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

// Encode is marshal
func (b *BlockHead) Encode() ([]byte, error) {
	bhByte, err := proto.Marshal(b)
	if err != nil {
		return nil, errors.New("fail to encode blockhead")
	}
	return bhByte, nil
}

// Decode is unmarshal
func (b *BlockHead) Decode(bhByte []byte) error {
	err := proto.Unmarshal(bhByte, b)
	if err != nil {
		return errors.New("fail to decode blockhead")
	}
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

// EncodeM is marshal
func (b *Block) EncodeM() ([]byte, error) {
	br := &BlockRaw{
		Head: b.Head,
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
	brByte, err := proto.Marshal(br)
	if err != nil {
		return nil, errors.New("fail to encode blockraw")
	}
	return brByte, nil
}

// DecodeM is unmarshal
func (b *Block) DecodeM(blockByte []byte) error {
	br := &BlockRaw{}
	err := proto.Unmarshal(blockByte, br)
	if err != nil {
		return errors.New("fail to decode blockraw")
	}
	b.Head = br.Head
	b.Sign = &crypto.Signature{
		Algorithm: crypto.Algorithm(br.Sign.Algorithm),
		Sig:       br.Sign.Sig,
		Pubkey:    br.Sign.PubKey,
	}
	return b.CalculateHeadHash()
}
