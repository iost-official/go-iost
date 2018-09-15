package trie

import (
	"sync"
)

// Constant of trie
const (
	FreeListSize = uint64(1048576)
)

// Node is node of trie
type Node struct {
	context  *Context
	value    interface{}
	children map[byte]*Node
}

func (n *Node) get(key []byte, i int) *Node {
	if i >= len(key) {
		return n
	}
	child := n.children[key[i]]
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
	child := n.children[key[i]]
	if child == nil {
		child = n.context.newNode()
		n.children[key[i]] = child
	}
	if child.context == nil {
		child = n.context.newNode()
		n.children[key[i]] = child
	}
	if child.context != n.context {
		child = child.forkWithContext(n.context)
		n.children[key[i]] = child
	}
	return child.put(key, value, i+1)
}

func (n *Node) forkWithContext(context *Context) *Node {
	node := context.newNode()
	node.value = n.value
	for k, v := range n.children {
		if v.context == nil {
			continue
		}
		node.children[k] = v
	}
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
			children: make(map[byte]*Node),
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

type Context struct {
	freelist *FreeList
}

func NewContext() *Context {
	c := &Context{
		freelist: NewFreeList(),
	}
	return c
}

func (c *Context) newNode() *Node {
	node := c.freelist.newNode()
	node.context = c
	return node
}

func (c *Context) freeNode(n *Node) {
	n.context = nil
	n.value = nil
	n.children = make(map[byte]*Node)
	c.freelist.freeNode(n)
}

func (c *Context) fork() *Context {
	context := &Context{
		freelist: c.freelist,
	}
	return context
}

type Trie struct {
	context *Context
	root    *Node
	rwmu    *sync.RWMutex
}

func New() *Trie {
	t := &Trie{
		context: NewContext(),
		rwmu:    new(sync.RWMutex),
	}
	t.root = t.context.newNode()
	return t
}

func (t *Trie) Get(key []byte) interface{} {
	t.rwmu.RLock()
	defer t.rwmu.RUnlock()

	node := t.root.get(key, 0)
	if node == nil {
		return nil
	}
	return node.value
}

func (t *Trie) Put(key []byte, value interface{}) {
	t.rwmu.RLock()
	defer t.rwmu.RUnlock()

	if t.root.context != t.context {
		root := t.root.forkWithContext(t.context)
		t.root = root
	}
	t.root.put(key, value, 0)
}

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

func (t *Trie) Free() {
	t.rwmu.Lock()
	defer t.rwmu.Unlock()

	if t.context == t.root.context {
		t.root.free()
	}
	t.context = nil
}
