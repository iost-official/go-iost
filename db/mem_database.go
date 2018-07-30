package db

import (
	"errors"
	"sync"
)

func CopyBytes(b []byte) (copiedBytes []byte) {
	if b == nil {
		return nil
	}
	copiedBytes = make([]byte, len(b))
	copy(copiedBytes, b)
	return
}

type MemDatabase struct {
	db   map[string][]byte
	lock sync.RWMutex
}

func NewMemDatabase() (*MemDatabase, error) {
	return &MemDatabase{db: make(map[string][]byte)}, nil
}

func NewMemDatabaseWithCap(size int) (*MemDatabase, error) {
	return &MemDatabase{db: make(map[string][]byte, size)}, nil
}

func (db *MemDatabase) Put(key []byte, value []byte) error {
	db.lock.Lock()
	defer db.lock.Unlock()
	db.db[string(key)] = CopyBytes(value)
	return nil
}

func (db *MemDatabase) PutHM(key []byte, args ...[]byte) error {
	return errors.New("Unsupported")
}

func (db *MemDatabase) Get(key []byte) ([]byte, error) {
	db.lock.RLock()
	defer db.lock.RUnlock()
	if value, ok := db.db[string(key)]; ok {
		return CopyBytes(value), nil
	}
	return nil, errors.New("Not found")
}

func (db *MemDatabase) GetHM(key []byte, args ...[]byte) ([][]byte, error) {
	return nil, errors.New("Unsupported")
}

func (db *MemDatabase) Has(key []byte) (bool, error) {
	db.lock.RLock()
	defer db.lock.RUnlock()
	_, ok := db.db[string(key)]
	return ok, nil
}

func (db *MemDatabase) Keys() [][]byte {
	db.lock.RLock()
	defer db.lock.RUnlock()
	keys := [][]byte{}
	for key := range db.db {
		keys = append(keys, []byte(key))
	}
	return keys
}

func (db *MemDatabase) Delete(key []byte) error {
	db.lock.Lock()
	defer db.lock.Unlock()
	delete(db.db, string(key))
	return nil
}

func (db *MemDatabase) Close() {}
