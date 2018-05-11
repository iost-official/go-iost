package tx

import (
	"fmt"

	"github.com/iost-official/prototype/common"
	"github.com/iost-official/prototype/db"
)

type TxPool interface {
	Add(tx *Tx) error
	Del(tx *Tx) error
	Get(hash []byte) (*Tx, error)
	Has(tx *Tx) (bool, error)
	Pop() (*Tx, error)
	Size() int
}

type TxPoolImpl struct {
	txMap map[string]*Tx
}

func NewTxPoolImpl() *TxPoolImpl {
	return &TxPoolImpl{txMap: make(map[string]*Tx)}
}

func (tp *TxPoolImpl) Add(tx *Tx) error {
	tp.txMap[common.Base58Encode(tx.Hash())] = tx
	return nil
}

func (tp *TxPoolImpl) Del(tx *Tx) error {
	delete(tp.txMap, common.Base58Encode(tx.Hash()))
	return nil
}

func (tp *TxPoolImpl) Get(hash []byte) (*Tx, error) {
	tx, _ := tp.txMap[common.Base58Encode(hash)]
	return tx, nil
}

func (tp *TxPoolImpl) Pop() (*Tx, error) {
	return nil, nil
}

func (tp *TxPoolImpl) Has(tx *Tx) (bool, error) {
	_, ok := tp.txMap[common.Base58Encode(tx.Hash())]
	return ok, nil
}

func (tp *TxPoolImpl) Size() int {
	return len(tp.txMap)
}

type TxPoolDbImpl struct {
	db db.Database
}

var txPrefix = []byte("t") //txPrefix+tx hash -> tx data

/*
	if you call NewTxPoolDbImpl() to generate a instance tp of TxPoolDbImpl
	then you must call tp.Close() to close the db at last
*/
func NewTxPoolDbImpl() (TxPool, error) {
	ldb, err := db.DatabaseFactor("ldb")
	if err != nil {
		return nil, fmt.Errorf("failed to init db %v", err)
	}

	return &TxPoolDbImpl{db: ldb}, nil
}

func (tp *TxPoolDbImpl) Add(tx *Tx) error {
	hash := tx.Hash()
	err := tp.db.Put(append(txPrefix, hash...), tx.Encode())
	if err != nil {
		return fmt.Errorf("failed to Put tx: %v", err)
	}
	return nil
}

func (tp *TxPoolDbImpl) Del(tx *Tx) error {
	return nil
}
/*
func (tp *TxPoolDbImpl) Get(hash []byte) (*Tx, error) {
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
*/
//todo
func (tp *TxPoolDbImpl) Has(tx *Tx) (bool, error) {
	return false, nil
}

func (tp *TxPoolDbImpl) Size() int {
	return 0
}

func (tp *TxPoolDbImpl) Close() {
	tp.db.Close()
}
func Pop() (*Tx, error) {
	return nil, nil
}

func (tp *TxPoolDbImpl) Pop() (*Tx, error) {
	return nil, nil
}
