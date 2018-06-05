package tx

import (
	"github.com/iost-official/prototype/core/message"
	"sync"
	"time"
)




type TxPoolServer struct{

	ChTx       chan message.Message // transactions of RPC and NET
	chBlock    chan message.Message

	blockTx		map[string]TransactionsList	// 缓存中block的交易
	all 		map[string]*Tx		// 所有的缓存交易
	pendingTx	map[string]*Tx		// 最长链上，去重的交易

	filterTime	time.Time
	mu sync.RWMutex
}

