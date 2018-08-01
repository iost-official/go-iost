package database

type MapHandler struct {
	db Database
}

const MapPrefix = "m-"
const Separator = "-"

func (m *MapHandler) MPut(key, field, value string) {
	m.db.Put(MapPrefix+key+Separator+field, value)
}

func (m *MapHandler) MGet(key, field string) (value string) {
	return m.db.Get(MapPrefix + key + Separator + field)
}

func (m *MapHandler) MHas(key, field string) bool {
	return m.db.Has(MapPrefix + key + Separator + field)
}

func (m *MapHandler) MKeys(key string) (fields []string) {
	prefixLen := len(MapPrefix + key + Separator)
	rawKeys := m.db.Keys(MapPrefix + key + Separator)

	fields = make([]string, 0)
	for _, k := range rawKeys {
		fields = append(fields, k[prefixLen:])
	}
	return
}

func (m *MapHandler) MDel(key, field string) {
	m.db.Del(MapPrefix + key + Separator + field)
}
