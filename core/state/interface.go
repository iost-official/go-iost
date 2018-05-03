package state

type StatePool interface {
	Copy() StatePool
	GetPatch() Patch
	Flush() error

	Put(key Key, value Value) error
	Get(key Key) (Value, error)
	Has(key Key) (bool, error)
	Delete(key Key) error

	GetHM(key, field Key) (Value, error)
	PutHM(key, field Key, value Value) error
}

