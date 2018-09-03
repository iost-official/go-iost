package kv

type StorageType string

const (
	LevelDBStorage StorageType = "leveldb"
	RocksDBStorage StorageType = "rocksdb"
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

func NewStorage(t StorageType) *Storage {
	return nil
}

func (s *Storage) Get(key []byte) ([]byte, error) {
	return nil, nil
}

func (s *Storage) Put(key []byte, value []byte) error {
	return nil
}

func (s *Storage) Has(key []byte) (bool, error) {
	return false, nil
}

func (s *Storage) Delete(key []byte) error {
	return nil
}

func (s *Storage) Keys(prefix []byte) ([][]byte, error) {
	return nil, nil
}

func (s *Storage) BeginBatch() error {
	return nil
}

func (s *Storage) CommitBatch() error {
	return nil
}

func (s *Storage) Close() error {
	return nil
}
