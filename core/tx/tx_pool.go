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
		chain:            chain,
		blockTx:          make(blockTx),
		listTx:           make(listTx),
		pendingTx:        make(listTx),
		longestBlockHash: blockHashList{},
		filterTime:       int64(filterTime),
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

			// 超时交易丢弃
			if pool.txTimeOut(&tx) {
				continue
			}

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
			bhl := pool.getLongestChainBlockHash(pool.chain.LongestChain())
			pool.updateLongestChainBlockHash(bhl)

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

func (pool *TxPoolServer) txTimeOut(tx *Tx) bool {

	nTime := time.Now().Unix()
	txTime := tx.Time / 1e9

	if nTime-txTime > pool.filterTime {
		return true
	}
	return false
}

// 删除超时的交易
func (pool *TxPoolServer) delTimeOutTx() {
	pool.mu.Lock()
	defer pool.mu.Unlock()

	nTime := time.Now().Unix()

	for hash, tx := range pool.listTx {
		txTime := tx.Time / 1e9
		if nTime-txTime > pool.filterTime {
			delete(pool.listTx, hash)
		}
	}
}

func (pool *TxPoolServer) updatePending() {
	pool.mu.Lock()
	defer pool.mu.Unlock()

	pool.delTimeOutTx()

}

func (pool *TxPoolServer) txExistTxPool(tx *Tx) bool {
	pool.mu.Lock()
	defer pool.mu.Unlock()

	for hash := range pool.longestBlockHash.GetList() {
		txList := pool.blockTx.Get(string(hash))
		if _, bool := txList[tx.HashString()]; bool{
			return true
		}
	}

	return false
}

func (pool *TxPoolServer) getLongestChainBlockHash(chain block.Chain) *blockHashList {

	bhl := &blockHashList{}
	iter := chain.Iterator()
	for {
		block := iter.Next()
		if block == nil {
			break
		}
		log.Log.I("getLongestChainBlockHash , block Number: %v, witness: %v", block.Head.Number, block.Head.Witness)
		bhl.Add(block.HashString())
	}
	return bhl
}

func (pool *TxPoolServer) updateLongestChainBlockHash(bhl *blockHashList) {
	pool.mu.Lock()
	defer pool.mu.Unlock()

	// todo 增量更新判断，不用重新生成pending
	pool.longestBlockHash.Clear()

	for _, hash := range bhl.GetList() {
		pool.longestBlockHash.Add(hash)
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

type blockTx map[string]map[string]struct{}

func (b blockTx) Add(bl *block.Block) {
	trHash := make(map[string]struct{}, 0)
	for _, tr := range bl.Content {
		trHash[tr.HashString()] = struct{}{}
	}

	b[bl.HashString()] = trHash
}

func (b blockTx) Exist(hash string) bool {
	if _, bool := b[hash]; bool {
		return true
	}

	return false
}

func (b blockTx) Get(hash string) map[string]struct{} {

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

type blockHashList struct {
	blockList []string
}

func (b *blockHashList) Add(hash string) {

	b.blockList = append(b.blockList, hash)
}

func (b *blockHashList) Clear() {

	b.blockList = []string{}
}

func (b *blockHashList) GetList() []string {

	return b.blockList
}
