package itest

import (
	"github.com/iost-official/go-iost/v3/common"
	"github.com/iost-official/go-iost/v3/core/tx"
	rpcpb "github.com/iost-official/go-iost/v3/rpc/pb"
)

// Receipt is the transaction receipt object
type Receipt struct {
	*tx.TxReceipt
}

// Success will return whether the receipt is successful
func (r *Receipt) Success() bool {
	return r.Status.Code == tx.Success
}

// NewReceiptFromPb returns a new Receipt instance from protobuffer struct.
func NewReceiptFromPb(tr *rpcpb.TxReceipt) *Receipt {
	ret := &tx.TxReceipt{
		TxHash:   common.Base58Decode(tr.TxHash),
		GasUsage: int64(tr.GasUsage * 100),
		RAMUsage: tr.RamUsage,
		Status: &tx.Status{
			Message: tr.Message,
			Code:    tx.StatusCode(tr.StatusCode),
		},
		Returns: tr.Returns,
	}
	for _, r := range tr.Receipts {
		ret.Receipts = append(ret.Receipts, &tx.Receipt{
			FuncName: r.FuncName,
			Content:  r.Content,
		})
	}
	return &Receipt{ret}
}
