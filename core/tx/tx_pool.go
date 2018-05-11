package tx

import (
	"fmt"
	"github.com/bouk/monkey"
	"github.com/iost-official/prototype/common"
	"github.com/iost-official/prototype/db"
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

/*
func (tp *TxPoolImpl) Copy(ttp *TxPool) error {
	var tttp *TxPoolImpl
	tttp = ttp
	tp.txMap = make(map[string]Tx)
	for k, v := range ttp.txMap {
		tp.txMap[k] = v
	}
	return nil
}
*/
func (tp *TxPoolImpl) Copy(ttp *TxPool) error {
	return nil
}
func (tp *TxPoolImpl) Size() int {
	return len(tp.txMap)
}

type TxPoolDbImpl struct {
	db db.Database
}

var txPrefix = []byte("t") //txPrefix+tx hash -> tx data

func NewTxPoolDbImpl() (*TxPoolDbImpl, error) {
	ldb, err := db.DatabaseFactor("ldb")
	if err != nil {
		return nil, fmt.Errorf("failed to init db %v", err)
	}

	return &TxPoolDbImpl{db: ldb}, nil
}

func (tp *TxPoolDbImpl) Add(tx Tx) error {
	hash := tx.Hash()
	err := tp.db.Put(append(txPrefix, hash...), tx.Encode())
	if err != nil {
		return fmt.Errorf("failed to Put tx: %v", err)
	}
	return nil
}

func (tp *TxPoolDbImpl) Del(tx Tx) error {
	return nil
}

func (tp *TxPoolDbImpl) Get(hash []byte) (*Tx, error) {
	txPtr := new(Tx)
	txData, err := tp.db.Get(append(txPrefix, hash...))
	if err != nil {

		return nil, fmt.Errorf("failed to Get the tx: %v", err)
	}

	guard := monkey.Patch(txPtr.Decode, func(_ []byte) error {
		return nil
	})
	defer guard.Unpatch()

	err = txPtr.Decode(txData) //something go wrong when call txPtr.Decode()
	if err != nil {

		return nil, fmt.Errorf("failed to Decode the tx: %v", err)
	}
	return txPtr, nil
}

func (tp *TxPoolDbImpl) GetSlice() ([]Tx, error) {
	return nil, nil
}

func (tp *TxPoolDbImpl) has(tx Tx) (bool, error) {
	return false, nil
}

func (tp *TxPoolDbImpl) Copy(ttp *TxPool) error {
	return nil
}

func (tp *TxPoolDbImpl) Size() int {
	return 0
}

func (tp *TxPoolDbImpl) Close() {
	tp.db.Close()
}
