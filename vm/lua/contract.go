package lua

import (
	"fmt"
	"github.com/iost-official/prototype/common"
	"github.com/iost-official/prototype/vm"
)

type Contract struct {
	info vm.ContractInfo
	code string
	main Method
	apis map[string]Method
}

func (c *Contract) Info() vm.ContractInfo {
	return c.info
}
func (c *Contract) SetPrefix(prefix string) {
	c.info.Prefix = prefix
}
func (c *Contract) SetSender(sender []byte) {
	c.info.Sender = sender
}
func (c *Contract) AddSigner(signer []byte) {
	c.info.Signers = append(c.info.Signers, signer)
}
func (c *Contract) Api(apiName string) (vm.Method, error) {
	if apiName == "main" {
		return &c.main, nil
	}
	rtn, ok := c.apis[apiName]
	if !ok {
		return nil, fmt.Errorf("not found")
	}
	return &rtn, nil
}
func (c *Contract) Encode() []byte {
	cr := contractRaw{
		info: c.info.Encode(),
		code: []byte(c.code),
	}
	b, err := cr.Marshal(nil)
	if err != nil {
		panic(err)
		return nil
	}
	return b
}
func (c *Contract) Decode(b []byte) error {
	var cr contractRaw
	_, err := cr.Unmarshal(b)
	var ci vm.ContractInfo
	err = ci.Decode(cr.info)
	if err != nil {
		return err
	}
	c.info = ci
	c.code = string(cr.code)
	return err
}
func (c *Contract) Hash() []byte {
	return common.Sha256(c.Encode())
}
