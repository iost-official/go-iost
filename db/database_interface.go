package db

import (
	"errors"
	//"io/ioutil"
	//"os"
)

type Database interface {
	Put(key []byte, value []byte) error
	PutHM(key []byte, args ...[]byte) error
	Get(key []byte) ([]byte, error)
	GetHM(key []byte, args ...[]byte) ([][]byte, error)
	Has(key []byte) (bool, error)
	Delete(key []byte) error
	Close()
	//NewBatch() Batch
}

func DatabaseFactor(target string) (Database, error) {
	switch target {
	case "redis":
		return NewRedisDatabase()
	case "ldb":
		//dirname, _ := ioutil.TempDir(os.TempDir(), "test_")
		dirname := "database"
		db, err := NewLDBDatabase(dirname, 0, 0)
		return db, err
	case "mem":
		db, err := NewMemDatabase()
		return db, err
	}
	return nil, errors.New("target Database not found")
}

func GetTx(hash []byte) (*Tx, error) {

	ldb, err := db.DatabaseFactor("ldb")
	if err != nil {

		return nil, fmt.Errorf("failed to init db %v", err)
	}
	defer ldb.Close()
	txData, err := ldb.Get(hash)
	if err != nil {

		return nil, fmt.Errorf("failed to Get the tx:", err)
	}
	RetTx := new(Tx)
	err = RetTx.Decode(txData)
	if err != nil {

		return nil, fmt.Errorf("failed to Decode the tx:", err)
	}
	return RetTx
}

/*
type Batch interface {
	Put(key []byte, value []byte) error
	ValueSize() int
	Write() error
	Reset()
}
*/
