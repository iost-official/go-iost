package host

import (
	"bytes"
	"fmt"
)

// Context thread unsafe context with global fields
type Context struct { // thread unsafe
	base   *Context
	value  map[string]any
	gValue map[string]any
}

// NewContext new context based on base
func NewContext(base *Context) *Context {
	if base != nil {
		return &Context{
			base:  base,
			value: make(map[string]any),
		}
	}

	return &Context{
		base:   nil,
		value:  make(map[string]any),
		gValue: make(map[string]any),
	}
}

func (c *Context) String() string {
	if c == nil {
		return "nil"
	}
	b := new(bytes.Buffer)
	for key, value := range c.value {
		fmt.Fprintf(b, "%s=%v ", key, value)
	}
	fmt.Fprint(b, "\n")
	for key, value := range c.gValue {
		fmt.Fprintf(b, "%s=%v ", key, value)
	}
	fmt.Fprint(b, "\n|\n")
	return b.String() + c.base.String()
}

// Value get value of key
func (c *Context) Value(key string) (value any) {
	cc := c
	for {
		var ok bool
		value, ok = cc.value[key]
		if ok {
			break
		}
		cc = cc.base
		if cc == nil {
			value = nil
			break
		}
	}
	// ilog.Debugf("get %s -> %v", key, value)
	return
}

// Set  set value of k
func (c *Context) Set(key string, value any) {
	// ilog.Debugf("set %s -> %v", key, value)
	c.value[key] = value
}

// GValue get global value of key
func (c *Context) GValue(key string) (value any) {
	cc := c
	for cc.base != nil {
		cc = cc.base
	}
	return cc.gValue[key]
}

// GSet set global value of key, thread unsafe
func (c *Context) GSet(key string, value any) {
	cc := c
	for cc.base != nil {
		cc = cc.base
	}
	cc.gValue[key] = value
}

// GClear clear global values
func (c *Context) GClear() {
	cc := c
	for cc.base != nil {
		cc = cc.base
	}
	cc.gValue = make(map[string]any)
}

// Base get base of context
func (c *Context) Base() *Context {
	return c.base
}
