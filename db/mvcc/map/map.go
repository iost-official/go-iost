package mvccmap

import "strings"

type Value interface {
}

type MVCCMap struct {
	data   map[string]Value
	parent *MVCCMap
	refs   []*MVCCMap
}

func New() *MVCCMap {
	return &MVCCMap{
		data:   make(map[string]Value),
		parent: nil,
		refs:   make([]*MVCCMap, 0),
	}
}

func (m *MVCCMap) Get(key []byte) Value {
	v, ok := m.data[string(key)]
	if !ok {
		if m.parent == nil {
			return nil
		}
		return m.parent.Get(key)
	}
	return v
}

func (m *MVCCMap) Put(key []byte, value Value) {
	m.data[string(key)] = value
}

func (m *MVCCMap) All(prefix []byte) []Value {
	values := make([]Value, 0)
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

func (m *MVCCMap) Fork() *MVCCMap {
	mvccmap := &MVCCMap{
		data:   make(map[string]Value),
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
