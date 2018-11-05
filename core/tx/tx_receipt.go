package tx

import (
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

// Status status of transaction execution result, including code and message
type Status struct {
	Code    StatusCode
	Message string
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
	Type    ReceiptType // system defined or user defined receipt type
	Content string      // can be a raw string or a json string
}

// TxReceipt Transaction Receipt
type TxReceipt struct { //nolint:golint
	TxHash   []byte
	GasUsage int64
	/*
		CpuTimeUsage    uint64
		NetUsage    uint64
		RAMUsage    uint64
	*/
	Status        *Status
	SuccActionNum int32
	Receipts      []*Receipt
}

// NewTxReceipt generate tx receipt for a tx hash
func NewTxReceipt(txHash []byte) *TxReceipt {
	var status = &Status{
		Code:    Success,
		Message: "",
	}
	return &TxReceipt{
		TxHash:        txHash,
		GasUsage:      0,
		Status:        status,
		SuccActionNum: 0,
		Receipts:      []*Receipt{},
	}
}

// ToPb convert TxReceipt to proto buf data structure.
func (r *TxReceipt) ToPb() *txpb.TxReceipt {
	tr := &txpb.TxReceipt{
		TxHash:   r.TxHash,
		GasUsage: r.GasUsage,
		Status: &txpb.Status{
			Code:    int32(r.Status.Code),
			Message: r.Status.Message,
		},
		SuccActionNum: r.SuccActionNum,
	}
	for _, re := range r.Receipts {
		tr.Receipts = append(tr.Receipts, &txpb.Receipt{
			Type:    int32(re.Type),
			Content: re.Content,
		})
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
func (r *TxReceipt) FromPb(tr *txpb.TxReceipt) {
	r.TxHash = tr.TxHash
	r.GasUsage = tr.GasUsage
	r.Status = &Status{
		Code:    StatusCode(tr.Status.Code),
		Message: tr.Status.Message,
	}
	r.SuccActionNum = tr.SuccActionNum
	r.Receipts = []*Receipt{}
	for _, re := range tr.Receipts {
		r.Receipts = append(r.Receipts, &Receipt{
			Type:    ReceiptType(re.Type),
			Content: re.Content,
		})
	}
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
		Status: &txpb.Status{
			Code:    int32(r.Status.Code),
			Message: r.Status.Message,
		},
		SuccActionNum: r.SuccActionNum,
	}
	return tr.String()
}
