package vm

type ContractInfo struct {
	Prefix   string
	Language string
	Version  int8

	GasLimit int64
	Price    float64

	Signers [][]byte
	Sender  []byte
}

func (c *ContractInfo) toRaw() ContractInfoRaw {
	return ContractInfoRaw{
		Language: c.Language,
		Version:  c.Version,
		GasLimit: c.GasLimit,
		Price:    c.Price,
		Signers:  c.Signers,
		Sender:   c.Sender,
	}
}

func (d *ContractInfoRaw) toC() ContractInfo {
	return ContractInfo{
		Language: d.Language,
		Version:  d.Version,
		GasLimit: d.GasLimit,
		Price:    d.Price,
		Signers:  d.Signers,
		Sender:   d.Sender,
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
	cir := ContractInfoRaw{}
	_, err:=cir.Unmarshal(b)
	if err != nil {
		return err
	}
	cc := cir.toC()
	c = &cc
	return nil
}