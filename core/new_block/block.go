package block

import (
	"fmt"
	"strconv"

	"github.com/iost-official/Go-IOS-Protocol/common"
	"github.com/iost-official/Go-IOS-Protocol/core/new_tx"
	"github.com/gogo/protobuf/proto"
	"github.com/iost-official/Go-IOS-Protocol/core/merkletree"
)

type Block struct {
	hash     []byte
	Head     BlockHead
	Txs      []tx.Tx
	Receipts []tx.TxReceipt
}

func (b *Block) CalculateTxsHash() []byte {
	treeHash := make([]byte, 0)
	for _, tx := range b.Txs {
		treeHash = append(treeHash, tx.Publisher.Sig...)
	}
	return common.Sha256(treeHash)
}

func (b *Block) CalculateMerkleHash() ([]byte, error) {
	m := merkletree.TXRMerkleTree{}
	err := m.Build(b.Receipts)
	if err != nil {
		return nil, err
	}
	return m.RootHash()
}

func (b *Block) Encode() []byte {
	txs := make([][]byte, 0)
	for _, t := range b.Txs {
		txs = append(txs, t.Encode())
	}
	rpts := make([][]byte, 0)
	for _, r := range b.Receipts {
		rpts = append(rpts, r.Encode())
	}
	br := &BlockRaw{
		Head: &b.Head,
		Txs: txs,
		Receipts: rpts,
	}

	brByte, err := proto.Marshal(br)
	if err != nil {
		panic(err)
	}
	b.hash = brByte
	return brByte
}

func (b *Block) Decode(bin []byte) (err error) {
	br := &BlockRaw{}
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("%v", r)
		}
	}()

	err = proto.Unmarshal(bin, br)
	b.Head = *br.Head
	for _, t := range br.Txs {
		var tt tx.Tx
		err = tt.Decode(t)
		if err != nil {
			return err
		}
		b.Txs = append(b.Txs, tt)
	}
	for _, r := range br.Receipts {
		var rcpt tx.TxReceipt
		err = rcpt.Decode(r)
		if err != nil {
			return err
		}
		b.Receipts = append(b.Receipts, rcpt)
	}
	return nil
}

func (b *Block) HashID() string {
	id := b.Head.Witness +
		strconv.FormatInt(b.Head.Time, 10) +
		strconv.FormatInt(b.Head.Number, 10) +
		strconv.FormatInt(b.Head.Version, 10)
	return id
}

func (b *Block) HeadHash() []byte {
	if b.hash == nil {
		b.hash = b.Head.Hash()
	}
	return b.hash
}

func (b *Block) GetTx(x int) tx.Tx {
	if x < len(b.Txs) {
		return b.Txs[x]
	} else {
		return tx.Tx{}
	}
}

func (b *Block) LenTx() int {
	return len(b.Txs)
}

func (b *BlockHead) Encode() []byte {
	bin, err := proto.Marshal(b)
	if err != nil {
		panic(err)
	}
	return bin
}

func (b *BlockHead) Decode(bin []byte) error {
	err := proto.Unmarshal(bin, b)
	return err
}

func (b *BlockHead) Hash() []byte {
	return common.Sha256(b.Encode())
}
