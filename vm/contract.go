package vm

import (
	"github.com/iost-official/prototype/core"
	"github.com/iost-official/prototype/common"
	"fmt"
)

type ContractInfo struct {
	Name     string
	Language string
	Version  string

	GasLimit uint64
	Price    float64

	Signers []Pubkey

	ApiList []string
}

type Contract interface {
	core.Serializable

	Info() ContractInfo
	Api(apiName string) (Method, error)
}

type ContractImpl struct {
	info  ContractInfo
	apis  map[string]Method
	signs []common.Signature
}

func (c *ContractImpl) Info() ContractInfo {
	return c.info
}
func (c *ContractImpl) Api(apiName string) (Method, error) {
	m, ok := c.apis[apiName]
	if !ok {
		return Method{}, fmt.Errorf("not found")
	}
	return m, nil
}
func (c *ContractImpl) Encode() []byte {
	return nil
}
func (c *ContractImpl) Decode([]byte) error {
	return nil
}
func (c *ContractImpl) Hash() []byte {
	return nil
}
