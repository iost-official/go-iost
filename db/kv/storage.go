package kv

import (
	"github.com/iost-official/Go-IOS-Protocol/db/kv/leveldb"
	"github.com/iost-official/Go-IOS-Protocol/db/kv/rocksdb"
)

type StorageType uint

const (
	_ StorageType = iota
	LevelDBStorage
	RocksDBStorage
)

type StorageBackend interface {
	Get(key []byte) ([]byte, error)
	Put(key []byte, value []byte) error
	Has(key []byte) (bool, error)
	Delete(key []byte) error
	Keys(prefix []byte) ([][]byte, error)
	BeginBatch() error
	CommitBatch() error
	Close() error
}

type Storage struct {
	StorageBackend
}

func NewStorage(path string, t StorageType) (*Storage, error) {
	switch t {
	case LevelDBStorage:
		sb, err := leveldb.NewDB(path)
		if err != nil {
			return nil, err
		}
		return &Storage{StorageBackend: sb}, nil
	case RocksDBStorage:
		sb, err := rocksdb.NewDB(path)
		if err != nil {
			return nil, err
		}
		return &Storage{StorageBackend: sb}, nil
	default:
		sb, err := leveldb.NewDB(path)
		if err != nil {
			return nil, err
		}
		return &Storage{StorageBackend: sb}, nil
	}
}
