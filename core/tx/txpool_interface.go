package tx

import (
	"errors"
)

type TxPool interface {
	Add(tx *Tx) error
	Del(tx *Tx) error
	Get(hash []byte) (*Tx, error)
	Has(tx *Tx) (bool, error)
	Top() (*Tx, error)
	Size() int
}

func TxPoolFactory(kind string) (TxPool, error) {
	switch kind {
	case "mem":
		return NewTxPoolImpl(), nil
	case "stack":
		return NewTxPoolStack()
	case "db":
		return NewTxPoolDb()
	}
	return nil, errors.New("this kind of TxPool not found")
}
