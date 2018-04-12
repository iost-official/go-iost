package core

import (
	"bytes"
	"fmt"
	"github.com/gomodule/redigo/redis"
	"sync"
)

//go:generate mockgen -destination mocks/mock_statepool.go -package core_mock -source utxo_pool.go -imports .=github.com/iost-official/prototype/core

/*
Current states of system ALERT: 正在施工，请不要调用
*/
type UTXOPool interface {
	Add(state UTXO) error
	Find(stateHash []byte) (UTXO, error)
	Del(StateHash []byte) error
	Transact(block *Block) error
	Flush() error
	Copy() UTXOPool
}

func BuildStatePoolCore(chain BlockChain) *StatePoolCore {
	var spc StatePoolCore
	spc.cli, _ = redis.Dial(Conn, DBAddr) // TODO : rebuild pool by block chain
	return &spc
}

type StatePoolImpl struct {
	*StatePoolCore
	addList []UTXO
	delList [][]byte
	base    *StatePoolImpl
}

var pCore *StatePoolCore
var once sync.Once

func NewUtxoPool(chain BlockChain) UTXOPool {
	if pCore == nil {
		once.Do(func() {
			pCore = BuildStatePoolCore(chain)
		})
	}
	spi := StatePoolImpl{
		StatePoolCore: pCore,
		addList:       make([]UTXO, 0),
		delList:       make([][]byte, 0),
		base:          nil,
	}
	return &spi
}

const (
	Conn   = "tcp"
	DBAddr = "localhost:6379"
)

func (sp *StatePoolImpl) Add(state UTXO) error {
	sp.addList = append(sp.addList, state)
	return nil
}

func (sp *StatePoolImpl) Find(stateHash []byte) (UTXO, error) {
	for _, u := range sp.addList {
		if bytes.Equal(u.Hash(), stateHash) {
			return u, nil
		}
	}

	for _, h := range sp.delList {
		if bytes.Equal(h, stateHash) {
			return UTXO{}, fmt.Errorf("not found")
		}
	}

	if sp.base != nil {
		return sp.base.Find(stateHash)
	}

	return sp.StatePoolCore.Find(stateHash)

	//reply, err := redis.Values(sp.cli.Do("HMGET", stateHash, "value", "script", "tx_hash"))
	//if err != nil {
	//	return s, err
	//}
	//_, err = redis.Scan(reply, &s.Value, &s.Script, s.BirthTxHash)
	//if err != nil {
	//	return s, err
	//}
}

func (sp *StatePoolImpl) Del(stateHash []byte) error {
	sp.delList = append(sp.delList, stateHash)
	return nil
}

func (sp *StatePoolImpl) Transact(block *Block) error {
	var txp TxPoolImpl
	txp.Decode(block.Content)
	txs, err := txp.GetSlice()
	if err != nil {
		return err
	}
	for _, tx := range txs {
		for _, in := range tx.Inputs {
			err = sp.Del(in.UTXOHash)
			if err != nil {
				return err
			}
		}
		for _, out := range tx.Outputs {
			err = sp.Add(out)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (sp *StatePoolImpl) Flush() error {
	if sp.base != nil {
		sp.base.Flush()
	}

	for _, h := range sp.delList {
		sp.StatePoolCore.Del(h)
	}

	for _, u := range sp.addList {
		sp.StatePoolCore.Add(u)
	}

	sp.base = nil
	sp.addList = make([]UTXO, 0)
	sp.delList = make([][]byte, 0)
	return nil
}

func (sp *StatePoolImpl) Copy() UTXOPool {
	spi := StatePoolImpl{
		base:          sp,
		addList:       make([]UTXO, 0),
		delList:       make([][]byte, 0),
		StatePoolCore: sp.StatePoolCore,
	}
	return &spi
}

type StatePoolCore struct {
	cli redis.Conn
}

func (spc *StatePoolCore) Add(state UTXO) error {
	_, err := spc.cli.Do("HMSET", state.Hash(),
		"value", state.Value,
		"script", state.Script,
		"tx_hash", state.BirthTxHash)
	if err != nil {
		return err
	}
	return nil
}

func (spc *StatePoolCore) Find(stateHash []byte) (UTXO, error) {
	s := UTXO{}
	reply, err := redis.Values(spc.cli.Do("HMGET", stateHash, "value", "script", "tx_hash"))
	if err != nil {
		return s, err
	}
	_, err = redis.Scan(reply, &s.Value, &s.Script, s.BirthTxHash)
	return s, err

}

func (spc *StatePoolCore) Del(StateHash []byte) error {
	_, err := spc.cli.Do("DEL", StateHash)
	if err != nil {
		return err
	}
	return nil
}

func (spc *StatePoolCore) Transact(block *Block) error {
	return nil
}
