package tx

import (
	"fmt"
	"sync"

	"github.com/iost-official/Go-IOS-Protocol/db"
)

//go:generate mockgen -destination ../mocks/mock_txdb.go -package core_mock github.com/iost-official/Go-IOS-Protocol/core/new_tx TxDB

type TxDB interface {
	Push(txs []*Tx) error
	Get(hash []byte) (*Tx, error)
	Has(tx *Tx) (bool, error)
}
type TxDBImpl struct {
	db *db.LDB
}

var txPrefix = []byte("t") //txPrefix+tx hash -> tx data
var PNPrefix = []byte("p")

var once sync.Once

var TxDBInst *TxDBImpl
var LdbPath string

func TxDBInstance() TxDB {
	if TxDBInst != nil {
		return TxDBInst
	}
	ldb, err := db.NewLDB(LdbPath+"txDB", 0, 0)
	if err != nil {
		panic(err)
	}
	once.Do(func() {
		TxDBInst = &TxDBImpl{
			db: ldb,
		}
	})
	return TxDBInst
}

func (tdb *TxDBImpl) Push(txs []*Tx) error {
	btch := tdb.db.Batch()
	for _, tx := range txs {
		hash := tx.Hash()
		btch.Put(append(txPrefix, hash...), tx.Encode())
	}
	return btch.Commit()
}

func (tdb *TxDBImpl) Get(hash []byte) (*Tx, error) {
	tx := Tx{}
	txData, err := tdb.db.Get(append(txPrefix, hash...))
	if err != nil {
		return nil, fmt.Errorf("failed to Get the tx: %v", err)
	}

	err = tx.Decode(txData)
	if err != nil {
		return nil, fmt.Errorf("failed to Decode the tx: %v", err)
	}
	return &tx, nil
}

func (tdb *TxDBImpl) Has(tx *Tx) (bool, error) {
	hash := tx.Hash()
	return tdb.db.Has(append(txPrefix, hash...))
}
