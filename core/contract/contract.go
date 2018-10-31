package contract

import (
	"encoding/base64"

	"github.com/gogo/protobuf/proto"
	"github.com/iost-official/go-iost/common"
)

//go:generate protoc --gofast_out=. contract.proto

// VersionCode version of contract
type VersionCode string

// PaymentCode payment mode of contract
type PaymentCode int32

// Payment mode
const (
	SelfPay PaymentCode = iota
	ContractPay
)

type FixedAmount struct {
	Token string
	Val  *common.Fixed
}

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

// Encode contract to string using proto buf
func (c *Contract) Encode() string {
	buf, err := proto.Marshal(c)
	if err != nil {
		panic(err)
	}
	return string(buf)
}

// Decode contract from string using proto buf
func (c *Contract) Decode(str string) error {
	err := proto.Unmarshal([]byte(str), c)
	if err != nil {
		return err
	}
	return nil
}

// B64Encode encode contract to base64 string
func (c *Contract) B64Encode() string {
	buf, err := proto.Marshal(c)
	if err != nil {
		panic(err)
	}
	return base64.StdEncoding.EncodeToString(buf)
}

// B64Decode decode contract from base64 string
func (c *Contract) B64Decode(str string) error {
	buf, err := base64.StdEncoding.DecodeString(str)
	if err != nil {
		return err
	}
	return proto.Unmarshal(buf, c)
}

// DecodeContract static method to decode contract from string
func DecodeContract(str string) *Contract {
	var c Contract
	err := proto.Unmarshal([]byte(str), &c)
	if err != nil {
		panic(err)
	}
	return nil
}

// ABI get abi from contract with specific name
func (c *Contract) ABI(name string) *ABI {
	for _, a := range c.Info.Abi {
		if a.Name == name {
			return a
		}
	}
	return nil
}
