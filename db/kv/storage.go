package kv

import (
	"github.com/iost-official/go-iost/db/kv/leveldb"
)

// StorageType is the type of storage, include leveldb and rocksdb
type StorageType uint8

// Storage type constant
const (
	_ StorageType = iota
	LevelDBStorage
)

// StorageBackend is the storage backend interface
type StorageBackend interface {
	Get(key []byte) ([]byte, error)
	Put(key []byte, value []byte) error
	Has(key []byte) (bool, error)
	Delete(key []byte) error
	Keys(prefix []byte) ([][]byte, error)
	KeysByRange(from []byte, to []byte, limit int) ([][]byte, error)
	BeginBatch() error
	CommitBatch() error
	Size() (int64, error)
	Close() error
	NewIteratorByPrefix(prefix []byte) interface{}
}

// Storage is a kv database
type Storage struct {
	StorageBackend
}

// NewStorage return the storage of the specify type
func NewStorage(path string, t StorageType) (*Storage, error) {
	switch t {
	case LevelDBStorage:
		sb, err := leveldb.NewDB(path)
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

// NewIteratorByPrefix returns a new iterator by prefix
func (s *Storage) NewIteratorByPrefix(prefix []byte) *Iterator {
	ib := s.StorageBackend.NewIteratorByPrefix(prefix).(IteratorBackend)
	return &Iterator{
		IteratorBackend: ib,
	}
}

// IteratorBackend is the storage iterator backend
type IteratorBackend interface {
	Next() bool
	Key() []byte
	Value() []byte
	Error() error
	Release()
}

// Iterator is the storage iterator
type Iterator struct {
	IteratorBackend
}
