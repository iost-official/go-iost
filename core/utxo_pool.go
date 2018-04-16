package core

// UTXO池，可以存储在内存，用map实现，以后可以移到数据库
type UTXOPool interface {
	Add(state UTXO)
	Find(stateHash []byte) (UTXO, error) // 以后可以增加按字段搜索的功能，下同
	GetSlice() []UTXO                    // 来不及实现按字段搜索，就提供一个返回全部utxo的切片的接口，以后可以标记为deprecated，下同
	Del(StateHash []byte) error
	Transact(block *Block) error
	Flush()                                                                  // 保存当前池到数据库
	Copy() UTXOPool                                                          // 获得当前池的拷贝
	FindUTXO(address string) []UTXO                                          // 查询一个地址所有的UTXO
	FindSpendableOutputs(address string, amount int) (int, map[string][]int) // 对于一个地址，找到最少amount的UTXO
}

//import (
//	"bytes"
//	"fmt"
//	"github.com/gomodule/redigo/redis"
//	"sync"
//)
//
//func BuildStatePoolCore(chain BlockChain) *StatePoolCore {
//	var spc StatePoolCore
//	spc.cli, _ = redis.Dial(Conn, DBAddr) // TODO : rebuild pool by block chain
//	return &spc
//}
//
//type StatePoolImpl struct {
//	*StatePoolCore
//	addList []UTXO
//	delList [][]byte
//	base    *StatePoolImpl
//}
//
//var pCore *StatePoolCore
//var once sync.Once
//
//func NewUtxoPool(chain BlockChain) UTXOPool {
//	if pCore == nil {
//		once.Do(func() {
//			pCore = BuildStatePoolCore(chain)
//		})
//	}
//	spi := StatePoolImpl{
//		StatePoolCore: pCore,
//		addList:       make([]UTXO, 0),
//		delList:       make([][]byte, 0),
//		base:          nil,
//	}
//	return &spi
//}
//
//const (
//	Conn   = "tcp"
//	DBAddr = "localhost:6379"
//)
//
//func (sp *StatePoolImpl) Add(state UTXO) error {
//	sp.addList = append(sp.addList, state)
//	return nil
//}
//
//func (sp *StatePoolImpl) Find(stateHash []byte) (UTXO, error) {
//	for _, u := range sp.addList {
//		if bytes.Equal(u.Hash(), stateHash) {
//			return u, nil
//		}
//	}
//
//	for _, h := range sp.delList {
//		if bytes.Equal(h, stateHash) {
//			return UTXO{}, fmt.Errorf("not found")
//		}
//	}
//
//	if sp.base != nil {
//		return sp.base.Find(stateHash)
//	}
//
//	return sp.StatePoolCore.Find(stateHash)
//
//	//reply, err := redis.Values(sp.cli.Do("HMGET", stateHash, "value", "script", "tx_hash"))
//	//if err != nil {
//	//	return s, err
//	//}
//	//_, err = redis.Scan(reply, &s.Value, &s.Script, s.BirthTxHash)
//	//if err != nil {
//	//	return s, err
//	//}
//}
//
//func (sp *StatePoolImpl) Del(stateHash []byte) error {
//	sp.delList = append(sp.delList, stateHash)
//	return nil
//}
//
//func (sp *StatePoolImpl) Transact(block *Block) error {
//	var txp TxPoolImpl
//	txp.Decode(block.Content)
//	txs, err := txp.GetSlice()
//	if err != nil {
//		return err
//	}
//	for _, tx := range txs {
//		for _, in := range tx.Inputs {
//			err = sp.Del(in.UTXOHash)
//			if err != nil {
//				return err
//			}
//		}
//		for _, out := range tx.Outputs {
//			err = sp.Add(out)
//			if err != nil {
//				return err
//			}
//		}
//	}
//	return nil
//}
//
//func (sp *StatePoolImpl) Flush() error {
//	if sp.base != nil {
//		sp.base.Flush()
//	}
//
//	for _, h := range sp.delList {
//		sp.StatePoolCore.Del(h)
//	}
//
//	for _, u := range sp.addList {
//		sp.StatePoolCore.Add(u)
//	}
//
//	sp.base = nil
//	sp.addList = make([]UTXO, 0)
//	sp.delList = make([][]byte, 0)
//	return nil
//}
//
//func (sp *StatePoolImpl) Copy() UTXOPool {
//	spi := StatePoolImpl{
//		base:          sp,
//		addList:       make([]UTXO, 0),
//		delList:       make([][]byte, 0),
//		StatePoolCore: sp.StatePoolCore,
//	}
//	return &spi
//}
//
//type StatePoolCore struct {
//	cli redis.Conn
//}
//
//func (spc *StatePoolCore) Add(state UTXO) error {
//	_, err := spc.cli.Do("HMSET", state.Hash(),
//		"value", state.Value,
//		"script", state.Script,
//		"tx_hash", state.BirthTxHash)
//	if err != nil {
//		return err
//	}
//	return nil
//}
//
//func (spc *StatePoolCore) Find(stateHash []byte) (UTXO, error) {
//	s := UTXO{}
//	reply, err := redis.Values(spc.cli.Do("HMGET", stateHash, "value", "script", "tx_hash"))
//	if err != nil {
//		return s, err
//	}
//	_, err = redis.Scan(reply, &s.Value, &s.Script, s.BirthTxHash)
//	return s, err
//
//}
//
//func (spc *StatePoolCore) Del(StateHash []byte) error {
//	_, err := spc.cli.Do("DEL", StateHash)
//	if err != nil {
//		return err
//	}
//	return nil
//}
//
//func (spc *StatePoolCore) Transact(block *Block) error {
//	return nil
//}
