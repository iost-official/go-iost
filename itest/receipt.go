package itest

import "github.com/iost-official/go-iost/core/tx"

// Receipt is the transaction receipt object
type Receipt struct {
	*tx.TxReceipt
}

// Success will return weather the receipt is successful
func (r *Receipt) Success() bool {
	return r.Status.Code == tx.Success
}
