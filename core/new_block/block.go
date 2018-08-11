package block

import (
	"errors"
	"fmt"

	"github.com/gogo/protobuf/proto"
	"github.com/iost-official/Go-IOS-Protocol/account"
	"github.com/iost-official/Go-IOS-Protocol/common"
	"github.com/iost-official/Go-IOS-Protocol/core/merkletree"
	"github.com/iost-official/Go-IOS-Protocol/core/new_tx"
	)

type Block struct {
	hash     []byte
	Head     BlockHead
	Txs      []*tx.Tx
	Receipts []*tx.TxReceipt
}

func GenGenesis(initTime int64) *Block {
	var code string
	for k, v := range account.GenesisAccount {
		code += fmt.Sprintf("@PutHM iost %v f%v\n", k, v)
	}

	txn := tx.Tx{
		Time: 0,
		// TODO what is the genesis tx?
	}

	genesis := &Block{
		Head: BlockHead{
			Version: 0,
			Number:  0,
			Time:    initTime,
		},
		Txs:      make([]*tx.Tx, 0),
		Receipts: make([]*tx.TxReceipt, 0),
	}
	genesis.Txs = append(genesis.Txs, &txn)
	return genesis
}

func (b *Block) CalculateTxsHash() ([]byte, error) {
	if len(b.Txs) == 0 {
		return nil, errors.New("txs length equals 0")
	}
	hash := make([]byte, 0)
	for _, tx := range b.Txs {
		hash = append(hash, tx.Publisher.Sig...)
	}
	return common.Sha3(hash), nil
}

func (b *Block) CalculateMerkleHash() ([]byte, error) {
	m := merkletree.TXRMerkleTree{}
	err := m.Build(b.Receipts)
	if err != nil {
		return nil, err
	}
	return m.RootHash()
}

func (b *Block) Encode() ([]byte, error) {
	txs := make([][]byte, 0)
	for _, t := range b.Txs {
		txs = append(txs, t.Encode())
	}
	rpts := make([][]byte, 0)
	for _, r := range b.Receipts {
		rpts = append(rpts, r.Encode())
	}
	br := &BlockRaw{
		Head:     &b.Head,
		Txs:      txs,
		Receipts: rpts,
	}
	brByte, err := proto.Marshal(br)
	if err != nil {
		return nil, errors.New("fail to encode blockraw")
	}
	return brByte, nil
}

func (b *Block) Decode(blockByte []byte) error {
	br := &BlockRaw{}
	err := proto.Unmarshal(blockByte, br)
	if err != nil {
		return errors.New("fail to decode blockraw")
	}
	b.Head = *br.Head

	for _, t := range br.Txs {
		var tt tx.Tx
		err = tt.Decode(t)
		if err != nil {
			return errors.New("fail to decode tx")
		}
		b.Txs = append(b.Txs, &tt)
	}
	for _, r := range br.Receipts {
		var rcpt tx.TxReceipt
		err = rcpt.Decode(r)
		if err != nil {
			return errors.New("fail to decode txr")
		}
		b.Receipts = append(b.Receipts, &rcpt)
	}
	return nil
}

func (b *Block) HeadHash() ([]byte, error) {
	var err error
	if b.hash == nil {
		b.hash, err = b.Head.Hash()
		if err != nil {
			return nil, err
		}
	}
	return b.hash, nil
}

func (b *Block) LenTx() int {
	return len(b.Txs)
}

func (b *BlockHead) Encode() ([]byte, error) {
	bhByte, err := proto.Marshal(b)
	if err != nil {
		return nil, errors.New("fail to encode blockhead")
	}
	return bhByte, nil
}

func (b *BlockHead) Decode(bhByte []byte) error {
	err := proto.Unmarshal(bhByte, b)
	if err != nil {
		return errors.New("fail to decode blockhead")
	}
	return nil
}

func (b *BlockHead) Hash() ([]byte, error) {
	bhByte, err := b.Encode()
	if err != nil {
		return nil, err
	}
	return common.Sha3(bhByte), nil
}
