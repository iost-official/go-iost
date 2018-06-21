package lua

import (
	"fmt"

	"sort"

	"github.com/iost-official/prototype/common"
	"github.com/iost-official/prototype/log"
	"github.com/iost-official/prototype/vm"
)

// contract lua智能合约的实现
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
func (c *Contract) SetSender(sender vm.IOSTAccount) {
	c.info.Publisher = sender
}
func (c *Contract) AddSigner(signer vm.IOSTAccount) {
	c.info.Signers = append(c.info.Signers, signer)
}
func (c *Contract) API(apiName string) (vm.Method, error) {
	if apiName == "main" {
		return &c.main, nil
	}
	rtn, ok := c.apis[apiName]
	if !ok {
		return nil, fmt.Errorf("api %v: not found", apiName)
	}
	return &rtn, nil
}
func (c *Contract) Code() string {
	return c.code
}
func (c *Contract) Encode() []byte {
	cr := contractRaw{
		info: c.info.Encode(),
		code: []byte(c.code),
	}
	mr := methodRaw{
		name: c.main.name,
		ic:   int32(c.main.inputCount),
		oc:   int32(c.main.outputCount),
		priv: int32(c.main.privilege),
	}
	cr.methods = []methodRaw{}
	for _, val := range c.apis {
		mr2 := methodRaw{
			name: val.name,
			ic:   int32(val.inputCount),
			oc:   int32(val.outputCount),
			priv: int32(val.privilege),
		}
		cr.methods = append(cr.methods, mr2)
		sort.Sort(methodRawSlice(cr.methods))
	}
	cr.methods = append(cr.methods, mr)
	b, err := cr.Marshal(nil)
	if err != nil {
		log.Log.E("Error in Encode of ", c.info.Prefix, err.Error())
		return nil
	}
	return append([]byte{0}, b...)
}
func (c *Contract) Decode(b []byte) error {
	var cr contractRaw
	_, err := cr.Unmarshal(b[1:])
	var ci vm.ContractInfo
	err = ci.Decode(cr.info)
	if err != nil {
		return err
	}
	c.info = ci
	c.code = string(cr.code)
	if c.apis == nil {
		c.apis = make(map[string]Method)
	}
	for i := 1; i < len(cr.methods); i++ {
		if cr.methods[i].name == "main" {
			c.main = Method{
				cr.methods[i].name,
				int(cr.methods[i].ic),
				int(cr.methods[i].oc),
				vm.Public,
			}
		}

		c.apis[cr.methods[i].name] = Method{
			cr.methods[i].name,
			int(cr.methods[i].ic),
			int(cr.methods[i].oc),
			vm.Privilege(cr.methods[i].priv),
		}
	}

	return err
}
func (c *Contract) Hash() []byte {
	return common.Sha256(c.Encode())
}

func NewContract(info vm.ContractInfo, code string, main Method, apis ...Method) Contract {
	c := Contract{
		info: info,
		code: code,
		main: main,
	}
	c.apis = make(map[string]Method)
	for _, api := range apis {
		c.apis[api.name] = api
	}
	return c
}

type methodRawSlice []methodRaw

func (m methodRawSlice) Len() int {
	return len(m)
}
func (m methodRawSlice) Less(i, j int) bool {
	return m[i].name < m[j].name
}
func (m methodRawSlice) Swap(i, j int) {
	m[i], m[j] = m[j], m[i]
}
