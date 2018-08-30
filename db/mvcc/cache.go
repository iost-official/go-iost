package mvcc

type Value interface{}

type Cache interface {
	Get(key []byte) interface{}
	Put(key []byte, value interface{})
	Free()
}

func NewCache() Cache {
	return nil
}
