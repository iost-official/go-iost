package itest

import (
	"math"
	"time"

	"github.com/iost-official/go-iost/core/tx"
)

// Constant of Transaction
const (
	GasLimit   = 10000                  // about 2000~10000 gas per tx
	GasPrice   = 100                    // 100 coin/gas
	Expiration = math.MaxInt64          // Max expired time is 90 seconds
	Delay      = int64(0 * time.Second) // No delay
)

type Transaction struct {
	*tx.Tx
}

func NewTransaction(actions []*tx.Action, signers []string) *Transaction {
	t := tx.NewTx(
		actions,
		signers,
		GasLimit,
		GasPrice,
		Expiration,
		Delay,
	)

	return &Transaction{
		Tx: t,
	}
}

func NewTransferTx(sender, recipient, amount string) *Transaction {
	return nil
}

func NewAccountTx() *Transaction {
	return nil
}
