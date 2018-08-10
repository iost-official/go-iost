package storage

import (
	"errors"

	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/util"
)

type LevelDB struct {
	db    *leveldb.DB
	batch *leveldb.Batch
}

func NewLevelDB(path string) (*LevelDB, error) {
	db, err := leveldb.OpenFile(path, nil)
	if err != nil {
		return nil, err
	}
	ldb := &LevelDB{
		db:    db,
		batch: nil,
	}
	return ldb, nil
}

func (d *LevelDB) Get(key []byte) ([]byte, error) {
	return d.db.Get(key, nil)
}

func (d *LevelDB) Put(key []byte, value []byte) error {
	if d.batch == nil {
		return d.db.Put(key, value, nil)
	}
	d.batch.Put(key, value)
	return nil
}

func (d *LevelDB) Del(key []byte) error {
	if d.batch == nil {
		return d.db.Delete(key, nil)
	}
	d.batch.Delete(key)
	return nil
}

func (d *LevelDB) Keys(prefix []byte) ([][]byte, error) {
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

func (d *LevelDB) BeginBatch() error {
	if d.batch != nil {
		return errors.New("not support nested batch write")
	}
	d.batch = new(leveldb.Batch)
	return nil
}

func (d *LevelDB) CommitBatch() error {
	if d.batch == nil {
		return errors.New("no batch write to commit")
	}
	err := d.db.Write(d.batch, nil)
	if err != nil {
		return err
	}
	d.batch = nil
	return nil
}

func (d *LevelDB) Close() error {
	return d.db.Close()
}
