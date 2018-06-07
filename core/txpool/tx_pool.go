package txpool

import (
	"fmt"
	"github.com/iost-official/prototype/core/block"
	"github.com/iost-official/prototype/core/message"
	"github.com/iost-official/prototype/log"
	"github.com/iost-official/prototype/network"
	"sort"
	"sync"
	"time"
	"github.com/iost-official/prototype/core/tx"
	"github.com/iost-official/prototype/core/blockcache"
	"github.com/iost-official/prototype/consensus/common"
)

var (
	clearTxInterval    = 8 * time.Second
	clearBlockInterval = 12 * time.Second
	filterTime         = 30
)

type TxPoolServer struct {
	chTx           chan message.Message // transactions of RPC and NET
	chBlock        chan message.Message // 上链的block数据
	chConfirmBlock chan block.Block

	chain  blockcache.BlockCache // blockCache
	router network.Router

	blockTx   blockTx // 缓存中block的交易
	listTx    listTx  // 所有的缓存交易
	pendingTx listTx  // 最长链上，去重的交易

	longestBlockHash blockHashList

	filterTime int64 // 过滤交易的时间间隔
	mu         sync.RWMutex
}

func NewTxPoolServer(chain blockcache.BlockCache, chConfirmBlock chan block.Block) (*TxPoolServer, error) {

	p := &TxPoolServer{
		chain:            chain,
		chConfirmBlock:   chConfirmBlock,
		blockTx:          blockTx{},
		listTx:           make(listTx),
		pendingTx:        make(listTx),
		longestBlockHash: blockHashList{},
		filterTime:       int64(filterTime),
	}
	p.router = network.Route
	if p.router == nil {
		return nil, fmt.Errorf("failed to network.Route is nil")
	}

	//	Tx chan init
	var err error
	p.chTx, err = p.router.FilteredChan(network.Filter{
		AcceptType: []network.ReqType{
			network.ReqPublishTx,
		}})
	if err != nil {
		return nil, err
	}

	return p, nil
}

func (pool *TxPoolServer) Start() {
	log.Log.I("TxPoolServer Start")
	go pool.loop()
}

func (pool *TxPoolServer) Stop() {
	log.Log.I("TxPoolServer Stop")
	close(pool.chTx)
	close(pool.chBlock)
}

func (pool *TxPoolServer) loop() {

	pool.initBlockTx()

	clearTx := time.NewTicker(clearTxInterval)
	defer clearTx.Stop()

	clearBlock := time.NewTicker(clearBlockInterval)
	defer clearBlock.Stop()

	for {
		select {
		case tr, ok := <-pool.chTx:
			if !ok {
				return
			}

			var tx tx.Tx
			tx.Decode(tr.Body)

			// 超时交易丢弃
			if pool.txTimeOut(&tx) {
				continue
			}
			if blockcache.VerifyTxSig(tx) {
				pool.addListTx(&tx)
			}

		case bl, ok := <-pool.chBlock: // 可以上链的block
			if !ok {
				return
			}

			var blk block.Block
			blk.Decode(bl.Body)

			pool.addBlockTx(&blk)
			// 根据最长链计算 pending tx
			bhl := pool.blockHash(pool.chain.LongestChain())
			pool.updateBlockHash(bhl)
			// todo 可以在外部要集合的时候在调用
			pool.updatePending()
		case <-clearBlock.C:
			pool.delTimeOutBlockTx()
		case <-clearTx.C:
			pool.delTimeOutTx()
		}
	}
}

func (pool *TxPoolServer) PendingTransactions() tx.TransactionsList {
	pool.mu.Lock()
	defer pool.mu.Unlock()

	var pendingList tx.TransactionsList
	for _, tx := range pool.pendingTx {
		pendingList = append(pendingList, tx)
	}

	// 排序
	sort.Sort(pendingList)

	return pendingList
}

func (pool *TxPoolServer) Transaction(hash string) *tx.Tx {
	return pool.listTx.Get(hash)
}

func (pool *TxPoolServer) ExistTransaction(hash string) bool {
	return pool.listTx.Exist(hash)
}

func (pool *TxPoolServer) TransactionNum() int {
	return len(pool.listTx)
}

func (pool *TxPoolServer) PendingTransactionNum() int {
	return len(pool.pendingTx)
}

