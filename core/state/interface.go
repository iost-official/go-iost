/*
Package state, implements of key-value using in vm
*/
package state

//Pool state pool of local state machine
type Pool interface {
	Copy() Pool
	GetPatch() Patch
	Flush() error
	MergeParent() (Pool, error)

	Put(key Key, value Value)
	Get(key Key) (Value, error)
	Has(key Key) bool
	Delete(key Key)

	GetHM(key, field Key) (Value, error)
	PutHM(key, field Key, value Value) error
}
