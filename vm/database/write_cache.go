package database

import "github.com/iost-official/go-iost/v3/core/version"

// WriteCache ...
type WriteCache struct {
	m  map[string]*Record
	db database
}

// Mode of this record
type Mode int

// const enum of Mode
const (
	Default Mode = iota
	Delete
)

// Record of one access
type Record struct {
	value string
	mode  Mode
}

// NewWriteCache ...
func NewWriteCache(db database) *WriteCache {
	return &WriteCache{
		m:  make(map[string]*Record),
		db: db,
	}
}

// Rules ...
func (w *WriteCache) Rules() *version.Rules {
	return w.db.Rules()
}

// Get ...
func (w *WriteCache) Get(key string) (value string) {
	if v, ok := w.m[key]; ok {
		if v.mode == Delete {
			return "n"
		}
		return v.value
	}
	return w.db.Get(key)
}

// Put ...
func (w *WriteCache) Put(key, value string) {
	w.m[key] = &Record{
		value: value,
		mode:  Default,
	}
}

// Has ...
func (w *WriteCache) Has(key string) bool {
	if v, ok := w.m[key]; ok {
		return v.mode != Delete
	}
	return w.db.Has(key)
}

// Del ...
func (w *WriteCache) Del(key string) {
	w.m[key] = &Record{
		value: "",
		mode:  Delete,
	}
}

// Flush ...
func (w *WriteCache) Flush() {
	for k, v := range w.m {
		if v.mode == Delete {
			w.db.Del(k)
		} else {
			w.db.Put(k, v.value)
		}
	}
}

// Drop ...
func (w *WriteCache) Drop() {
	w.m = make(map[string]*Record)
}
