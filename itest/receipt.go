package itest

import "github.com/iost-official/go-iost/core/tx"

type Receipt struct {
	*tx.TxReceipt
}

func (r *Receipt) Success() bool {
	return r.Status.Code == tx.Success
}
