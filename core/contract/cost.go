package contract

// Cost ...
type Cost struct {
	Data int64
	Net  int64
	CPU  int64
}

// ToGas convert cost to gas
func (c Cost) ToGas() int64 {
	return 10*c.Net + c.CPU
}

// AddAssign add cost to self
func (c *Cost) AddAssign(a Cost) {
	c.Data += a.Data
	c.Net += a.Net
	c.CPU += a.CPU
}

// Multiply a cost to int64, return new cost
func (c Cost) Multiply(i int64) Cost {
	var d Cost
	d.Data = i * c.Data
	d.Net = i * c.Net
	d.CPU = i * c.CPU
	return d
}

// IsOverflow decide if exceed limit
func (c Cost) IsOverflow(limit Cost) bool {
	if c.Data > limit.Data ||
		c.Net > limit.Net ||
		c.CPU > limit.CPU {
		return true
	}

	return false
}

// Cost0 construct zero cost
func Cost0() Cost {
	return Cost{}
}

// NewCost construct cost with specific data, net, cpu
func NewCost(data, net, cpu int64) Cost {
	return Cost{
		Data: data,
		Net:  net,
		CPU:  cpu,
	}
}
