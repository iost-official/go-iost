package tx

import (
	"github.com/iost-official/prototype/core/message"
	"sync"
	"github.com/iost-official/prototype/consensus/common"
	"github.com/iost-official/prototype/network"
	"fmt"
	"github.com/iost-official/prototype/log"
	"time"
	"github.com/iost-official/prototype/core/block"
)

var (
	clearTxInterval    = 8 * time.Second
	clearBlockInterval = 12 * time.Second
	filterTime         = 30
)

type TxPoolServer struct {
	ChTx    chan message.Message // transactions of RPC and NET
	chBlock chan message.Message // 上链的block数据

	chain  consensus_common.BlockCache // blockCache
	router network.Router

	blockTx   map[string][]string // 缓存中block的交易
	listTx    map[string]*Tx              // 所有的缓存交易
	pendingTx map[string]*Tx              // 最长链上，去重的交易

	filterTime int64 // 过滤交易的时间间隔
	mu         sync.RWMutex
}

func NewTxPoolServer(chain consensus_common.BlockCache) (*TxPoolServer, error) {

	p := &TxPoolServer{
		chain:      chain,
		blockTx:    make(map[string][]string),
		listTx:     make(map[string]*Tx),
		pendingTx:  make(map[string]*Tx),
		filterTime: int64(filterTime),
	}
	p.router = network.Route
	if p.router == nil {
		return nil, fmt.Errorf("failed to network.Route is nil")
	}

	return p, nil
}

func (pool *TxPoolServer) Start() {
	log.Log.I("TxPoolServer Start")
	go pool.loop()
}

func (pool *TxPoolServer) loop() {

	clearTx := time.NewTicker(clearTxInterval)
	defer clearTx.Stop()

	clearBlock := time.NewTicker(clearBlockInterval)
	defer clearBlock.Stop()

	for {
		select {
		case tr, ok := <-pool.ChTx:
			if !ok {
				return
			}

			var tx Tx
			tx.Decode(tr.Body)
			if consensus_common.VerifyTxSig(tx) {
				pool.addListTx(&tx)
			}

		case bl, ok := <-pool.chBlock:
			if !ok {
				return
			}

			var blk block.Block
			blk.Decode(bl.Body)

			pool.addBlockTx(&blk)
		}
	}
}

func (pool *TxPoolServer) addListTx(tx *Tx) {
	pool.mu.Lock()
	defer pool.mu.Unlock()

	if _, bool := pool.listTx[tx.HashString()]; !bool {
		pool.listTx[tx.HashString()] = tx
	}

}

// 保存一个block的所有交易数据
func (pool *TxPoolServer) addBlockTx(bl *block.Block) {
	pool.mu.Lock()
	defer pool.mu.Unlock()

	if _, bool := pool.blockTx[bl.HashString()]; bool {
		return
	}

	trHash := make([]string, 0)
	for _, tr:=range bl.Content{
		trHash = append(trHash, tr.HashString())
	}

	pool.blockTx[bl.HashString()] = trHash

}


type blockTx map[string][]string

func (b *blockTx) Add(bl *block.Block){

}

func (b *blockTx) IsHas(hash string) bool{
	if _, bool := b[hash]; bool {
		return
	}
}
