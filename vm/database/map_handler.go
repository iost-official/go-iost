package database

import (
	"strings"

	"github.com/iost-official/go-iost/common"
)

// MapHandler handler of map
type MapHandler struct {
	db database
}

// MapPrefix prefix of map key
const MapPrefix = "m-"

// Separator separator of map key
const Separator = "-"

// ApplicationSeparator separator of value
const ApplicationSeparator = "@"

// MPut put value in kfv storage o(1)
func (m *MapHandler) MPut(key, field, value string) {
	//fmt.Println("map put,", key, field, value)
	m.addField(key, field)
	m.db.Put(MapPrefix+key+Separator+field, value)
}

func (m *MapHandler) addField(key, field string) {
	if getHeight(m.db) >= common.Ver3_0_3 {
		m.addField303(key, field)
		return
	}

	if m.MHas(key, field) {
		return
	}
	s := m.db.Get(MapPrefix + key)
	if s == "n" {
		m.db.Put(MapPrefix+key, ApplicationSeparator+field)
		return
	}

	if strings.Count(s, ApplicationSeparator) > 256 {
		return
	}

	s = s + ApplicationSeparator + field
	m.db.Put(MapPrefix+key, s)
}

func (m *MapHandler) addField303(key, field string) {
	if m.MHas(key, field) {
		return
	}
	s := m.clearField303(key)
	if s == "n" {
		m.db.Put(MapPrefix+key, ApplicationSeparator+ApplicationSeparator+field)
		return
	}

	if strings.Count(s, ApplicationSeparator) > 257 {
		return
	}

	s = s + ApplicationSeparator + field
	m.db.Put(MapPrefix+key, s)
	return
}

func (m *MapHandler) delField(key, field string) {
	if getHeight(m.db) >= common.Ver3_0_3 {
		m.delField303(key, field)
		return
	}

	s := m.db.Get(MapPrefix + key)
	s2 := strings.Replace(s, ApplicationSeparator+field, "", 1)
	if s2 == "" {
		m.db.Del(MapPrefix + key)
		return
	}
	m.db.Put(MapPrefix+key, s2)
}

func (m *MapHandler) delField303(key, field string) {
	s := m.clearField303(key)
	var s2 string
	if strings.HasSuffix(s, ApplicationSeparator+field) {
		s2 = s[:len(s)-len(ApplicationSeparator+field)]
	} else {
		s2 = strings.Replace(s, ApplicationSeparator+field+ApplicationSeparator, ApplicationSeparator, 1)
	}
	if s2 == "@" {
		m.db.Del(MapPrefix + key)
		return
	}
	m.db.Put(MapPrefix+key, s2)

	return
}

func (m *MapHandler) clearField303(key string) string {
	s := m.db.Get(MapPrefix + key)
	if s == "n" {
		return s
	}

	if !strings.HasPrefix(s, "@@") {
		var sb strings.Builder
		var same = make(map[string]struct{})

		for _, f := range strings.Split(s, ApplicationSeparator)[1:] {
			if m.MHas(key, f) {
				if _, ok := same[f]; !ok {
					sb.WriteString(ApplicationSeparator + f)
					same[f] = struct{}{}
				}
			}
		}

		return ApplicationSeparator + sb.String()
	}
	return s
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
	if getHeight(m.db) >= common.Ver3_0_3 {
		s := m.db.Get(MapPrefix + key)
		return strings.Split(s, ApplicationSeparator)[2:]
	}
	s := m.db.Get(MapPrefix + key)
	return strings.Split(s, ApplicationSeparator)[1:]
}

// MDel delete field of map o(1)
func (m *MapHandler) MDel(key, field string) {
	if !m.MHas(key, field) {
		return
	}
	m.db.Del(MapPrefix + key + Separator + field)
	m.delField(key, field)
}
