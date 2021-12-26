package mvccmap

import (
	"strings"
	"sync"
)

// MVCCMap is the mvcc map
type MVCCMap struct {
	data   map[string]any
	parent *MVCCMap
	refs   []*MVCCMap
	rwmu   *sync.RWMutex
}

// New returns new map
func New() *MVCCMap {
	return &MVCCMap{
		data:   make(map[string]any),
		parent: nil,
		refs:   make([]*MVCCMap, 0),
		rwmu:   new(sync.RWMutex),
	}
}

func (m *MVCCMap) getFromLink(key string) any {
	v, ok := m.data[key]
	if !ok {
		if m.parent == nil {
			return nil
		}
		return m.parent.getFromLink(key)
	}
	return v
}

// Get returns the value of specify key
func (m *MVCCMap) Get(key []byte) any {
	m.rwmu.RLock()
	defer m.rwmu.RUnlock()

	return m.getFromLink(string(key))
}

// Put will insert the key-value pair
func (m *MVCCMap) Put(key []byte, value any) {
	m.rwmu.Lock()
	defer m.rwmu.Unlock()

	m.data[string(key)] = value
}

func (m *MVCCMap) allFromLink(prefix []byte) []any {
	values := make([]any, 0)
	for k, v := range m.data {
		if strings.HasPrefix(k, string(prefix)) {
			values = append(values, v)
		}
	}
	if m.parent == nil {
		return values
	}
	return append(m.parent.allFromLink(prefix), values...)
}

// All returns the list of nodes prefixed with prefix
func (m *MVCCMap) All(prefix []byte) []any {
	m.rwmu.RLock()
	defer m.rwmu.RUnlock()

	return m.allFromLink(prefix)
}

// Fork will fork the map
// thread safe between all forks of the map
func (m *MVCCMap) Fork() any {
	m.rwmu.Lock()
	defer m.rwmu.Unlock()

	mvccmap := &MVCCMap{
		data:   make(map[string]any),
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

// Free will free the memory of trie
func (m *MVCCMap) Free() {
	m.rwmu.Lock()
	defer m.rwmu.Unlock()

	m.freeFromLink()
}
