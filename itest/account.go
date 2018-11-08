package itest

import (
	"github.com/iost-official/go-iost/account"
	"github.com/iost-official/go-iost/core/tx"
)

type Account struct {
	ID      string
	Balance int64
	key     *Key
}

func (a *Account) Sign(t *Transaction) (*Transaction, error) {
	st, err := tx.SignTx(t.Tx, a.ID, []*account.KeyPair{a.key.KeyPair})
	if err != nil {
		return nil, err
	}

	transaction := &Transaction{
		Tx: st,
	}
	return transaction, nil
}
