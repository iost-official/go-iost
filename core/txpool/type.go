package txpool

import (
	"sync"
	"time"

	"github.com/iost-official/Go-IOS-Protocol/core/block"
	"github.com/iost-official/Go-IOS-Protocol/core/blockcache"
	"github.com/iost-official/Go-IOS-Protocol/core/tx"
	"github.com/iost-official/Go-IOS-Protocol/metrics"
)

var (
	clearInterval = 10 * time.Second
	expiration    = int64(60 * time.Second)
	filterTime    = int64(expiration + expiration/2)
	//expiration    = 60*60*24*7

	metricsReceivedTxCount = metrics.NewCounter("iost_tx_received_count", []string{"from"})
)

type FRet uint

const (
	NotFound FRet = iota
	FoundPending
	FoundChain
)

type TFork uint

const (
	NotFork TFork = iota
	Fork
	ForkError
)

type TAddTx uint

const (
	Success TAddTx = iota
	TimeError
	VerifyError
	DupError
	GasPriceError
)

type ForkChain struct {
	NewHead       *blockcache.BlockCacheNode
	OldHead       *blockcache.BlockCacheNode
	ForkBlockHash []byte
}

type TxsList []*tx.Tx

func (s TxsList) Len() int { return len(s) }
func (s TxsList) Less(i, j int) bool {
	if s[i].GasPrice > s[j].GasPrice {
		return true
	}

	if s[i].GasPrice == s[j].GasPrice {
		if s[i].Time > s[j].Time {
			return false
		} else {
			return true
		}
	}
	return false
}
func (s TxsList) Swap(i, j int) { s[i], s[j] = s[j], s[i] }

func (s *TxsList) Push(x *tx.Tx) {
	*s = append(*s, x)
}

type blockTx struct {
	txMap      *sync.Map
	ParentHash []byte
	cTime      int64
}

func newBlockTx() *blockTx {
	b := &blockTx{
		txMap:      new(sync.Map),
		ParentHash: make([]byte, 32),
	}

	return b
}

func (b *blockTx) time() int64 {
	return b.cTime
}

func (b *blockTx) setTime(t int64) {
	b.cTime = t
}

func (b *blockTx) addBlock(ib *block.Block) {

	for _, v := range ib.Txs {
		b.txMap.Store(string(v.Hash()), nil)
	}

	//b.txMap.Range(func(key, value interface{}) bool {
	//	fmt.Println("range:", key.(string))
	//	return true
	//})

	b.ParentHash = ib.Head.ParentHash
}

func (b *blockTx) existTx(hash []byte) bool {

	_, r := b.txMap.Load(string(hash))

	return r
}
