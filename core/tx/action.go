package tx

import (
	"github.com/golang/protobuf/proto"
	"github.com/iost-official/go-iost/common"
	txpb "github.com/iost-official/go-iost/core/tx/pb"
)

// Action implement
type Action struct {
	Contract   string `json:"contract"`   // contract name
	ActionName string `json:"actionName"` // method name of contract
	Data       string `json:"data"`       // parameters of method, with json format
}

// NewAction constructor of Action
func NewAction(contract string, name string, data string) *Action {
	return &Action{
		Contract:   contract,
		ActionName: name,
		Data:       data,
	}
}

// ToPb convert Action to proto buf data structure.
func (a *Action) ToPb() *txpb.Action {
	return &txpb.Action{
		Contract:   a.Contract,
		ActionName: a.ActionName,
		Data:       a.Data,
	}
}

// FromPb convert Action from proto buf data structure.
func (a *Action) FromPb(ac *txpb.Action) *Action {
	a.Contract = ac.Contract
	a.ActionName = ac.ActionName
	a.Data = ac.Data
	return a
}

// Encode encode action as byte array
func (a *Action) Encode() []byte {
	b, err := proto.Marshal(a.ToPb())
	if err != nil {
		panic(err)
	}
	return b
}

// Decode action from byte array
func (a *Action) Decode(b []byte) error {
	ac := &txpb.Action{}
	err := proto.Unmarshal(b, ac)
	if err != nil {
		return err
	}
	a.FromPb(ac)
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

// ToBytes converts Action to a specific byte slice.
func (a *Action) ToBytes() []byte {
	se := common.NewSimpleEncoder()
	se.WriteString(a.Contract)
	se.WriteString(a.ActionName)
	se.WriteString(a.Data)
	return se.Bytes()
}

// Equal returns whether two actions are equal.
func (a *Action) Equal(ac *Action) bool {
	return a.ActionName == ac.ActionName && a.Data == ac.Data && a.Contract == ac.Contract
}
