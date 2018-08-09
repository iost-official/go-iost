package contract

//type CostOld struct {
//	Data int64
//	Net  int64
//	CPU  int64
//}

func (c Cost) ToGas() int64 {
	return c.Data + c.Net + c.CPU
}

func (c *Cost) AddAssign(a *Cost) {
	if a == nil {
		return
	}
	c.Data += a.Data
	c.Net += a.Net
	c.CPU += a.CPU
}

func Cost0() *Cost {
	return &Cost{}
}

func NewCost(data, net, cpu int64) *Cost {
	return &Cost{data, net, cpu}
}
