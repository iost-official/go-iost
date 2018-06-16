package vm

// ContractInfo 是智能合约的相关信息
// 编译之后Language，version, gas limit， price有值，
// 打包tx后signer，publisher有值
// 被block chain收录之后prefix有值（=txhash的base58编码）
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
