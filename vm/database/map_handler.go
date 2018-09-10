package database

import "strings"

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
	m.addField(key, field)
}

func (m *MapHandler) addField(key, field string) {
	if m.MHas(key, field) {
		return
	}
	s := m.db.Get(MapPrefix + key)
	if s == "n" {
		m.db.Put(MapPrefix+key, "@"+field)
	}
	s = s + "@" + field
	m.db.Put(MapPrefix+key, s)
}

func (m *MapHandler) delField(key, field string) {
	s := m.db.Get(MapPrefix + key)
	s2 := strings.Replace(s, "@"+field, "", 1)
	m.db.Put(MapPrefix+key, s2)
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
	s := m.db.Get(MapPrefix + key)
	return strings.Split(s, "@")[1:]
}

// MDel ...
func (m *MapHandler) MDel(key, field string) {
	if !m.MHas(key, field) {
		return
	}
	m.db.Del(MapPrefix + key + Separator + field)
	m.delField(key, field)
}
