package database

type database interface {
	Get(key string) (value string)
	Put(key, value string)
	Has(key string) bool
	Del(key string)
}

const (
	// StateTable name
	StateTable = "state"
)

type chainbaseAdapter struct {
	cb  IMultiValue
	err error // todo handle error
}

func (c *chainbaseAdapter) Get(key string) (value string) {
	var err error
	value, err = c.cb.Get(StateTable, key)
	if err != nil {
		c.err = err
		panic(c.err)
		return NilPrefix
	}
	if value == "" {
		return NilPrefix
	}
	return
}

func (c *chainbaseAdapter) Put(key, value string) {
	c.err = c.cb.Put(StateTable, key, value)
	if c.err != nil {
		panic(c.err)
	}
}

func (c *chainbaseAdapter) Has(key string) bool {
	ok, err := c.cb.Has(StateTable, key)
	if err != nil {
		c.err = err
		panic(c.err)
		return false
	}
	return ok
}

func (c *chainbaseAdapter) Del(key string) {
	c.err = c.cb.Del(StateTable, key)
	if c.err != nil {
		panic(c.err)
	}
}

func newChainbaseAdapter(cb IMultiValue) *chainbaseAdapter {
	return &chainbaseAdapter{cb, nil}
}
