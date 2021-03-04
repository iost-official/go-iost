package contract

import (
	"encoding/base64"
	"errors"

	"encoding/json"
	"io/ioutil"

	"github.com/golang/protobuf/proto"
	"github.com/iost-official/go-iost/v3/common"
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

const codeSizeLimit = 49152

// FixedAmount the limit amount of token used by contract
type FixedAmount struct {
	Token string
	Val   *common.Fixed
}

//type ContractInfo struct {
//	Name     string
//	Lang     string
//	Version  VersionCode
//	Payment  PaymentCode
//	Limit    Cost
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

// VerifySelf verify contract's size
func (c *Contract) VerifySelf() error {
	if len(c.Code) > codeSizeLimit {
		return errors.New("code size invalid")
	}
	if c.Info == nil {
		return errors.New("invalid code info")
	}
	return nil
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

// Compile read src and abi file, generate contract structure
func Compile(id, src, abi string) (*Contract, error) {
	bs, err := ioutil.ReadFile(src)
	if err != nil {
		return nil, err
	}
	code := string(bs)

	as, err := ioutil.ReadFile(abi)
	if err != nil {
		return nil, err
	}

	var info Info
	err = json.Unmarshal(as, &info)
	if err != nil {
		return nil, err
	}
	c := Contract{
		ID:   id,
		Info: &info,
		Code: code,
	}

	return &c, nil
}

// ToBytes converts Amount to bytes.
func (a *Amount) ToBytes() []byte {
	se := common.NewSimpleEncoder()
	se.WriteString(a.Token)
	se.WriteString(a.Val)
	return se.Bytes()
}

// Equal returns whether two amount are equal.
func (a *Amount) Equal(am *Amount) bool {
	return a.Token == am.Token && a.Val == am.Val
}
