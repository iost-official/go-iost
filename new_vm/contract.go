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

func (c *Contract) Encode() string {
	return ""
}

func (c *Contract) Decode(string) {

}

func DecodeContract(str string) *Contract {
	return nil
}
