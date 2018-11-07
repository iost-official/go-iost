package tx

import (
	"fmt"

	"github.com/iost-official/go-iost/common"
	txpb "github.com/iost-official/go-iost/core/tx/pb"

	"github.com/golang/protobuf/proto"
)

// StatusCode status code of transaction execution result
type StatusCode int32

// tx execution result
const (
	Success StatusCode = iota
	ErrorGasRunOut
	ErrorBalanceNotEnough
	ErrorParamter // parameter mismatch when calling function
	ErrorRuntime  // runtime error
	ErrorTimeout
	ErrorTxFormat         // tx format errors
	ErrorDuplicateSetCode // more than one set code action in a tx
	ErrorUnknown          // other errors
)

// Return is the result of txreceipt.
type Return struct {
	FuncName string
	Value    string
}

// ToPb convert Return to proto buf data structure.
func (r *Return) ToPb() *txpb.Return {
	return &txpb.Return{
		FuncName: r.FuncName,
		Value:    r.Value,
	}
}

// FromPb convert Return from proto buf data structure.
func (r *Return) FromPb(s *txpb.Return) *Return {
	r.FuncName = s.FuncName
	r.Value = s.Value
	return r
}

// Status status of transaction execution result, including code and message
type Status struct {
	Code    StatusCode
	Message string
}

// ToPb convert Status to proto buf data structure.
func (s *Status) ToPb() *txpb.Status {
	return &txpb.Status{
		Code:    int32(s.Code),
		Message: s.Message,
	}
}

// FromPb convert Status from proto buf data structure.
func (s *Status) FromPb(st *txpb.Status) *Status {
	s.Code = StatusCode(st.Code)
	s.Message = st.Message
	return s
}

// Receipt generated when applying transaction
type Receipt struct {
	FuncName string
	Content  string // can be a raw string or a json string
}

// ToPb convert Receipt to proto buf data structure.
func (r *Receipt) ToPb() *txpb.Receipt {
	return &txpb.Receipt{
		FuncName: r.FuncName,
		Content:  r.Content,
	}
}

// FromPb convert Receipt from proto buf data structure.
func (r *Receipt) FromPb(rp *txpb.Receipt) *Receipt {
	r.FuncName = rp.FuncName
	r.Content = rp.Content
	return r
}

// TxReceipt Transaction Receipt
type TxReceipt struct { //nolint:golint
	TxHash   []byte
	GasUsage int64
	RAMUsage map[string]int64
	Status   *Status
	Returns  []*Return
	Receipts []*Receipt
}

// NewTxReceipt generate tx receipt for a tx hash
func NewTxReceipt(txHash []byte) *TxReceipt {
	var status = &Status{
		Code:    Success,
		Message: "",
	}
	return &TxReceipt{
		TxHash:   txHash,
		GasUsage: 0,
		RAMUsage: make(map[string]int64),
		Status:   status,
		Returns:  []*Return{},
		Receipts: []*Receipt{},
	}
}

// ToPb convert TxReceipt to proto buf data structure.
func (r *TxReceipt) ToPb() *txpb.TxReceipt {
	tr := &txpb.TxReceipt{
		TxHash:   r.TxHash,
		GasUsage: r.GasUsage,
		RamUsage: r.RAMUsage,
		Status:   r.Status.ToPb(),
		Returns:  []*txpb.Return{},
		Receipts: []*txpb.Receipt{},
	}
	for _, rt := range r.Returns {
		if rt == nil {
			fmt.Println("rt is nil")
			break
		}
		tr.Returns = append(tr.Returns, rt.ToPb())
	}
	for _, re := range r.Receipts {
		tr.Receipts = append(tr.Receipts, re.ToPb())
	}
	return tr
}

// Encode TxReceipt as byte array
func (r *TxReceipt) Encode() []byte {
	b, err := proto.Marshal(r.ToPb())
	if err != nil {
		panic(err)
	}
	return b
}

// FromPb convert TxReceipt from proto buf data structure
func (r *TxReceipt) FromPb(tr *txpb.TxReceipt) *TxReceipt {
	r.TxHash = tr.TxHash
	r.GasUsage = tr.GasUsage
	s := &Status{}
	r.Status = s.FromPb(tr.Status)
	for _, rt := range tr.Returns {
		re := &Return{}
		r.Returns = append(r.Returns, re.FromPb(rt))
	}
	for _, re := range tr.Receipts {
		rc := &Receipt{}
		r.Receipts = append(r.Receipts, rc.FromPb(re))
	}
	return r
}

// Decode TxReceipt from byte array
func (r *TxReceipt) Decode(b []byte) error {
	tr := &txpb.TxReceipt{}
	err := tr.Unmarshal(b)
	if err != nil {
		return err
	}
	r.FromPb(tr)
	return nil
}

// Hash return byte hash
func (r *TxReceipt) Hash() []byte {
	return common.Sha3(r.Encode())
}

func (r *TxReceipt) String() string {
	tr := &txpb.TxReceipt{
		TxHash:   r.TxHash,
		GasUsage: r.GasUsage,
		Status:   r.Status.ToPb(),
	}
	return tr.String()
}
