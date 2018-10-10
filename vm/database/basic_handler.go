package database

// BasicPrefix prefix of basic types
const BasicPrefix = "b-"

// BasicHandler handler of basic type
type BasicHandler struct {
	db database
}

// Put put to k-v
func (m *BasicHandler) Put(key, value string) {
	m.db.Put(BasicPrefix+key, value)
}

// Get get v from k
func (m *BasicHandler) Get(key string) (value string) {
	return m.db.Get(BasicPrefix + key)
}

// Has determine if k exist
func (m *BasicHandler) Has(key string) bool {
	return m.db.Has(BasicPrefix + key)
}

// Del del key, if key is nil do nothing
func (m *BasicHandler) Del(key string) {
	m.db.Del(BasicPrefix + key)
}

func (m *BasicHandler) PrintCache() {
	m.db.PrintCache()
}
