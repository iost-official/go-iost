package contract

import (
	"encoding/base64"

	"github.com/gogo/protobuf/proto"
)

//go:generate protoc --gofast_out=. contract.proto

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
	buf, err := proto.Marshal(c)
	if err != nil {
		panic(err)
	}
	return string(buf)
}

func (c *Contract) Decode(str string) error {
	err := proto.Unmarshal([]byte(str), c)
	if err != nil {
		return err
	}
	return nil
}

func (c *Contract) B64Encode() string {
	buf, err := proto.Marshal(c)
	if err != nil {
		panic(err)
	}
	return base64.StdEncoding.EncodeToString(buf)
}

func (c *Contract) B64Decode(str string) error {
	buf, err := base64.StdEncoding.DecodeString(str)
	if err != nil {
		return err
	}
	return proto.Unmarshal(buf, c)
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
