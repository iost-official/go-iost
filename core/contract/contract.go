package contract

type VersionCode string

type PaymentCode int

const (
	SelfPay PaymentCode = iota
	ContractPay
)

type ContractInfo struct {
	Name     string
	Lang     string
	Version  VersionCode
	Payment  PaymentCode
	Limit    *Cost
	GasPrice uint64
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
