package global

import (
	"errors"
	"fmt"

	"github.com/iost-official/go-iost/core/tx"
	"github.com/iost-official/go-iost/db/kv"
)

//go:generate mockgen -destination ../mocks/mock_txdb.go -package core_mock github.com/iost-official/go-iost/core/global TxDB

// TxDB defines the functions of tx database.
type TxDB interface {
	Push(hash []byte, txs []*tx.Tx, receipts []*tx.TxReceipt) error
	GetTx(hash []byte) (*tx.Tx, error)
	HasTx(hash []byte) (bool, error)
	GetReceipt(Hash []byte) (*tx.TxReceipt, error)
	GetReceiptByTxHash(Hash []byte) (*tx.TxReceipt, error)
	HasReceipt(hash []byte) (bool, error)
	Close()
}

// TxDBImpl is the implementation of TxDB.
type TxDBImpl struct {
	txDB *kv.Storage
}

var (
	txPrefix        = []byte("t") // txPrefix+tx hash -> tx data
	bTxPrefix       = []byte("B") // txPrefix+tx hash -> tx data
	txReceiptPrefix = []byte("h") // receiptHashPrefix + tx hash -> receipt hash
	receiptPrefix   = []byte("r") // receiptPrefix + receipt hash -> receipt data
	bReceiptPrefix  = []byte("b") // txPrefix+tx hash -> tx data
)

// NewTxDB returns a TxDB instance.
func NewTxDB(path string) (TxDB, error) {
	ldb, err := kv.NewStorage(path, kv.LevelDBStorage)
	if err != nil {
		return nil, err
	}
	return &TxDBImpl{txDB: ldb}, nil
}

// Push save the tx to database
func (tdb *TxDBImpl) Push(hash []byte, txs []*tx.Tx, receipts []*tx.TxReceipt) error {
	err := tdb.txDB.BeginBatch()
	if err != nil {
		return errors.New("fail to begin batch")
	}

	for i, tx := range txs {
		tHash := tx.Hash()
		tdb.txDB.Put(append(txPrefix, tHash...), append(hash, tHash...))
		tdb.txDB.Put(append(bTxPrefix, append(hash, tHash...)...), tx.Encode())

		// save receipt
		rHash := receipts[i].Hash()
		tdb.txDB.Put(append(txReceiptPrefix, tHash...), append(hash, rHash...))
		tdb.txDB.Put(append(receiptPrefix, rHash...), append(hash, rHash...))
		tdb.txDB.Put(append(bReceiptPrefix, append(hash, rHash...)...), receipts[i].Encode())
	}

	err = tdb.txDB.CommitBatch()
	if err != nil {
		return fmt.Errorf("fail to put block, err:%s", err)
	}
	return nil
}

// GetTx gets tx with tx's hash.
func (tdb *TxDBImpl) GetTx(hash []byte) (*tx.Tx, error) {
	tx := tx.Tx{}
	bTx, err := tdb.txDB.Get(append(txPrefix, hash...))
	if err != nil {
		return nil, fmt.Errorf("failed to Get the tx: %v", err)
	}
	txData, err := tdb.txDB.Get(append(bTxPrefix, bTx...))
	if err != nil {
		return nil, fmt.Errorf("failed to Get the tx: %v", err)
	}
	if len(txData) == 0 {
		return nil, fmt.Errorf("failed to Get the tx: not found")
	}
	err = tx.Decode(txData)
	if err != nil {
		return nil, fmt.Errorf("failed to Decode the tx: %v", err)
	}
	return &tx, nil
}

// HasTx checks if database has tx.
func (tdb *TxDBImpl) HasTx(hash []byte) (bool, error) {
	return tdb.txDB.Has(append(txPrefix, hash...))
}

// GetReceipt gets receipt with receipt's hash
func (tdb *TxDBImpl) GetReceipt(hash []byte) (*tx.TxReceipt, error) {
	bReHash, err := tdb.txDB.Get(append(receiptPrefix, hash...))
	if err != nil {
		return nil, fmt.Errorf("failed to Get the receipt: %v", err)
	}
	reData, err := tdb.txDB.Get(append(bReceiptPrefix, bReHash...))
	if err != nil {
		return nil, fmt.Errorf("failed to Get the receipt: %v", err)
	}
	if len(reData) == 0 {
		return nil, fmt.Errorf("failed to Get the receipt: not found")
	}
	re := tx.TxReceipt{}
	err = re.Decode(reData)
	if err != nil {
		return nil, fmt.Errorf("failed to Decode the receipt: %v", err)
	}
	return &re, nil
}

// GetReceiptByTxHash gets receipt with tx's hash
func (tdb *TxDBImpl) GetReceiptByTxHash(hash []byte) (*tx.TxReceipt, error) {
	bReHash, err := tdb.txDB.Get(append(txReceiptPrefix, hash...))
	if err != nil {
		return nil, fmt.Errorf("failed to Get the receipt hash: %v", err)
	}
	reData, err := tdb.txDB.Get(append(bReceiptPrefix, bReHash...))
	if err != nil {
		return nil, fmt.Errorf("failed to Get the receipt: %v", err)
	}
	if len(reData) == 0 {
		return nil, fmt.Errorf("failed to Get the receipt: not found")
	}
	re := tx.TxReceipt{}
	err = re.Decode(reData)
	if err != nil {
		return nil, fmt.Errorf("failed to Decode the receipt: %v", err)
	}
	return &re, nil
}

// HasReceipt checks if database has receipt.
func (tdb *TxDBImpl) HasReceipt(hash []byte) (bool, error) {
	return tdb.txDB.Has(append(receiptPrefix, hash...))
}

// Close is close database
func (tdb *TxDBImpl) Close() {
	tdb.txDB.Close()
}
