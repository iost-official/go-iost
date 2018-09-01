package tx

import (
	"fmt"

	"github.com/iost-official/Go-IOS-Protocol/db"
)

//go:generate mockgen -destination ../mocks/mock_txdb.go -package core_mock github.com/iost-official/Go-IOS-Protocol/core/tx TxDB

type TxDB interface {
	Push(txs []*Tx, receipts []*TxReceipt) error
	GetTx(hash []byte) (*Tx, error)
	HasTx(hash []byte) (bool, error)
	GetReceipt(Hash []byte) (*TxReceipt, error)
	GetReceiptByTxHash(Hash []byte) (*TxReceipt, error)
	HasReceipt(hash []byte) (bool, error)
}
type TxDBImpl struct {
	txDB *db.LDB
}

var (
	txPrefix          = []byte("t") // txPrefix+tx hash -> tx data
	receiptHashPrefix = []byte("h") // receiptHashPrefix + tx hash -> receipt hash
	receiptPrefix     = []byte("r") // receiptPrefix + receipt hash -> receipt data
)

func NewTxDB(path string) (TxDB, error) {
	ldb, err := db.NewLDB(path, 0, 0)
	if err != nil {
		return nil, err
	}
	return &TxDBImpl{txDB: ldb}, nil
}

func (tdb *TxDBImpl) Push(txs []*Tx, receipts []*TxReceipt) error {
	txBth := tdb.txDB.Batch()

	for i, tx := range txs {
		tHash := tx.Hash()
		txBth.Put(append(txPrefix, tHash...), tx.Encode())

		// save receipt
		rHash := receipts[i].Hash()
		txBth.Put(append(receiptHashPrefix, tHash...), rHash)

		txBth.Put(append(receiptPrefix, rHash...), receipts[i].Encode())
	}

	return txBth.Commit()
}

func (tdb *TxDBImpl) GetTx(hash []byte) (*Tx, error) {
	tx := Tx{}
	txData, err := tdb.txDB.Get(append(txPrefix, hash...))
	if err != nil {
		return nil, fmt.Errorf("failed to Get the tx: %v", err)
	}

	err = tx.Decode(txData)
	if err != nil {
		return nil, fmt.Errorf("failed to Decode the tx: %v", err)
	}
	return &tx, nil
}

func (tdb *TxDBImpl) HasTx(hash []byte) (bool, error) {

	return tdb.txDB.Has(append(txPrefix, hash...))
}

func (tdb *TxDBImpl) GetReceipt(Hash []byte) (*TxReceipt, error) {
	re := TxReceipt{}
	reData, err := tdb.txDB.Get(append(receiptPrefix, Hash...))
	if err != nil {
		return nil, fmt.Errorf("failed to Get the receipt: %v", err)
	}

	err = re.Decode(reData)
	if err != nil {
		return nil, fmt.Errorf("failed to Decode the receipt: %v", err)
	}
	return &re, nil
}

func (tdb *TxDBImpl) GetReceiptByTxHash(Hash []byte) (*TxReceipt, error) {

	reHash, err := tdb.txDB.Get(append(receiptHashPrefix, Hash...))
	if err != nil {
		return nil, fmt.Errorf("failed to Get the receipt hash: %v", err)
	}

	return tdb.GetReceipt(reHash)
}

func (tdb *TxDBImpl) HasReceipt(hash []byte) (bool, error) {

	return tdb.txDB.Has(append(receiptPrefix, hash...))
}

func (tdb *TxDBImpl) Close() {
	tdb.txDB.Close()
}
