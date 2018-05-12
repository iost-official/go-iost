package tx

import (
	"fmt"

	"github.com/iost-official/prototype/db"
)

type TxPoolDb struct {
	db db.Database
}

var txPrefix = []byte("t") //txPrefix+tx hash -> tx data

/*
	if you call NewTxPoolDb() to generate a instance tp of TxPoolDb
	then you must call tp.Close() to close the db at last
*/
func NewTxPoolDb() (TxPool, error) {
	ldb, err := db.DatabaseFactor("ldb")
	if err != nil {
		return nil, fmt.Errorf("failed to init db %v", err)
	}

	return &TxPoolDb{db: ldb}, nil
}

func (tp *TxPoolDb) Add(tx *Tx) error {
	hash := tx.Hash()
	err := tp.db.Put(append(txPrefix, hash...), tx.Encode())
	if err != nil {
		return fmt.Errorf("failed to Put tx: %v", err)
	}
	return nil
}

func (tp *TxPoolDb) Del(tx *Tx) error {
	return nil
}

func (tp *TxPoolDb) Get(hash []byte) (*Tx, error) {
	tx := Tx{}
	txData, err := tp.db.Get(append(txPrefix, hash...))
	if err != nil {

		return nil, fmt.Errorf("failed to Get the tx: %v", err)
	}

	err = tx.Decode(txData) //something go wrong when call txPtr.Decode()
	if err != nil {

		return nil, fmt.Errorf("failed to Decode the tx: %v", err)
	}
	return &tx, nil
}

//todo
func (tp *TxPoolDb) Has(tx *Tx) (bool, error) {
	return false, nil
}

func (tp *TxPoolDb) Size() int {
	return 0
}

func (tp *TxPoolDb) Close() {
	tp.db.Close()
}
func Pop() (*Tx, error) {
	return nil, nil
}

func (tp *TxPoolDb) Pop() (*Tx, error) {
	return nil, nil
}
