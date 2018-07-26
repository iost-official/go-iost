package db

import (
	"sync"

	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/errors"
	"github.com/syndtr/goleveldb/leveldb/filter"
	"github.com/syndtr/goleveldb/leveldb/iterator"
	"github.com/syndtr/goleveldb/leveldb/opt"
)

type LDBDatabase struct {
	fn string
	db *leveldb.DB

	quitLock sync.Mutex
	quitChan chan chan error
}

var ldbMap map[string]*LDBDatabase
var mutex sync.Mutex

func NewLDBDatabase(file string, cache int, handles int) (*LDBDatabase, error) {
	if ldbMap == nil {
		ldbMap = make(map[string]*LDBDatabase)
	}
	mutex.Lock()
	if _, ok := ldbMap[file]; !ok {
		if cache < 16 {
			cache = 16
		}
		if handles < 16 {
			handles = 16
		}
		db, err := leveldb.OpenFile(file, &opt.Options{
			OpenFilesCacheCapacity: handles,
			BlockCacheCapacity:     cache / 2 * opt.MiB,
			WriteBuffer:            cache / 4 * opt.MiB,
			Filter:                 filter.NewBloomFilter(10),
		})

		//fmt.Println(file, err)
		if _, corrupted := err.(*errors.ErrCorrupted); corrupted {
			db, err = leveldb.RecoverFile(file, nil)
		}
		if err != nil {
			return nil, err
		}
		ldbMap[file] = &LDBDatabase{
			fn: file,
			db: db,
		}
	}
	mutex.Unlock()
	return ldbMap[file], nil
}

func (db *LDBDatabase) Path() string {
	return db.fn
}

func (db *LDBDatabase) Put(key []byte, value []byte) error {
	return db.db.Put(key, value, nil)
}

func (db *LDBDatabase) PutHM(key []byte, args ...[]byte) error {
	return errors.New("Unsupported")
}

func (db *LDBDatabase) Get(key []byte) ([]byte, error) {
	value, err := db.db.Get(key, nil)
	if err != nil {
		return nil, err
	}
	return value, nil
}

func (db *LDBDatabase) GetHM(key []byte, args ...[]byte) ([][]byte, error) {
	return nil, errors.New("Unsupported")
}

func (db *LDBDatabase) Has(key []byte) (bool, error) {
	return db.db.Has(key, nil)
}

func (db *LDBDatabase) Delete(key []byte) error {
	return db.db.Delete(key, nil)
}

func (db *LDBDatabase) NewIterator() iterator.Iterator {
	return db.db.NewIterator(nil, nil)
}

func (db *LDBDatabase) Close() {
	db.quitLock.Lock()
	defer db.quitLock.Unlock()
	db.db.Close()
}

func (db *LDBDatabase) DB() *leveldb.DB {
	return db.db
}

func (db *LDBDatabase) IsEmpty() (bool, error) {
	isEmpty := true
	iter := db.NewIterator()
	for iter.Next() {
		isEmpty = false
		break
	}
	iter.Release()
	err := iter.Error()

	return isEmpty, err
}
