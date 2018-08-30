package mvccmap

import "strings"

type MVCCMap struct {
	data   map[string]interface{}
	parent *MVCCMap
	refs   []*MVCCMap
}

func New() *MVCCMap {
	return &MVCCMap{
		data:   make(map[string]interface{}),
		parent: nil,
		refs:   make([]*MVCCMap, 0),
	}
}

func (m *MVCCMap) Get(key []byte) interface{} {
	v, ok := m.data[string(key)]
	if !ok {
		if m.parent == nil {
			return nil
		}
		return m.parent.Get(key)
	}
	return v
}

func (m *MVCCMap) Put(key []byte, value interface{}) {
	m.data[string(key)] = value
}

func (m *MVCCMap) All(prefix []byte) []interface{} {
	values := make([]interface{}, 0)
	for k, v := range m.data {
		if strings.HasPrefix(string(k), string(prefix)) {
			values = append(values, v)
		}
	}
	if m.parent == nil {
		return values
	}
	return append(m.parent.All(prefix), values...)
}

func (m *MVCCMap) Fork() interface{} {
	mvccmap := &MVCCMap{
		data:   make(map[string]interface{}),
		parent: m,
		refs:   make([]*MVCCMap, 0),
	}
	m.refs = append(m.refs, mvccmap)
	return mvccmap
}

func (m *MVCCMap) Free() {
	if m.parent != nil {
		m.parent.Free()
	}
	for _, ref := range m.refs {
		ref.parent = nil
	}
	m.parent = nil
	m.refs = nil
	m.data = nil
}
