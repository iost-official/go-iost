package state

// state pool存储对状态机进行修改的链，可以方便地记录所有的状态同时易于回退
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
