package vm

import (
	"fmt"
	"github.com/iost-official/prototype/common"
)

//go:generate gencode go -schema=structs.schema -package=vm

type Contract interface {
	Info() ContractInfo
	SetPrefix(prefix string)
	SetSender(sender []byte)
	AddSigner(signer []byte)
	Api(apiName string) (Method, error)
	common.Serializable
}

type LuaContract struct {
	info ContractInfo
	code string
	main Method
}

func (c *LuaContract) Info() ContractInfo {
	return c.info
}
func (c *LuaContract) SetPrefix(prefix string) {
	c.info.Prefix = prefix
}
func (c *LuaContract) SetSender(sender []byte) {
	c.info.Sender = sender
}
func (c *LuaContract) AddSigner(signer []byte) {
	c.info.Signers = append(c.info.Signers, signer)
}
func (c *LuaContract) Api(apiName string) (Method, error) {
	if apiName == "main" {
		return c.main, nil
	}
	return nil, fmt.Errorf("not found")
}
func (c *LuaContract) Encode() []byte {
	cr := ContractRaw{
		info: c.info.toRaw(),
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
	c.info = cr.info.toC()
	c.code = string(cr.code)
	return err
}
func (c *LuaContract) Hash() []byte {
	return common.Sha256(c.Encode())
}
