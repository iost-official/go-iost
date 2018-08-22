package database

// MapHandler ...
type MapHandler struct {
	db database
}

// MapPrefix ...
const MapPrefix = "m-"

// Separator ...
const Separator = "-"

// MPut ...
func (m *MapHandler) MPut(key, field, value string) {
	m.db.Put(MapPrefix+key+Separator+field, value)
}

// MGet ...
func (m *MapHandler) MGet(key, field string) (value string) {
	return m.db.Get(MapPrefix + key + Separator + field)
}

// MHas ...
func (m *MapHandler) MHas(key, field string) bool {
	return m.db.Has(MapPrefix + key + Separator + field)
}

// MKeys ...
func (m *MapHandler) MKeys(key string) (fields []string) {
	prefixLen := len(MapPrefix + key + Separator)
	rawKeys := m.db.Keys(MapPrefix + key + Separator)

	fields = make([]string, 0)
	for _, k := range rawKeys {
		fields = append(fields, k[prefixLen:])
	}
	return
}

// MDel ...
func (m *MapHandler) MDel(key, field string) {
	m.db.Del(MapPrefix + key + Separator + field)
}
