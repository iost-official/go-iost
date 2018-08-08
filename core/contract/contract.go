package contract

import "github.com/gogo/protobuf/proto"

//go:generate protoc --gofast_out=. *.proto

type VersionCode string

type PaymentCode int32

const (
	SelfPay PaymentCode = iota
	ContractPay
)

//type ContractInfo struct {
//	Name     string
//	Lang     string
//	Version  VersionCode
//	Payment  PaymentCode
//	Limit    *Cost
//	GasPrice uint64
//}
//
//type ContractOld struct {
//	ContractInfo
//	Code string
//}

func (c *Contract) Encode() string {
	buf, err := c.Marshal()
	if err != nil {
		panic(err)
	}
	return string(buf)
}

func (c *Contract) Decode(str string) {
	err := proto.Unmarshal([]byte(str), c)
	if err != nil {
		panic(err)
	}
}

func DecodeContract(str string) *Contract {
	var c Contract
	err := proto.Unmarshal([]byte(str), &c)
	if err != nil {
		panic(err)
	}
	return nil
}

func (c *Contract) ABI(name string) *ABI {
	for _, a := range c.Info.Abis {
		if a.Name == name {
			return a
		}
	}
	return nil
}
