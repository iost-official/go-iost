package trie

import (
	"sync"
)

// Constant of trie
const (
	FreeListSize = uint64(0)
)

// Node is node of trie
type Node struct {
	context  *Context
	value    interface{}
	keys     []byte
	children []*Node
}

func (n *Node) get(key []byte, i int) *Node {
	if i >= len(key) {
		return n
	}

	var child *Node
	for k := range n.keys {
		if n.keys[k] == key[i] {
			child = n.children[k]
			break
		}
	}

	if child == nil {
		return nil
	}
	if child.context == nil {
		return nil
	}
	return child.get(key, i+1)
}

func (n *Node) all() []*Node {
	nodelist := []*Node{n}
	for _, c := range n.children {
		if c.context == nil {
			continue
		}
		nodelist = append(nodelist, c.all()...)
	}
	return nodelist
}

func (n *Node) put(key []byte, value interface{}, i int) *Node {
	if i >= len(key) {
		n.value = value
		return n
	}

	var k int
	var child *Node
	for k = range n.keys {
		if n.keys[k] == key[i] {
			child = n.children[k]
			break
		}
	}

	if child == nil {
		child = n.context.newNode()
		n.keys = append(n.keys, key[i])
		n.children = append(n.children, child)
	}
	if child.context == nil {
		child = n.context.newNode()
		n.children[k] = child
	}
	if child.context != n.context {
		child = child.forkWithContext(n.context)
		n.children[k] = child
	}
	return child.put(key, value, i+1)
}

func (n *Node) forkWithContext(context *Context) *Node {
	node := context.newNode()
	node.value = n.value

	node.keys = make([]byte, len(n.keys), cap(n.keys))
	copy(node.keys, n.keys)
	node.children = make([]*Node, len(n.children), cap(n.children))
	copy(node.children, n.children)

	return node
}

func (n *Node) free() {
	for _, c := range n.children {
		if c.context == n.context {
			c.free()
		}
	}
	n.context.freeNode(n)
}

// FreeList is a free list of trie node
// All trie that fork from the same trie share the same free list
type FreeList struct {
	mu       *sync.Mutex
	freelist []*Node
}

// NewFreeList create a default FreeList
func NewFreeList() *FreeList {
	f := &FreeList{
		mu:       new(sync.Mutex),
		freelist: make([]*Node, 0, FreeListSize),
	}
	return f
}

func (f *FreeList) newNode() *Node {
	f.mu.Lock()
	defer f.mu.Unlock()

	i := len(f.freelist) - 1
	if i < 0 {
		return &Node{
			context:  nil,
			value:    nil,
			keys:     make([]byte, 0),
			children: make([]*Node, 0),
		}
	}
	node := f.freelist[i]
	f.freelist[i] = nil
	f.freelist = f.freelist[:i]

	return node
}

func (f *FreeList) freeNode(n *Node) {
	f.mu.Lock()
	defer f.mu.Unlock()

	if len(f.freelist) < cap(f.freelist) {
		f.freelist = append(f.freelist, n)
	}
}

// Context is the write context of trie
type Context struct {
	freelist *FreeList
}

// NewContext returns new context
func NewContext() *Context {
	c := &Context{
		freelist: NewFreeList(),
	}
	return c
}

func (c *Context) newNode() *Node {
	// node := c.freelist.newNode()
	node := &Node{
		context:  nil,
		value:    nil,
		keys:     make([]byte, 0),
		children: make([]*Node, 0),
	}

	node.context = c
	return node
}

func (c *Context) freeNode(n *Node) {
	n.context = nil
	n.value = nil
	n.keys = nil
	n.children = nil
	// c.freelist.freeNode(n)
}

func (c *Context) fork() *Context {
	context := &Context{
		freelist: c.freelist,
	}
	return context
}

// Trie is the mvcc trie
type Trie struct {
	context *Context
	root    *Node
	rwmu    *sync.RWMutex
}

// New returns new trie
func New() *Trie {
	t := &Trie{
		context: NewContext(),
		rwmu:    new(sync.RWMutex),
	}
	t.root = t.context.newNode()
	return t
}

// Get returns the value of specify key
func (t *Trie) Get(key []byte) interface{} {
	t.rwmu.RLock()
	defer t.rwmu.RUnlock()

	node := t.root.get(key, 0)
	if node == nil {
		return nil
	}
	return node.value
}

// Put will insert the key-value pair
func (t *Trie) Put(key []byte, value interface{}) {
	t.rwmu.RLock()
	defer t.rwmu.RUnlock()

	if t.root.context != t.context {
		root := t.root.forkWithContext(t.context)
		t.root = root
	}
	t.root.put(key, value, 0)
}

// All returns the list of node prefixed with prefix
func (t *Trie) All(prefix []byte) []interface{} {
	t.rwmu.RLock()
	defer t.rwmu.RUnlock()

	node := t.root.get(prefix, 0)
	valuelist := []interface{}{}
	for _, n := range node.all() {
		if n.value != nil {
			valuelist = append(valuelist, n.value)
		}
	}
	return valuelist
}

// Fork will fork the trie
// thread safe between all forks of the trie
func (t *Trie) Fork() interface{} {
	t.rwmu.RLock()
	defer t.rwmu.RUnlock()

	trie := &Trie{
		context: t.context.fork(),
		root:    t.root,
		rwmu:    t.rwmu,
	}
	return trie
}

// Free will free the memory of trie
func (t *Trie) Free() {
	t.rwmu.Lock()
	defer t.rwmu.Unlock()

	if t.context == t.root.context {
		t.root.free()
	}
	t.context = nil
}
