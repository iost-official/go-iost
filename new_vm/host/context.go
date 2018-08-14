package host

type Context struct { // thread unsafe
	base   *Context
	value  map[string]interface{}
	gValue map[string]interface{}
}

func NewContext(base *Context) *Context {

	if base != nil {
		return &Context{
			base:  base,
			value: make(map[string]interface{}),
		}
	} else {
		return &Context{
			base:   nil,
			value:  make(map[string]interface{}),
			gValue: make(map[string]interface{}),
		}
	}

}

func (c *Context) Value(key string) (value interface{}) {
	cc := c
	for {
		var ok bool
		value, ok = cc.value[key]
		if ok {
			return
		} else {
			cc = cc.base
			if cc == nil {
				return nil
			}
		}
	}
}

func (c *Context) Set(key string, value interface{}) {
	c.value[key] = value
}

func (c *Context) GValue(key string) (value interface{}) {
	cc := c
	for cc.base != nil {
		cc = cc.base
	}
	return cc.gValue[key]
}

func (c *Context) GSet(key string, value interface{}) {
	cc := c
	for cc.base != nil {
		cc = cc.base
	}
	cc.gValue[key] = value
}

func (c *Context) Base() *Context {
	return c.base
}
