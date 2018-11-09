package itest

import (
	"encoding/json"

	"github.com/iost-official/go-iost/account"
	"github.com/iost-official/go-iost/core/tx"
	"github.com/iost-official/go-iost/crypto"
)

type Account struct {
	ID      string
	Balance int64
	key     *Key
}

type AccountJSON struct {
	ID        string           `json:"id"`
	Seckey    []byte           `json:"seckey"`
	Algorithm crypto.Algorithm `json:"algorithm"`
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

func (a *Account) UnmarshalJSON(b []byte) error {
	aux := &AccountJSON{}
	err := json.Unmarshal(b, aux)
	if err != nil {
		return err
	}

	a.ID = aux.ID
	a.key = NewKey(aux.Seckey, aux.Algorithm)
	return nil
}

func (a *Account) MarshalJSON() ([]byte, error) {
	aux := &AccountJSON{
		ID:        a.ID,
		Seckey:    a.key.Seckey,
		Algorithm: a.key.Algorithm,
	}
	return json.Marshal(aux)
}
