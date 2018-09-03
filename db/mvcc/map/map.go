package mvccmap

import (
	"strings"
	"sync"
)

type MVCCMap struct {
	data   map[string]interface{}
	parent *MVCCMap
	refs   []*MVCCMap
	rwmu   *sync.RWMutex
}

func New() *MVCCMap {
	return &MVCCMap{
		data:   make(map[string]interface{}),
		parent: nil,
		refs:   make([]*MVCCMap, 0),
		rwmu:   new(sync.RWMutex),
	}
}

func (m *MVCCMap) getFromLink(key []byte) interface{} {
	v, ok := m.data[string(key)]
	if !ok {
		if m.parent == nil {
			return nil
		}
		return m.parent.getFromLink(key)
	}
	return v
}

func (m *MVCCMap) Get(key []byte) interface{} {
	m.rwmu.RLock()
	defer m.rwmu.RUnlock()

	return m.getFromLink(key)
}

func (m *MVCCMap) Put(key []byte, value interface{}) {
	m.rwmu.RLock()
	defer m.rwmu.RUnlock()

	m.data[string(key)] = value
}

func (m *MVCCMap) allFromLink(prefix []byte) []interface{} {
	values := make([]interface{}, 0)
	for k, v := range m.data {
		if strings.HasPrefix(string(k), string(prefix)) {
			values = append(values, v)
		}
	}
	if m.parent == nil {
		return values
	}
	return append(m.parent.allFromLink(prefix), values...)
}

func (m *MVCCMap) All(prefix []byte) []interface{} {
	m.rwmu.RLock()
	defer m.rwmu.RUnlock()

	return m.allFromLink(prefix)
}

func (m *MVCCMap) Fork() interface{} {
	m.rwmu.Lock()
	defer m.rwmu.Unlock()

	mvccmap := &MVCCMap{
		data:   make(map[string]interface{}),
		parent: m,
		refs:   make([]*MVCCMap, 0),
		rwmu:   m.rwmu,
	}
	m.refs = append(m.refs, mvccmap)
	return mvccmap
}

func (m *MVCCMap) freeFromLink() {
	if m.parent != nil {
		m.parent.freeFromLink()
	}
	for _, ref := range m.refs {
		ref.parent = nil
	}
	m.parent = nil
	m.refs = nil
	m.data = nil
}

func (m *MVCCMap) Free() {
	m.rwmu.Lock()
	defer m.rwmu.Unlock()

	m.freeFromLink()
}
