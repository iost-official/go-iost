package vm

type ContractInfo struct {
	Prefix   string
	Language string
	Version  int8

	GasLimit int64
	Price    float64

	Signers []IOSTAccount
	Sender  IOSTAccount
}

func (c *ContractInfo) toRaw() contractInfoRaw {
	return contractInfoRaw{
		Language: c.Language,
		Version:  c.Version,
		GasLimit: c.GasLimit,
		Price:    c.Price,
	}
}

func (d *contractInfoRaw) toC() ContractInfo {
	return ContractInfo{
		Language: d.Language,
		Version:  d.Version,
		GasLimit: d.GasLimit,
		Price:    d.Price,
	}
}

func (c *ContractInfo) Encode() []byte {
	cir := c.toRaw()
	buf, err := cir.Marshal(nil)
	if err != nil {
		panic(err)
	}
	return buf
}

func (c *ContractInfo) Decode(b []byte) error {
	cir := contractInfoRaw{}
	_, err := cir.Unmarshal(b)
	if err != nil {
		return err
	}
	cc := cir.toC()
	c = &cc
	return nil
}
