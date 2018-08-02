package database

//go:generate mockgen -destination mvcc_mock_test.go -package database github.com/iost-official/Go-IOS-Protocol/new_vm/database IMultiValue

type IMultiValue interface {
	Get(table string, key string) (string, error)
	Put(table string, key string, value string) error
	Del(table string, key string) error
	Has(table string, key string) (bool, error)
	Keys(table string, prefix string) ([]string, error)
	Tables(table string) ([]string, error)
	Commit() (string, error)
	Rollback() error
	Tag(tag string) error
	Fork(revision string) string
	Checkout(revision string) string
	Flush(revision string) error
}
