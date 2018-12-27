package leveldb

import (
	"fmt"

	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/iterator"
	"github.com/syndtr/goleveldb/leveldb/util"
)

// DB is the leveldb databse
type DB struct {
	db    *leveldb.DB
	batch *leveldb.Batch
}

// NewDB return new leveldb
func NewDB(path string) (*DB, error) {
	db, err := leveldb.OpenFile(path, nil)
	if err != nil {
		return nil, err
	}
	return &DB{
		db:    db,
		batch: nil,
	}, nil
}

// Get return the value of the specify key
func (d *DB) Get(key []byte) ([]byte, error) {
	value, err := d.db.Get(key, nil)
	if err == leveldb.ErrNotFound {
		return []byte{}, nil
	}

	return value, err
}

// Has returns whether the specified key exists
func (d *DB) Has(key []byte) (bool, error) {
	return d.db.Has(key, nil)
}

// Put will insert the key-value pair
func (d *DB) Put(key []byte, value []byte) error {
	if d.batch == nil {
		return d.db.Put(key, value, nil)
	}
	d.batch.Put(key, value)
	return nil
}

// Delete will remove the specify key
func (d *DB) Delete(key []byte) error {
	if d.batch == nil {
		return d.db.Delete(key, nil)
	}
	d.batch.Delete(key)
	return nil
}

// Keys returns the list of key prefixed with prefix
func (d *DB) Keys(prefix []byte) ([][]byte, error) {
	iter := d.db.NewIterator(util.BytesPrefix(prefix), nil)
	keys := make([][]byte, 0)

	for iter.Next() {
		key := make([]byte, len(iter.Key()))
		copy(key, iter.Key())
		keys = append(keys, key)
	}
	iter.Release()
	err := iter.Error()
	if err != nil {
		return nil, err
	}
	return keys, nil
}

// BeginBatch will start the batch transaction
func (d *DB) BeginBatch() error {
	if d.batch != nil {
		return fmt.Errorf("not support nested batch write")
	}
	d.batch = new(leveldb.Batch)
	return nil
}

// CommitBatch will commit the batch transaction
func (d *DB) CommitBatch() error {
	if d.batch == nil {
		return fmt.Errorf("no batch write to commit")
	}
	err := d.db.Write(d.batch, nil)
	if err != nil {
		return err
	}
	d.batch = nil
	return nil
}

// Size returns the size of leveldb
func (d *DB) Size() (int64, error) {
	stats := &leveldb.DBStats{}
	if err := d.db.Stats(stats); err != nil {
		return 0, err
	}
	total := int64(0)
	for _, size := range stats.LevelSizes {
		total += size
	}
	return total, nil
}

// Close will close the database
func (d *DB) Close() error {
	return d.db.Close()
}

// NewIteratorByPrefix returns a new iterator by prefix
func (d *DB) NewIteratorByPrefix(prefix []byte) interface{} {
	iter := d.db.NewIterator(util.BytesPrefix(prefix), nil)
	return &Iter{
		iter: iter,
	}
}

// Iter is the iterator for leveldb
type Iter struct {
	iter iterator.Iterator
}

// Next do next item of iterator
func (i *Iter) Next() bool {
	return i.iter.Next()
}

// Key returns the key of current item
func (i *Iter) Key() []byte {
	return i.iter.Key()
}

// Value returns the value of current item
func (i *Iter) Value() []byte {
	return i.iter.Value()
}

// Error returns the error of iterator
func (i *Iter) Error() error {
	return i.iter.Error()
}

// Release will release the iterator
func (i *Iter) Release() {
	i.iter.Release()
}
