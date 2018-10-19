package database

// Access enum of type of access
type Access int

// const .
const (
	Read Access = iota
	Write
)

// Watcher of db access
type Watcher struct {
	m map[string]Access
	database
}

// Get record and get
func (r *Watcher) Get(key string) (value string) {
	if r.m[key] != Write {
		r.m[key] = Read
	}
	return r.database.Get(key)
}

// Put ...
func (r *Watcher) Put(key, value string) {
	r.m[key] = Write
	r.database.Put(key, value)
}

// Del ...
func (r *Watcher) Del(key string) {
	r.m[key] = Write
	r.database.Del(key)
}

// Has ...
func (r *Watcher) Has(key string) bool {
	if r.m[key] != Write {
		r.m[key] = Read
	}
	return r.database.Has(key)
}

// Map map the access of this watcher
func (r *Watcher) Map() map[string]Access {
	return r.m
}

// NewWatcher ...
func NewWatcher(db database) *Watcher {
	return &Watcher{
		m:        make(map[string]Access),
		database: db,
	}
}
