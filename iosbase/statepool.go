package iosbase

import (
	"github.com/gomodule/redigo/redis"
)

//go:generate mockgen -destination mocks/mock_statepool.go -package iosbase_mock -source statepool.go -imports .=github.com/iost-official/Go-IOS-Protocol/iosbase

/*
Current states of system
*/
type StatePool interface {
	Init() error
	Close() error
	Add(state State) error
	Find(stateHash []byte) (State, error)
	Del(StateHash []byte) error
	Transact(block *Block) error
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

func (sp *StatePoolImpl) Add(state State) error {
	_, err := sp.cli.Do("HMSET", state.Hash(),
		"value", state.Value,
		"script", state.Script,
		"tx_hash", state.BirthTxHash)
	if err != nil {
		return err
	}
	return nil
}

func (sp *StatePoolImpl) Find(stateHash []byte) (State, error) {
	var s State
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
