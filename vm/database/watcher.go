package database

type Access int

const (
	Read Access = iota
	Write
)

type Watcher struct {
	m map[string]Access
	database
}

func (r *Watcher) Get(key string) (value string) {
	r.m[key] = Read
	return r.database.Get(key)
}
func (r *Watcher) Put(key, value string) {
	r.m[key] = Write
	r.database.Put(key, value)
}
func (r *Watcher) Del(key string) {
	r.m[key] = Read
	r.database.Del(key)
}
func (r *Watcher) Has(key string) bool {
	r.m[key] = Read
	return r.database.Has(key)
}
func (r *Watcher) Map() map[string]Access {
	return r.m
}

func NewWatcher(db database) *Watcher {
	return &Watcher{
		m:        make(map[string]Access),
		database: db,
	}
}
