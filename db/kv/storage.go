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
	Close() error
	Write(batch interface{}) error
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

func (s *Storage) Keys(prefix []byte) ([][]byte, error) {
	return nil, nil
}

func (s *Storage) Close() error {
	return nil
}

func (s *Storage) Write(batch Batch) error {
	return nil
}

func (s *Storage) NewBatch() *Batch {
	return nil
}

type BatchBackend interface {
	Put(key []byte, value []byte) error
	Delete(key []byte) error
}

type Batch struct {
	BatchBackend
}

func newBatch(t StorageType) *Batch {
	return nil
}
