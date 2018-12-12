package tx

import (
	"github.com/golang/protobuf/proto"
	"github.com/iost-official/go-iost/common"
	txpb "github.com/iost-official/go-iost/core/tx/pb"
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

// Status status of transaction execution result, including code and message
type Status struct {
	Code    StatusCode
	Message string
}

// ToPb convert Status to proto buf data structure.
func (s *Status) ToPb() *txpb.Status {
	if s == nil {
		return nil
	}
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

// ToBytes converts Return to a specific byte slice.
func (s *Status) ToBytes() []byte {
	sn := common.NewSimpleNotation()
	sn.WriteInt32((int32(s.Code)), true)
	sn.WriteString(s.Message, true)
	return sn.Bytes()
}

// ReceiptType type of single receipt
type ReceiptType int32

const (
	// SystemDefined system receipt, recording info of calling a method
	SystemDefined ReceiptType = iota
	// UserDefined user defined receipt, usually a json string
	UserDefined
)

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

// ToBytes converts Receipt to a specific byte slice.
func (r *Receipt) ToBytes() []byte {
	sn := common.NewSimpleNotation()
	sn.WriteString(r.FuncName, true)
	sn.WriteString(r.Content, true)
	return sn.Bytes()
}

// TxReceipt Transaction Receipt
type TxReceipt struct { //nolint:golint
	TxHash   []byte
	GasUsage int64
	RAMUsage map[string]int64
	Status   *Status
	Returns  []string
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
		Returns:  []string{},
		Receipts: []*Receipt{},
	}
}

// ToPb convert TxReceipt to proto buf data structure.
func (r *TxReceipt) ToPb() *txpb.TxReceipt {
	tr := &txpb.TxReceipt{
		TxHash:   r.TxHash,
		GasUsage: r.GasUsage,
		Status:   r.Status.ToPb(),
		Returns:  []string{},
		Receipts: []*txpb.Receipt{},
	}

	tr.RamUsage = r.RAMUsage

	for _, rt := range r.Returns {
		tr.Returns = append(tr.Returns, rt)
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
	r.RAMUsage = tr.RamUsage
	s := &Status{}
	r.Status = s.FromPb(tr.Status)
	for _, rt := range tr.Returns {
		r.Returns = append(r.Returns, rt)
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
	err := proto.Unmarshal(b, tr)
	if err != nil {
		return err
	}
	r.FromPb(tr)
	return nil
}

// ToBytes converts TxReceipt to a specific byte slice.
func (r *TxReceipt) ToBytes() []byte {
	sn := common.NewSimpleNotation()
	sn.WriteBytes(r.TxHash, false)
	sn.WriteInt64(r.GasUsage, true)
	sn.WriteBytes(r.Status.ToBytes(), false)
	sn.WriteMapStringToI64(r.RAMUsage, true)

	returnBytes := make([][]byte, 0, len(r.Returns))
	for _, rt := range r.Returns {
		returnBytes = append(returnBytes, []byte(rt))
	}
	sn.WriteBytesSlice(returnBytes, false)

	receiptBytes := make([][]byte, 0, len(r.Receipts))
	for _, re := range r.Receipts {
		receiptBytes = append(receiptBytes, re.ToBytes())
	}
	sn.WriteBytesSlice(receiptBytes, false)

	return sn.Bytes()
}

// Hash return byte hash
func (r *TxReceipt) Hash() []byte {
	return common.Sha3(r.ToBytes())
}

func (r *TxReceipt) String() string {
	if r == nil {
		return "<nil TxReceipt>"
	}
	tr := r.ToPb()
	return tr.String()
}
