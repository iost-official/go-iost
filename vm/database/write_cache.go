package database

type WriteCache struct {
	m  map[string]*Record
	db database
}

type Mode int

const (
	Default Mode = 0
	Delete  Mode = 1
)

type Record struct {
	value string
	mode  Mode
}

func NewWriteCache(db database) *WriteCache {
	return &WriteCache{
		m:  make(map[string]*Record),
		db: db,
	}
}

func (w *WriteCache) Get(key string) (value string) {
	if v, ok := w.m[key]; ok {
		if v.mode == Delete {
			return "n"
		}
		return v.value
	} else {
		return w.db.Get(key)
	}
}
func (w *WriteCache) Put(key, value string) {
	w.m[key] = &Record{
		value: value,
		mode:  Default,
	}
}
func (w *WriteCache) Has(key string) bool {
	if v, ok := w.m[key]; ok {
		return v.mode != Delete
	} else {
		return w.db.Has(key)
	}
}
func (w *WriteCache) Del(key string) {
	w.m[key] = &Record{
		value: "",
		mode:  Delete,
	}
}
func (w *WriteCache) Flush() {
	for k, v := range w.m {
		if v.mode == Delete {
			w.db.Del(k)
		} else {
			w.db.Put(k, v.value)
		}
	}
}
func (w *WriteCache) Drop() {
	w.m = make(map[string]*Record)
}
