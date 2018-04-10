package core

import (
	"github.com/gomodule/redigo/redis"
)

//go:generate mockgen -destination mocks/mock_statepool.go -package core_mock -source utxo_pool.go -imports .=github.com/iost-official/prototype/core

/*
Current states of system ALERT: 正在施工，请不要调用
*/
type UTXOPool interface {
	Init() error
	Close() error
	Add(state UTXO) error
	Find(stateHash []byte) (UTXO, error)
	Del(StateHash []byte) error
	Transact(block *Block) error
	Patch(diff UTXOPoolDiff) error
	Copy() UTXOPool
}

type StatePoolImpl struct {
	cli redis.Conn
}

const (
	Conn   = "tcp"
	DBAddr = "localhost:6379"
)

func (sp *StatePoolImpl) Init() error {
	var err error
	sp.cli, err = redis.Dial(Conn, DBAddr)
	if err != nil {
		return err
	}
	return nil
}

func (sp *StatePoolImpl) Close() error {
	if sp.cli != nil {
		sp.cli.Close()
	}
	return nil
}

func (sp *StatePoolImpl) Add(state UTXO) error {
	_, err := sp.cli.Do("HMSET", state.Hash(),
		"value", state.Value,
		"script", state.Script,
		"tx_hash", state.BirthTxHash)
	if err != nil {
		return err
	}
	return nil
}

func (sp *StatePoolImpl) Find(stateHash []byte) (UTXO, error) {
	var s UTXO
	reply, err := redis.Values(sp.cli.Do("HMGET", stateHash, "value", "script", "tx_hash"))
	if err != nil {
		return s, err
	}
	_, err = redis.Scan(reply, &s.Value, &s.Script, s.BirthTxHash)
	if err != nil {
		return s, err
	}
	return s, nil
}

func (sp *StatePoolImpl) Del(stateHash []byte) error {
	_, err := sp.cli.Do("DEL", stateHash)
	if err != nil {
		return err
	}
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
			err = sp.Del(in.StateHash)
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

func (sp *StatePoolImpl) Patch(diff UTXOPoolDiff) error {
	for _, utxo := range diff.del {
		err := sp.Del(utxo.Hash())
		if err != nil {
			return err
		}
	}
	for _, utxo := range diff.add {
		sp.Add(utxo)
	}
	return nil
}

func (sp *StatePoolImpl) Copy() UTXOPool {
	return sp
}
