package trie

import "sync"

const (
	FreeListSize = uint64(65536)
)

type Value interface {
}

type Node struct {
	context  *Context
	value    Value
	children map[byte]*Node
	edge     byte
	refs     []*Node
}

func (n *Node) get(key []byte, i int) *Node {
	if i >= len(key) {
		return n
	}
	child := n.children[key[i]]
	if child == nil {
		return nil
	}
	return child.get(key, i+1)
}

func (n *Node) all() []*Node {
	nodelist := []*Node{n}
	for _, c := range n.children {
		nodelist = append(nodelist, c.all()...)
	}
	return nodelist
}

func (n *Node) put(key []byte, value Value, i int) *Node {
	if i >= len(key) {
		n.value = value
		return n
	}
	child := n.children[key[i]]
	if child == nil {
		child = n.context.newNode()
		n.children[key[i]] = child
		child.edge = key[i]
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
		node.children[k] = v
		v.refs = append(v.refs, node)
	}
	return node
}

func (n *Node) free() {
	for _, c := range n.children {
		c.free()
	}
	for _, p := range n.refs {
		delete(p.children, n.edge)
	}
	n.value = nil
	n.children = nil
	n.edge = byte(0)
	n.refs = nil
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
	freelist *FreeList
}

func NewContext() *Context {
	c := &Context{
		freelist: NewFreeList(),
	}
	return c
}

func (c *Context) newNode() *Node {
	node := &Node{
		context:  NewContext(),
		children: make(map[byte]*Node),
		refs:     make([]*Node, 0),
	}
	return node
}

func (c *Context) freeNode(n *Node) {
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
	size    uint64
	tag     string
}

func New() *Trie {
	t := &Trie{
		context: NewContext(),
		size:    0,
		tag:     "",
	}
	t.root = t.context.newNode()
	return t
}

func (t *Trie) Get(key []byte) Value {
	node := t.root.get(key, 0)
	if node == nil {
		return nil
	}
	return node.value
}

func (t *Trie) Put(key []byte, value Value) {
	if t.root.context != t.context {
		root := t.root.forkWithContext(t.context)
		t.root = root
	}
	t.root.put(key, value, 0)
}

func (t *Trie) All(prefix []byte) []Value {
	node := t.root.get(prefix, 0)
	valuelist := []Value{}
	for _, n := range node.all() {
		if n.value != nil {
			valuelist = append(valuelist, n.value)
		}
	}
	return valuelist
}

func (t *Trie) Fork() *Trie {
	trie := &Trie{
		context: t.context.fork(),
		root:    t.root,
		size:    t.size,
		tag:     "",
	}
	return trie
}

func (t *Trie) Free() {
	t.root.free()
	t.context = nil
	t.size = 0
	t.tag = ""
}

func (t *Trie) Tag() string {
	return t.tag
}

func (t *Trie) SetTag(tag string) {
	t.tag = tag
}
