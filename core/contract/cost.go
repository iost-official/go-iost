package contract

type Cost struct {
	Data uint64
	Net  uint64
	CPU  uint64
}

func (c Cost) ToGas() uint64 {
	return c.Data + c.Net + c.CPU
}

func (c *Cost) AddAssign(a *Cost) {
	c.Data += a.Data
	c.Net += a.Net
	c.CPU += a.CPU
}
