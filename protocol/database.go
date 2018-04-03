package protocol

import (
	"fmt"
	"github.com/iost-official/PrototypeWorks/iosbase"
)

//go:generate mockgen -destination mocks/mock_database.go -package protocol_mock github.com/iost-official/PrototypeWorks/protocol Database

type Database interface {
	NewViewSignal() (chan View, error)

	VerifyTx(tx iosbase.Tx) error
	VerifyTxWithCache(tx iosbase.Tx, cachePool iosbase.TxPool) error
	VerifyBlock(block *iosbase.Block) error
	VerifyBlockWithCache(block *iosbase.Block, cachePool iosbase.TxPool) error

	PushBlock(block *iosbase.Block) error

	GetStatePool() (iosbase.StatePool, error)
	GetBlockChain() (iosbase.BlockChain, error)
	GetCurrentView() (View, error)
}

func DatabaseFactory(target string, chain iosbase.BlockChain, pool iosbase.StatePool) (Database, error) {
	switch target {
	case "base":
		return &DatabaseImpl{
			bc:         chain,
			sp:         pool,
			chViewList: []chan View{},
		}, nil
	}
	return nil, fmt.Errorf("target Database not found")
}

type DatabaseImpl struct {
	bc   iosbase.BlockChain
	sp   iosbase.StatePool
	view View

	chViewList []chan View
}

func (d *DatabaseImpl) NewViewSignal() (chan View, error) {
	chView := make(chan View)
	d.chViewList = append(d.chViewList, chView)
	return chView, nil
}

func (d *DatabaseImpl) VerifyTx(tx iosbase.Tx) error {
	// here only existence of Tx inputs will be verified
	for _, in := range tx.Inputs {
		if _, err := d.sp.Find(in.StateHash); err != nil {
			return fmt.Errorf("some input not found")
		}
	}
	return nil
}
func (d *DatabaseImpl) VerifyTxWithCache(tx iosbase.Tx, cachePool iosbase.TxPool) error {
	err := d.VerifyTx(tx)
	if err != nil {
		return err
	}
	txs, _ := cachePool.GetSlice()
	for _, existedTx := range txs {
		if iosbase.Equal(existedTx.Hash(), tx.Hash()) {
			return fmt.Errorf("has included")
		}
		if TxConflict(existedTx, tx) {
			return fmt.Errorf("conflicted")
		} else if SliceIntersect(existedTx.Inputs, tx.Inputs) {
			return fmt.Errorf("conflicted")
		}
	}
	return nil
}
func (d *DatabaseImpl) VerifyBlock(block *iosbase.Block) error {
	var blkTxPool iosbase.TxPool
	blkTxPool.Decode(block.Content)

	txs, _ := blkTxPool.GetSlice()
	for i, tx := range txs {
		if i == 0 { // verify coinbase tx
			continue
		}
		err := d.VerifyTx(tx)
		if err != nil {
			return err
		}
	}
	return nil
}
func (d *DatabaseImpl) VerifyBlockWithCache(block *iosbase.Block, cachePool iosbase.TxPool) error {
	var blkTxPool iosbase.TxPool
	blkTxPool.Decode(block.Content)

	txs, _ := blkTxPool.GetSlice()
	for i, tx := range txs {
		if i == 0 { // TODO: verify coinbase tx
			continue
		}
		err := d.VerifyTxWithCache(tx, cachePool)
		if err != nil {
			return err
		}
	}
	return nil
}
func (d *DatabaseImpl) PushBlock(block *iosbase.Block) error {
	d.bc.Push(*block)
	var err error
	d.view, err = ViewFactory("dpos")
	if err != nil {
		return err
	}
	d.view.Init(d.bc)

	for _, chv := range d.chViewList {
		chv <- d.view
	}
	return nil
}
func (d *DatabaseImpl) GetStatePool() (iosbase.StatePool, error) {
	return d.sp, nil
}
func (d *DatabaseImpl) GetBlockChain() (iosbase.BlockChain, error) {
	return d.bc, nil
}
func (d *DatabaseImpl) GetCurrentView() (View, error) {
	return d.view, nil
}

func SliceEqualI(a, b []iosbase.TxInput) bool {
	if len(a) != len(b) {
		return false
	}
	for i := 0; i < len(a); i++ {
		if !iosbase.Equal(a[i].Hash(), b[i].Hash()) {
			return false
		}
	}
	return true
}

func SliceEqualS(a, b []iosbase.State) bool {
	if len(a) != len(b) {
		return false
	}
	for i := 0; i < len(a); i++ {
		if !iosbase.Equal(a[i].Hash(), b[i].Hash()) {
			return false
		}
	}
	return true
}

func SliceIntersect(a []iosbase.TxInput, b []iosbase.TxInput) bool {
	for _, ina := range a {
		for _, inb := range b {
			if iosbase.Equal(ina.Hash(), inb.Hash()) {
				return true
			}
		}
	}
	return false
}

func TxConflict(a, b iosbase.Tx) bool {
	if SliceEqualI(a.Inputs, b.Inputs) &&
		SliceEqualS(a.Outputs, b.Outputs) &&
		a.Recorder != b.Recorder {
		return true
	} else {
		return false
	}
}
