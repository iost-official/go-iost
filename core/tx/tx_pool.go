package tx

import (
	"github.com/iost-official/prototype/common"
)

type TxPool interface {
	Add(tx Tx) error
	Del(tx Tx) error
	Get([]byte) (Tx, error)
	GetSlice() ([]Tx, error)
	Has(tx Tx) (bool, error)
	Copy(ttp TxPool) error
	Size() int32
}

type TxPoolImpl struct {
	txMap map[string]Tx
}

func NewTxPoolImpl() *TxPoolImpl {
	return &TxPoolImpl{txMap: make(map[string]Tx)}
}

func (tp *TxPoolImpl) Add(tx Tx) error {
	tp.txMap[common.Base58Encode(tx.Hash())] = tx
	return nil
}

func (tp *TxPoolImpl) Del(tx Tx) error {
	delete(tp.txMap, common.Base58Encode(tx.Hash()))
	return nil
}

func (tp *TxPoolImpl) GetSlice() ([]Tx, error) {
	var txs []Tx = make([]Tx, 0)
	for _, tx := range tp.txMap {
		txs = append(txs, tx)
	}
	return txs, nil
}

func (tp *TxPoolImpl) Has(tx Tx) (bool, error) {
	_, ok := tp.txMap[common.Base58Encode(tx.Hash())]
	return ok, nil
}

func (tp *TxPoolImpl) Copy(ttp TxPoolImpl) error {
	return nil
}

func (tp *TxPoolImpl) Size() int {
	return len(tp.txMap)
}
