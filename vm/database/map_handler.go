package database

import (
	"strings"
)

// MapHandler handler of map
type MapHandler struct {
	db database
}

// MapPrefix prefix of map key
const MapPrefix = "m-"

// Separator separator of map key
const Separator = "-"

// These two values are same... a historical mistake
const RAMOwnerSeparator = "@"

// MapKeysSeparator separator of value
const MapKeysSeparator = "@"

// MPut put value in kfv storage o(1)
func (m *MapHandler) MPut(key, field, value string) {
	//fmt.Println("map put,", key, field, value)
	m.addField(key, field)
	m.db.Put(MapPrefix+key+Separator+field, value)
}

func (m *MapHandler) addField(key, field string) {
	if m.MHas(key, field) {
		return
	}
	s := m.db.Get(MapPrefix + key)
	if s == "n" {
		m.db.Put(MapPrefix+key, MapKeysSeparator+field)
		return
	}

	if strings.Count(s, MapKeysSeparator) > 256 {
		return
	}

	s = s + MapKeysSeparator + field
	m.db.Put(MapPrefix+key, s)
}

func (m *MapHandler) delField(key, field string) {
	s := m.db.Get(MapPrefix + key)
	fixed := m.db.Get(MapPrefix + key + MapKeysSeparator)
	fields := strings.Split(s, MapKeysSeparator)[1:]
	s2 := ""
	if fixed == "n" {
		for _, f := range fields {
			if f == field {
				continue
			}
			if m.MHas(key, f) {
				s2 = s2 + MapKeysSeparator + f
			}
		}
		m.db.Put(MapPrefix+key+MapKeysSeparator, "1")
	} else {
		for _, f := range fields {
			if f == field {
				continue
			}
			s2 = s2 + MapKeysSeparator + f
		}
	}
	if s2 == "" {
		m.db.Del(MapPrefix + key)
		return
	}
	m.db.Put(MapPrefix+key, s2)
}

// MGet get value from storage o(1)
func (m *MapHandler) MGet(key, field string) (value string) {
	//fmt.Println("map get,", key, field)
	return m.db.Get(MapPrefix + key + Separator + field)
}

// MHas if has map and field
func (m *MapHandler) MHas(key, field string) bool {
	return m.db.Has(MapPrefix + key + Separator + field)
}

// MKeys list fields of map o(1)
func (m *MapHandler) MKeys(key string) (fields []string) {
	s := m.db.Get(MapPrefix + key)
	return strings.Split(s, MapKeysSeparator)[1:]
}

// MDel delete field of map o(1)
func (m *MapHandler) MDel(key, field string) {
	if !m.MHas(key, field) {
		return
	}
	m.db.Del(MapPrefix + key + Separator + field)
	m.delField(key, field)
}
