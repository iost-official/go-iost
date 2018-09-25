package database

//go:generate mockgen -destination mvcc_mock.go -package database github.com/iost-official/Go-IOS-Protocol/vm/database IMultiValue

// IMultiValue mvcc database interface
type IMultiValue interface {
	Get(table string, key string) (string, error)
	Put(table string, key string, value string) error
	Del(table string, key string) error
	Has(table string, key string) (bool, error)
	Commit()
	Rollback()
}
