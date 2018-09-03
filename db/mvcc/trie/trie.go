package trie

import (
	"math"
	"sync"
	"time"
)

const (
	FreeListSize = uint64(65536)
)

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
	if child.context.timestamp > n.context.timestamp {
		return nil
	}
	return child.get(key, i+1)
}

func (n *Node) all() []*Node {
	nodelist := []*Node{n}
	for _, c := range n.children {
		if c.context.timestamp > n.context.timestamp {
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
	if child.context.timestamp > n.context.timestamp {
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
	node := &Node{
		context:  context,
		value:    n.value,
		children: make(map[byte]*Node, len(n.children)),
	}
	for k, v := range n.children {
		if v.context.timestamp > n.context.timestamp {
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
	n.value = nil
	n.children = nil
	n.context.freeNode(n)
}

type FreeList struct {
	mu       *sync.Mutex
	freelist []*Node
}

func NewFreeList() *FreeList {
	f := &FreeList{
		mu:       new(sync.Mutex),
		freelist: make([]*Node, 0, FreeListSize),
	}
	return f
}

type Context struct {
	freelist  *FreeList
	timestamp int64
}

func NewContext() *Context {
	c := &Context{
		freelist:  NewFreeList(),
		timestamp: time.Now().UnixNano(),
	}
	return c
}

func (c *Context) newNode() *Node {
	node := &Node{
		context:  NewContext(),
		children: make(map[byte]*Node),
	}
	return node
}

func (c *Context) freeNode(n *Node) {
}

func (c *Context) fork() *Context {
	context := &Context{
		freelist:  c.freelist,
		timestamp: time.Now().UnixNano(),
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

	t.root.free()
	t.context.timestamp = math.MaxInt64
}
