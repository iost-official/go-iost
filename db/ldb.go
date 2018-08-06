package db

import (
	"sync"

	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/errors"
	"github.com/syndtr/goleveldb/leveldb/filter"
	"github.com/syndtr/goleveldb/leveldb/opt"
)

type LDB struct {
	fn       string
	db       *leveldb.DB
	quitLock sync.Mutex
}

var ldbMap map[string]*LDB
var mutex sync.RWMutex

func NewLDB(file string, cache int, handles int) (*LDB, error) {
	if ldbMap == nil {
		ldbMap = make(map[string]*LDB)
	}
	mutex.RLock()
	ldb, ok := ldbMap[file]
	mutex.RUnlock()
	if !ok {
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
		ldb = &LDB{
			fn: file,
			db: db,
		}
		mutex.Lock()
		ldbMap[file] = ldb
		mutex.Unlock()
	}
	return ldb, nil
}

func (ldb *LDB) Put(key []byte, value []byte) error {
	return ldb.db.Put(key, value, nil)
}

func (ldb *LDB) Get(key []byte) ([]byte, error) {
	value, err := ldb.db.Get(key, nil)
	if err != nil {
		return nil, err
	}
	return value, nil
}

func (ldb *LDB) Has(key []byte) (bool, error) {
	return ldb.db.Has(key, nil)
}

func (ldb *LDB) Delete(key []byte) error {
	return ldb.db.Delete(key, nil)
}

func (ldb *LDB) Close() {
	ldb.quitLock.Lock()
	defer ldb.quitLock.Unlock()
	ldb.db.Close()
}

func (ldb *LDB) Path() string {
	return ldb.fn
}

func (ldb *LDB) DB() *leveldb.DB {
	return ldb.db
}

func (ldb *LDB) Batch() *LDBBatch {
	return &LDBBatch{db: ldb.db, btch: new(leveldb.Batch)}
}

type LDBBatch struct {
	db   *leveldb.DB
	btch *leveldb.Batch
}

func (ldbBtch *LDBBatch) Reset() {
	ldbBtch.btch.Reset()
}

func (ldbBtch *LDBBatch) Put(key []byte, value []byte) error {
	ldbBtch.btch.Put(key, value)
	return nil
}

func (ldbBtch *LDBBatch) Commit() error {
	return ldbBtch.db.Write(ldbBtch.btch, nil)
}
