package transaction

import (
	"time"
	"bytes"
	"encoding/gob"
	"log"
	"crypto/sha256"
	"github.com/iost-official/prototype/tx/min_framework"
)


// datastructure of a block in blockchain
// 时间戳
// 交易
// 前一个块哈希
// 当前块哈希
// 难度值
type Block struct {
	Timestamp int64
	Transactions []*Transaction
	PrevBlockHash []byte
	Hash []byte
	Nonce int
}


func (b *Block) Serialize() []byte {
	var result bytes.Buffer
	encoder := gob.NewEncoder(&result)

	err := encoder.Encode(b)
	if err != nil {
		log.Panic(err)
	}

	return result.Bytes()
}

// Deserialize a block
func DeserializeBlock(d []byte) *Block {
	var block Block

	decoder := gob.NewDecoder(bytes.NewReader(d))
	err := decoder.Decode(&block)
	if err != nil {
		log.Panic(err)
	}

	return &block
}


func (block *Block) HashBlock(nonce int) []byte{
	data := bytes.Join(
		[][]byte{
			block.PrevBlockHash,
			block.HashTransactions(),
			min_framework.IntToHex(block.Timestamp),
			min_framework.IntToHex(int64(min_framework.TargetBits)),
			min_framework.IntToHex(int64(nonce)),
		},
		[]byte{},
	)
	hash := sha256.Sum256(data)

	return hash[:]
}


// NewBlock creates and returns Block
func NewBlock(transactions []*Transaction, prevBlockHash []byte) *Block {
	block := &Block{
		Timestamp:     time.Now().Unix(),
		Transactions:  transactions,
		PrevBlockHash: prevBlockHash,
		Hash:          []byte{},
		Nonce:         0}
	pow := NewProofOfWork(block)
	nonce, hash := pow.Mine()

	block.Hash = hash[:]
	block.Nonce = nonce

	return block
}


// NewGenesisBlock creates and returns Genesis Block
func NewGenesisBlock(coinbase *Transaction) *Block {
	return NewBlock([]*Transaction{coinbase}, []byte{})
}

// 计算区块里所有交易的哈希
func (b *Block) HashTransactions() []byte {
	var txHashes [][]byte
	var txHash [32]byte

	for _, tx := range b.Transactions {
		txHashes = append(txHashes, tx.ID)
	}
	txHash = sha256.Sum256(bytes.Join(txHashes, []byte{}))

	return txHash[:]
}