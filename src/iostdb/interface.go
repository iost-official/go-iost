package iostdb

type Database interface {
	Put(key []byte, value []byte) error
	Get(key []byte) ([]byte, error)
	Has(key []byte) (bool, error)
	Delete(key []byte) error
	Close()
	NewBatch() Batch
}

type Batch interface {
	Put(key []byte, value []byte) error
	ValueSize() int
	Write() error
	Reset()
}