func (pool *TxPoolServer) BlockTxNum() int {
	return len(pool.blockTx.blkTx)
}

// 初始化blocktx,缓存验证一笔交易，是否已经存在
func (pool *TxPoolServer) initBlockTx() {
	chain := pool.chain.BlockChain()
	timeNow := time.Now().Unix()

	for i := chain.Length() - 1; i > 0; i-- {
		blk := chain.GetBlockByNumber(i)
		if blk == nil {
			return
		}

		t := pool.slotToSec(blk.Head.Time)
		if timeNow-t >= pool.filterTime {
			pool.blockTx.Add(blk)
		}
	}

}

func (pool *TxPoolServer) slotToSec(t int64) int64 {
	slot := consensus_common.Timestamp{Slot: t}
	return slot.ToUnixSec()
}

func (pool *TxPoolServer) addListTx(tx *tx.Tx) {
	pool.mu.Lock()
	defer pool.mu.Unlock()

	if !pool.listTx.Exist(tx.HashString()) {
		pool.listTx.Add(tx)
	}

}

func (pool *TxPoolServer) txTimeOut(tx *tx.Tx) bool {

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

// 删除超时的交易
// 小于区块确定时间 && 小于当前减去过滤时间
func (pool *TxPoolServer) delTimeOutBlockTx() {
	pool.mu.Lock()
	defer pool.mu.Unlock()

	nTime := time.Now().Unix()
	chain := pool.chain.BlockChain()
	blk := chain.Top()

	var confirmTime int64
	if blk != nil {
		confirmTime = pool.slotToSec(blk.Head.Time)
	}

	for hash, t := range pool.blockTx.blkTime {

		if t < confirmTime && nTime-pool.filterTime > t {
			pool.blockTx.Del(hash)
		}
	}
}

func (pool *TxPoolServer) updatePending() {
	pool.mu.Lock()
	defer pool.mu.Unlock()

	pool.pendingTx = nil
	pool.delTimeOutTx()

	for hash, tr := range pool.listTx {
		if !pool.txExistTxPool(hash) {
			pool.pendingTx.Add(tr)
		}
	}
}

func (pool *TxPoolServer) txExistTxPool(hash string) bool {
	pool.mu.Lock()
	defer pool.mu.Unlock()

	for hash := range pool.longestBlockHash.GetList() {
		txList := pool.blockTx.Tx(string(hash))
		if _, b := txList[string(hash)]; b {
			return true
		}
	}

	return false
}

func (pool *TxPoolServer) blockHash(chain block.Chain) *blockHashList {

	bhl := &blockHashList{}
	iter := chain.Iterator()
	for {
		blk := iter.Next()
		if blk == nil {
			break
		}
		log.Log.I("getLongestChainBlockHash , blk Number: %v, witness: %v", blk.Head.Number, blk.Head.Witness)
		bhl.Add(blk.HashString())
	}
	return bhl
}

func (pool *TxPoolServer) updateBlockHash(bhl *blockHashList) {
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

type blockTx struct {
	blkTx   map[string]map[string]struct{}
	blkTime map[string]int64
}

func (b *blockTx) Add(bl *block.Block) {
	trHash := make(map[string]struct{}, 0)
	for _, tr := range bl.Content {
		trHash[tr.HashString()] = struct{}{}
	}

	b.blkTx[bl.HashString()] = trHash
	slot :=consensus_common.Timestamp{Slot: bl.Head.Time}
	b.blkTime[bl.HashString()] = slot.ToUnixSec()

}

func (b *blockTx) Exist(hash string) bool {
	if _, b := b.blkTx[hash]; b {
		return true
	}

	return false
}

func (b blockTx) Tx(hash string) map[string]struct{} {

	return b.blkTx[hash]
}

func (b blockTx) Time(hash string) int64 {

	return b.blkTime[hash]
}

func (b blockTx) Del(hash string) {

	delete(b.blkTime, hash)
	delete(b.blkTx, hash)
}

type listTx map[string]*tx.Tx

func (l listTx) Add(tx *tx.Tx) {

	l[tx.HashString()] = tx
}

func (l listTx) Exist(hash string) bool {
	if _, b := l[hash]; b {
		return true
	}

	return false
}

func (l listTx) Get(hash string) *tx.Tx {

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
