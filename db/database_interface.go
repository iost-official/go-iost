package db

import (
	"errors"
	//"io/ioutil"
	//"os"
)

//go:generate mockgen -destination mocks/mock_database.go -package db_mock github.com/iost-official/Go-IOS-Protocol/db Database

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

func DatabaseFactory(target string) (Database, error) {
	switch target {
	case "redis":
		return NewRedisDatabase()
	case "mem":
		db, err := NewMemDatabase()
		return db, err
	}
	return nil, errors.New("target Database not found")
}
