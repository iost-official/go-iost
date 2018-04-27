package vm

import (
	"fmt"
	"github.com/iost-official/prototype/common"
)

//go:generate gencode go -schema=structs.schema -package=vm

type Contract interface {
	Info() ContractInfo
	Api(apiName string) (Method, error)
}

type LuaContract struct {
	info ContractInfo
	code string
	main Method
}

func (c *LuaContract) Info() ContractInfo {
	return c.info
}
func (c *LuaContract) Api(apiName string) (Method, error) {
	if apiName == "main" {
		return c.main, nil
	}
	return nil, fmt.Errorf("not found")
}
func (c *LuaContract) Encode() []byte {
	cr := ContractRaw{
		info: c.info,
		code: []byte(c.code),
	}
	b, err := cr.Marshal(nil)
	if err != nil {
		panic(err)
		return nil
	}
	return b
}
func (c *LuaContract) Decode(b []byte) error {
	var cr ContractRaw
	_, err := cr.Unmarshal(b)
	c.info = cr.info
	c.code = string(cr.code)
	return err
}
func (c *LuaContract) Hash() []byte {
	return common.Sha256(c.Encode())
}
