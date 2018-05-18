package tx

import (
	"errors"
	"fmt"
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
		if TxDb == nil {
			panic(fmt.Errorf("TxDb cannot be nil pointer"))
		}
		return TxDb, nil
	}
	return nil, errors.New("this kind of TxPool not found")
}
