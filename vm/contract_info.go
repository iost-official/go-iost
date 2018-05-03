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
		Language:c.Language,
		Version:c.Version,
		GasLimit:c.GasLimit,
		Price:c.Price,
		Signers:c.Signers,
		Sender:c.Sender,
	}
}

func (c *ContractInfoRaw) toC() ContractInfo {
	return ContractInfo{
		Language:c.Language,
		Version:c.Version,
		GasLimit:c.GasLimit,
		Price:c.Price,
		Signers:c.Signers,
		Sender:c.Sender,
	}
}
