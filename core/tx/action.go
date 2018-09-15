package tx

import (
	"github.com/golang/protobuf/proto"
)

// Action 的实现
type Action struct {
	Contract   string // 合约地址，为空则视为调用系统合约
	ActionName string // 方法名称
	Data       string // json
}

// NewAction constructor of Action
func NewAction(contract string, name string, data string) Action {
	return Action{
		Contract:   contract,
		ActionName: name,
		Data:       data,
	}
}

// Encode encode action as byte array
func (a *Action) Encode() []byte {
	ar := &ActionRaw{
		Contract:   a.Contract,
		ActionName: a.ActionName,
		Data:       a.Data,
	}
	b, err := proto.Marshal(ar)
	if err != nil {
		panic(err)
	}
	return b
}

// Decode action from byte array
func (a *Action) Decode(b []byte) error {
	ar := &ActionRaw{}
	err := proto.Unmarshal(b, ar)
	if err != nil {
		return err
	}
	a.Contract = ar.Contract
	a.ActionName = ar.ActionName
	a.Data = ar.Data
	return nil
}

// String return human readable string
func (a *Action) String() string {
	str := "Action{"
	str += "Contract: " + a.Contract + ", "
	str += "ActionName: " + a.ActionName + ", "
	str += "Data: " + a.Data
	str += "}\n"
	return str
}
