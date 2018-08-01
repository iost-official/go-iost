package new_vm

type VersionCode string

type ContractInfo struct {
	Name    string
	Lang    string
	Version VersionCode
}

type Contract struct {
	ContractInfo
	Code string
}

func (c *Contract) Encode() string { // todo
	return ""
}

func (c *Contract) Decode(string) { // todo

}

func DecodeContract(str string) *Contract { // todo
	return nil
}
