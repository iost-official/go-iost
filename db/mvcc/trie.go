package mvcc

type Value struct {
	Revision string
	Raw      []byte
	Deleted  bool
}

type TrieNode struct {
	revision string
	value    *Value
	children [256]*TrieNode
}

func (t *TrieNode) Revision() string {
	return t.revision
}

func (t *TrieNode) Get(key []byte, i int) (*Value, bool) {
	if i >= len(key) {
		if t.value == nil {
			return nil, false
		}
		return t.value, true
	}
	ascii := int(key[i])
	child := t.children[ascii]
	if child == nil {
		return nil, false
	}
	return child.Get(key, i+1)
}

func (t *TrieNode) Put(key []byte, value *Value, i int) {
	if i >= len(key) {
		t.value = value
		return
	}
	ascii := int(key[i])
	child := t.children[ascii]
	if child == nil {
		child = &TrieNode{revision: value.Revision}
		t.children[ascii] = child
	}
	if child.revision != value.Revision {
		child = child.Fork(value.Revision)
		t.children[ascii] = child
	}
	child.Put(key, value, i+1)
}

func (t *TrieNode) Fork(revision string) *TrieNode {
	trienode := &TrieNode{
		revision: revision,
		value:    t.value,
		children: t.children,
	}
	return trienode
}
