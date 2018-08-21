package database

const BasicPrefix = "b-"

type BasicHandler struct {
	db Database
}

func (m *BasicHandler) Put(key, value string) {
	m.db.Put(BasicPrefix+key, value)
}

func (m *BasicHandler) Get(key string) (value string) {
	return m.db.Get(BasicPrefix + key)
}

func (m *BasicHandler) Has(key string) bool {
	return m.db.Has(BasicPrefix + key)
}

func (m *BasicHandler) Del(key string) {
	m.db.Del(BasicPrefix + key)
}
