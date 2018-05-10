package tx

import (
	"github.com/iost-official/prototype/common"
)

type TxPool interface {
	Add(tx Tx) error
	Del(tx Tx) error
	Get(hash []byte) (*Tx, error)
	GetSlice() ([]Tx, error)
	Has(tx Tx) (bool, error)
	Copy(ttp *TxPool) error
	Size() int
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

func (tp *TxPoolImpl) Get(hash []byte) (*Tx, error) {
	tx, ok := tp.txMap[common.Base58Encode(hash)]
	return &tx, nil
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

func (tp *TxPoolImpl) Copy(ttp *TxPool) error {
	var tttp *TxPoolImpl
	tttp = ttp
	tp.txMap = make(map[string]Tx)
	for k, v := range ttp.txMap {
		tp.txMap[k] = v
	}
	return nil
}

func (tp *TxPoolImpl) Size() int {
	return len(tp.txMap)
}
