package vm

type ContractInfo struct {
	Prefix   string
	Language string
	Version  int8

	GasLimit int64
	Price    float64

	Signers   []IOSTAccount
	Publisher IOSTAccount
}

func (c *ContractInfo) toRaw() contractInfoRaw {
	return contractInfoRaw{
		Language: c.Language,
		Version:  c.Version,
		GasLimit: c.GasLimit,
		Price:    c.Price,
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
	c.Language = cir.Language
	c.Version = cir.Version
	c.GasLimit = cir.GasLimit
	c.Price = cir.Price
	return nil
}
