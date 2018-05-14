package tx

import (
	"fmt"

	"github.com/iost-official/prototype/db"
)

type TxPoolDb struct {
	db db.Database
}

var txPrefix = []byte("t") //txPrefix+tx hash -> tx data

func NewTxPoolDb() (TxPool, error) {
	ldb, err := db.DatabaseFactor("ldb")
	if err != nil {
		return nil, fmt.Errorf("failed to init db %v", err)
	}

	return &TxPoolDb{db: ldb}, nil
}

//Add tx to db
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

//Get tx from db
func (tp *TxPoolDb) Get(hash []byte) (*Tx, error) {
	tx := Tx{}
	txData, err := tp.db.Get(append(txPrefix, hash...))
	if err != nil {

		return nil, fmt.Errorf("failed to Get the tx: %v", err)
	}

	err = tx.Decode(txData)
	if err != nil {

		return nil, fmt.Errorf("failed to Decode the tx: %v", err)
	}
	return &tx, nil
}

// 判断一个Tx是否在Tx Pool
func (tp *TxPoolDb) Has(tx *Tx) (bool, error) {
	hash := tx.Hash()
	return tp.db.Has(append(txPrefix, hash...))
}

// 获取TxPool中tx的数量
func (tp *TxPoolDb) Size() int {
	return 0
}

/*
no need to Close ldb any more,cause we changed db.DatabaseFactor() to sync.Once.
So,the ldb would be always open...
func (tp *TxPoolDb) Close() {
	tp.db.Close()
}
*/

// 在Tx Pool 获取第一个Tx
func (tp *TxPoolDb) Top() (*Tx, error) {
	return nil, nil
}
