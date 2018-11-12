package itest

import (
	"math"
	"time"

	"github.com/iost-official/go-iost/core/tx"
)

// Constant of Transaction
var (
	GasLimit   = int64(10000)           // about 2000~10000 gas per tx
	GasPrice   = int64(100)             // 100 coin/gas
	Expiration = int64(math.MaxInt64)   // Max expired time is 90 seconds
	Delay      = int64(0 * time.Second) // No delay
	Signers    = make([]string, 0)      // No mutiple signers
)

// Transaction is the transaction object
type Transaction struct {
	*tx.Tx
}

// NewTransaction will return a new transaction by actions
func NewTransaction(actions []*tx.Action) *Transaction {
	t := tx.NewTx(
		actions,
		Signers,
		GasLimit,
		GasPrice,
		Expiration,
		Delay,
	)

	return &Transaction{
		Tx: t,
	}
}
