package tx

import (
	"encoding/binary"
	"fmt"
	"sync"

	"github.com/iost-official/prototype/db"
)

type TxPoolDb struct {
	db db.Database
}

var txPrefix = []byte("t") //txPrefix+tx hash -> tx data
var PNPrefix = []byte("p")

var TxDb TxPool
var once sync.Once

var LdbPath string

func TxDbInstance() TxPool {
	if TxDb != nil {
		return TxDb
	}
	ldb, err := db.NewLDBDatabase(LdbPath+"txDB", 0, 0)
	if err != nil {
		panic(err)
	}
	once.Do(func() {
		TxDb = &TxPoolDb{
			db: ldb,
		}
	})

	return TxDb
}

//Add tx to db
func (tp *TxPoolDb) Add(tx *Tx) error {
	hash := tx.Hash()
	err := tp.db.Put(append(txPrefix, hash...), tx.Encode())
	if err != nil {
		return fmt.Errorf("failed to Put hash->tx: %v", err)
	}
	//no need to check Pblisher here,it was checked earlier
	PubKey := tx.Publisher.Pubkey
	NonceRaw := make([]byte, 8)
	binary.BigEndian.PutUint64(NonceRaw, uint64(tx.Nonce))

	err = tp.db.Put(append(PNPrefix, append(NonceRaw, PubKey...)...), hash)

	//fmt.Println(append(PNPrefix, append(NonceRaw, PubRaw...)...))

	if err != nil {
		return fmt.Errorf("failed to Put NP->hash: %v", err)
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

//Get Tx by its Publisher and Nonce
func (tp *TxPoolDb) GetByPN(Nonce int64, PubKey []byte) (*Tx, error) {
	NonceRaw := make([]byte, 8)
	binary.BigEndian.PutUint64(NonceRaw, uint64(Nonce))
	hash, err := tp.db.Get(append(PNPrefix, append(NonceRaw, PubKey...)...))

	//fmt.Println(append(PNPrefix, append(NonceRaw, PubRaw...)...))

	if err != nil {

		return nil, fmt.Errorf("failed to Get the tx hash: %v", err)
	}
	tx, err := tp.Get(hash)
	if err != nil {
		return nil, err
	}
	return tx, nil
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
