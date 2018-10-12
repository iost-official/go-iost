package host

// Context thread unsafe context with global fields
type Context struct { // thread unsafe
	base   *Context
	value  map[string]interface{}
	gValue map[string]interface{}
}

// NewContext new context based on base
func NewContext(base *Context) *Context {

	if base != nil {
		return &Context{
			base:  base,
			value: make(map[string]interface{}),
		}
	}

	return &Context{
		base:   nil,
		value:  make(map[string]interface{}),
		gValue: make(map[string]interface{}),
	}

}

// Value get value of key
func (c *Context) Value(key string) (value interface{}) {
	cc := c
	for {
		var ok bool
		value, ok = cc.value[key]
		if ok {
			return
		}
		cc = cc.base
		if cc == nil {
			return nil
		}

	}
}

// Set  set value of k
func (c *Context) Set(key string, value interface{}) {
	c.value[key] = value
}

// GValue get global value of key
func (c *Context) GValue(key string) (value interface{}) {
	cc := c
	for cc.base != nil {
		cc = cc.base
	}
	return cc.gValue[key]
}

// GSet set global value of key, thread unsafe
func (c *Context) GSet(key string, value interface{}) {
	cc := c
	for cc.base != nil {
		cc = cc.base
	}
	cc.gValue[key] = value
}

func (c *Context) GClear() {
	cc := c
	for cc.base != nil {
		cc = cc.base
	}
	cc.gValue = make(map[string]interface{})
}

// Base get base of context
func (c *Context) Base() *Context {
	return c.base
}
