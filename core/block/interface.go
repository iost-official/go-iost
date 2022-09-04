package block

import "github.com/iost-official/go-iost/v3/core/tx"

//go:generate mockgen -destination ../mocks/mock_blockchain.go -package core_mock -source interface.go

// Chain defines Chain's API.
type Chain interface {
	Push(block *Block) error
	Length() int64
	TxTotal() int64
	CheckLength()
	SetLength(i int64)
	Top() (*Block, error)
	GetHashByNumber(number int64) ([]byte, error)
	GetBlockByNumber(number int64) (*Block, error)
	GetBlockByHash(blockHash []byte) (*Block, error)
	GetTx(hash []byte) (*tx.Tx, error)
	HasTx(hash []byte) (bool, error)
	GetReceipt(Hash []byte) (*tx.TxReceipt, error)
	GetReceiptByTxHash(Hash []byte) (*tx.TxReceipt, error)
	HasReceipt(hash []byte) (bool, error)
	Size() (int64, error)
	Close()
	Draw(int64, int64) string
	GetBlockNumberByTxHash(hash []byte) (int64, error)
}
