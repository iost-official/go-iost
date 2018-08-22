package database

// BasicPrefix ...
const BasicPrefix = "b-"

// BasicHandler ...
type BasicHandler struct {
	db database
}

// Put ...
func (m *BasicHandler) Put(key, value string) {
	m.db.Put(BasicPrefix+key, value)
}

// Get ...
func (m *BasicHandler) Get(key string) (value string) {
	return m.db.Get(BasicPrefix + key)
}

// Has ...
func (m *BasicHandler) Has(key string) bool {
	return m.db.Has(BasicPrefix + key)
}

// Del ...
func (m *BasicHandler) Del(key string) {
	m.db.Del(BasicPrefix + key)
}
