package state

type Pool interface {
	Copy() Pool
	GetPatch() Patch
	Flush() error

	Put(key Key, value Value)
	Get(key Key) (Value, error)
	Has(key Key) bool
	Delete(key Key)

	GetHM(key, field Key) (Value, error)
	PutHM(key, field Key, value Value) error
}


