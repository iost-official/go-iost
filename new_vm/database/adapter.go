package database

type Database interface {
	Get(key string) (value string)
	Put(key, value string)
	Has(key string) bool
	Keys(prefix string) []string
	Del(key string)
}

const (
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
		return "n"
	}
	return
}
func (c *chainbaseAdapter) Put(key, value string) {
	c.err = c.cb.Put(StateTable, key, value)
}
func (c *chainbaseAdapter) Has(key string) bool {
	ok, err := c.cb.Has(StateTable, key)
	if err != nil {
		c.err = err
		return false
	}
	return ok
}
func (c *chainbaseAdapter) Keys(prefix string) []string {
	var rtn []string
	rtn, c.err = c.cb.Keys(StateTable, prefix)
	return rtn
}

func (c *chainbaseAdapter) Del(key string) {
	c.err = c.cb.Del(StateTable, key)
}

func newChainbaseAdapter(cb IMultiValue) *chainbaseAdapter {
	return &chainbaseAdapter{cb, nil}
}
