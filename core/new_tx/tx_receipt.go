package tx

import (
	"github.com/golang/protobuf/proto"
	"github.com/iost-official/Go-IOS-Protocol/common"
)

/**
 * Describtion: tx
 * User: wangyu
 * Date: 18-7-30
 */

type StatusCode int32

// tx execution result
const (
	Success StatusCode = iota
	ErrorGasRunOut
	ErrorBalanceNotEnough
	ErrorOverFlow // number calculation over flow
	ErrorParamter // paramter mismatch when calling function
	ErrorRuntime  // runtime error
	ErrorUnknown  // other errors
)

type Status struct {
	Code    StatusCode
	Message string
}

type ReceiptType int32

const (
	SystemDefined ReceiptType = iota
	UserDefined
)

// Receipt generated when applying transaction
type Receipt struct {
	Type    ReceiptType // system defined or user defined receipt type
	Content string      // can be a raw string or a json string
}

// TxReceipt Transaction Receipt 实现
type TxReceipt struct {
	TxHash   []byte
	GasUsage uint64
	// 目前只收gas，这些可以先没有
	/*
		CpuTimeUsage    uint64
		NetUsage    uint64
		RAMUsage    uint64
	*/
	Status        Status
	SuccActionNum int32 // 执行成功的 action 个数
	Receipts      []Receipt
}

// NewTxReceipt generate tx receipt for a tx hash
func NewTxReceipt(txHash []byte) TxReceipt {
	var status = Status{
		Code:    Success,
		Message: "",
	}
	return TxReceipt{
		TxHash:        txHash,
		GasUsage:      0,
		Status:        status,
		SuccActionNum: 0,
		Receipts:      []Receipt{},
	}
}

func (r *TxReceipt) Encode() []byte {
	tr := &TxReceiptRaw{
		TxHash:   r.TxHash,
		GasUsage: r.GasUsage,
		Status: &StatusRaw{
			Code:    int32(r.Status.Code),
			Message: r.Status.Message,
		},
		SuccActionNum: r.SuccActionNum,
	}
	for _, re := range r.Receipts {
		tr.Receipts = append(tr.Receipts, &ReceiptRaw{
			Type:    int32(re.Type),
			Content: re.Content,
		})
	}
	b, err := proto.Marshal(tr)
	if err != nil {
		panic(err)
	}
	return b
}

func (r *TxReceipt) Decode(b []byte) error {
	tr := &TxReceiptRaw{}
	err := proto.Unmarshal(b, tr)
	if err != nil {
		return err
	}
	r.TxHash = tr.TxHash
	r.GasUsage = tr.GasUsage
	r.Status = Status{
		Code:    StatusCode(tr.Status.Code),
		Message: tr.Status.Message,
	}
	r.SuccActionNum = tr.SuccActionNum
	r.Receipts = []Receipt{}
	for _, re := range tr.Receipts {
		r.Receipts = append(r.Receipts, Receipt{
			Type:    ReceiptType(re.Type),
			Content: re.Content,
		})
	}
	return nil
}

// hash
func (r *TxReceipt) Hash() []byte {
	return common.Sha256(r.Encode())
}

func (r *TxReceipt) String() string {
	tr := &TxReceiptRaw{
		TxHash:   r.TxHash,
		GasUsage: r.GasUsage,
		Status: &StatusRaw{
			Code:    int32(r.Status.Code),
			Message: r.Status.Message,
		},
		SuccActionNum: r.SuccActionNum,
	}
	return tr.String()
}
