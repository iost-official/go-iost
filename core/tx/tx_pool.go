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

	blockTx   blockTx // 缓存中block的交易
	listTx    listTx  // 所有的缓存交易
	pendingTx listTx  // 最长链上，去重的交易

	longestBlockHash blockHashList

	filterTime int64 // 过滤交易的时间间隔
	mu         sync.RWMutex
}

func NewTxPoolServer(chain consensus_common.BlockCache) (*TxPoolServer, error) {

	p := &TxPoolServer{
		chain:      chain,
		blockTx:    make(blockTx),
		listTx:     make(listTx),
		pendingTx:  make(listTx),
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
			// 根据最长链计算 pending tx
			pool.chain.LongestChain()
		}
	}
}

func (pool *TxPoolServer) addListTx(tx *Tx) {
	pool.mu.Lock()
	defer pool.mu.Unlock()

	if !pool.listTx.Exist(tx.HashString()) {
		pool.listTx.Add(tx)
	}

}

func (pool *TxPoolServer) getLongestChainBlockHash(chain block.Chain) blockHashList {

	iter := chain.Iterator()
	for {
		block := iter.Next()
		if block == nil {
			break
		}
		log.Log.I("getLongestChainBlockHash , block Number: %v, witness: %v", block.Head.Number, block.Head.Witness)

	}
}


// 保存一个block的所有交易数据
func (pool *TxPoolServer) addBlockTx(bl *block.Block) {
	pool.mu.Lock()
	defer pool.mu.Unlock()

	if !pool.blockTx.Exist(bl.HashString()) {
		pool.blockTx.Add(bl)
	}
}

type blockTx map[string][]string

func (b blockTx) Add(bl *block.Block) {
	trHash := make([]string, 0)
	for _, tr := range bl.Content {
		trHash = append(trHash, tr.HashString())
	}

	b[bl.HashString()] = trHash
}

func (b blockTx) Exist(hash string) bool {
	if _, bool := b[hash]; bool {
		return true
	}

	return false
}

func (b blockTx) Get(hash string) []string {

	return b[hash]
}

type listTx map[string]*Tx

func (l listTx) Add(tx *Tx) {

	l[tx.HashString()] = tx
}

func (l listTx) Exist(hash string) bool {
	if _, bool := l[hash]; bool {
		return true
	}

	return false
}

func (l listTx) Get(hash string) *Tx {

	return l[hash]
}

type blockHashList struct{
	blockList []string
}

func (b *blockHashList) Add(hash string) {

	b.blockList = append(b.blockList, hash)
}