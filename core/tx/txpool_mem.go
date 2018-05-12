package tx

import (
	"errors"

	"github.com/iost-official/prototype/common"
)

type TxPoolImpl struct {
	txMap map[string]*Tx
}

func NewTxPoolImpl() *TxPoolImpl {
	return &TxPoolImpl{txMap: make(map[string]*Tx)}
}

func (tp *TxPoolImpl) Add(tx *Tx) error {
	tp.txMap[common.Base58Encode(tx.Hash())] = tx
	return nil
}

func (tp *TxPoolImpl) Del(tx *Tx) error {
	delete(tp.txMap, common.Base58Encode(tx.Hash()))
	return nil
}

func (tp *TxPoolImpl) Get(hash []byte) (*Tx, error) {
	tx, _ := tp.txMap[common.Base58Encode(hash)]
	return tx, nil
}

func (tp *TxPoolImpl) Top() (*Tx, error) {
	for _, tx := range tp.txMap {
		return tx, nil
	}
	return nil, errors.New("Empty")
}

func (tp *TxPoolImpl) Has(tx *Tx) (bool, error) {
	_, ok := tp.txMap[common.Base58Encode(tx.Hash())]
	return ok, nil
}

func (tp *TxPoolImpl) Size() int {
	return len(tp.txMap)
}
