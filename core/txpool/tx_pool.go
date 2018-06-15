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
	"github.com/prometheus/client_golang/prometheus"
)

var (
	clearInterval = 8 * time.Second
	filterTime    = 40
	//filterTime    = 60*60*24*7

	receivedTransactionCount = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "received_transaction_count",
			Help: "Count of received transaction by current node",
		},
	)
)

func init() {
	prometheus.MustRegister(receivedTransactionCount)
}

type TxPoolServer struct {
	chTx           chan message.Message // transactions of RPC and NET
	chConfirmBlock chan *block.Block

	chain  blockcache.BlockCache // blockCache
	router network.Router

	blockTx   blockTx // 缓存中block的交易
	listTx    listTx  // 所有的缓存交易
	pendingTx listTx  // 最长链上，去重的交易

	checkIterateBlockHash blockHashList

	filterTime int64 // 过滤交易的时间间隔
	mu         sync.RWMutex
}

var TxPoolS *TxPoolServer

func NewTxPoolServer(chain blockcache.BlockCache, chConfirmBlock chan *block.Block) (*TxPoolServer, error) {

	p := &TxPoolServer{
		chain:          chain,
		chConfirmBlock: chConfirmBlock,
		blockTx: blockTx{
			blkTx:   make(map[string]*hashMap),
			blkTime: make(map[string]int64),
		},
		listTx:                listTx{list: make(map[string]*tx.Tx)},
		pendingTx:             listTx{list: make(map[string]*tx.Tx)},
		checkIterateBlockHash: blockHashList{blockList: make(map[string]struct{}, 0)},
		filterTime:            int64(filterTime),
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

	TxPoolS = p
	return p, nil
}

func (pool *TxPoolServer) Start() {
	log.Log.I("TxPoolServer Start")
	go pool.loop()
}

func (pool *TxPoolServer) Stop() {
	log.Log.I("TxPoolServer Stop")
	close(pool.chTx)
	close(pool.chConfirmBlock)
}

func (pool *TxPoolServer) loop() {

	pool.initBlockTx()

	clearTx := time.NewTicker(clearInterval)
	defer clearTx.Stop()

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
				//log.Log.I("tx timeout:%v", tx.Time)
				continue
			}

			if blockcache.VerifyTxSig(tx) {
				pool.mu.Lock()
				pool.addListTx(&tx)
				receivedTransactionCount.Inc()
				pool.mu.Unlock()
			}

		case bl, ok := <-pool.chConfirmBlock: // 可以上链的block
			if !ok {
				return
			}
			pool.addBlockTx(bl)
			// 根据最长链计算 pending tx
			bhl := pool.blockHash(pool.chain.LongestChain())
			pool.updateBlockHash(bhl)
			// todo 可以在外部要集合的时候在调用
			//pool.updatePending()
		case <-clearTx.C:
			pool.delTimeOutTx()
			pool.delTimeOutBlockTx()
		}
	}
}

func (pool *TxPoolServer) AddTransaction(tx message.Message) {
	pool.chTx <- tx
}

func (pool *TxPoolServer) PendingTransactions() tx.TransactionsList {

	pool.updatePending()

	pool.mu.Lock()
	defer pool.mu.Unlock()

	var pendingList tx.TransactionsList
	list := pool.pendingTx.GetList()
	pool.pendingTx.Lock()

	for _, tx := range list {
		pendingList = append(pendingList, tx)
	}
	pool.pendingTx.UnLock()
	// 排序
	sort.Sort(pendingList)

	return pendingList
}

func (pool *TxPoolServer) Transaction(hash string) *tx.Tx {
	pool.mu.Lock()
	defer pool.mu.Unlock()

	return pool.listTx.Get(hash)
}

func (pool *TxPoolServer) ExistTransaction(hash string) bool {
	pool.mu.Lock()
	defer pool.mu.Unlock()

	return pool.listTx.Exist(hash)
}

func (pool *TxPoolServer) TransactionNum() int {

	return pool.listTx.Len()
}

func (pool *TxPoolServer) PendingTransactionNum() int {
	return pool.pendingTx.Len()
}

func (pool *TxPoolServer) BlockTxNum() int {
	return pool.blockTx.Len()
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
			pool.checkIterateBlockHash.Add(blk.HashID())
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

	if !pool.listTx.Exist(tx.TxID()) {
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

	nTime := time.Now().Unix()
	hashList := make([]string, 0)

	list := pool.listTx.GetList()
	pool.listTx.Lock()
	for hash, tx := range list {
		txTime := tx.Time / 1e9
		if nTime-txTime > pool.filterTime {
			hashList = append(hashList, hash)
		}
	}
	pool.listTx.UnLock()
	for _, hash := range hashList {
		pool.listTx.Del(hash)
	}

}

// 删除超时的交易
// 小于区块确定时间 && 小于当前减去过滤时间
func (pool *TxPoolServer) delTimeOutBlockTx() {

	nTime := time.Now().Unix()
	chain := pool.chain.BlockChain()
	blk := chain.Top()

	var confirmTime int64
	if blk != nil {
		confirmTime = pool.slotToSec(blk.Head.Time)
	}

	list := pool.blockTx.GetListTime()

	hashList := make([]string, 0)
	pool.blockTx.Lock()
	for hash, t := range list {

		if t < confirmTime && nTime-pool.filterTime > t {
			hashList = append(hashList, hash)
		}
	}
	pool.blockTx.UnLock()

	for _, hash := range hashList {

		pool.blockTx.Del(hash)
		pool.checkIterateBlockHash.Del(hash)
	}
}

func (pool *TxPoolServer) updatePending() {

	pool.mu.Lock()
	defer pool.mu.Unlock()

	pool.pendingTx.Clear()

	list := pool.listTx.GetList()
	pool.listTx.Lock()
	for hash, tr := range list {
		if !pool.txExistTxPool(hash) {
			//fmt.Println("Add pending tr hash:",tr.TxID(), " tr nonce:", tr.Nonce)
			pool.pendingTx.Add(tr)
		}
	}
	pool.listTx.UnLock()
}

func (pool *TxPoolServer) txExistTxPool(hash string) bool {

	for blockHash := range pool.checkIterateBlockHash.GetList() {
		//fmt.Println("##blockhash: ", blockHash)
		txList := pool.blockTx.TxList(string(blockHash))
		if txList != nil {
			if b := txList.Exist(hash); b {
				return true
			}
		}
	}

	return false
}

func (pool *TxPoolServer) blockHash(chain block.Chain) *blockHashList {

	bhl := &blockHashList{blockList: make(map[string]struct{}, 0)}
	iter := chain.Iterator()
	for {
		blk := iter.Next()
		if blk == nil {
			break
		}
		//log.Log.I("getLongestChainBlockHash , blk Number: %v, witness: %v", blk.Head.Number, blk.Head.Witness)
		bhl.Add(blk.HashID())
	}
	return bhl
}

func (pool *TxPoolServer) updateBlockHash(bhl *blockHashList) {

	for hash := range bhl.GetList() {
		pool.checkIterateBlockHash.Add(hash)
	}

}

// 保存一个block的所有交易数据
func (pool *TxPoolServer) addBlockTx(bl *block.Block) {

	pool.mu.Lock()
	defer pool.mu.Unlock()

	if !pool.blockTx.Exist(bl.HashID()) {
		pool.blockTx.Add(bl)
	}
}

type hashMap struct {
	hashList map[string]struct{}
	smu      sync.RWMutex
}

func (h *hashMap) Add(hash string) {
	h.smu.Lock()
	defer h.smu.Unlock()

	h.hashList[hash] = struct{}{}
}

func (h *hashMap) Exist(txHash string) bool {
	h.smu.Lock()
	defer h.smu.Unlock()

	if _, b := h.hashList[txHash]; b {
		return true
	}

	return false
}

func (h *hashMap) Del(hash string) {
	h.smu.Lock()
	defer h.smu.Unlock()

	delete(h.hashList, hash)
}

func (h *hashMap) Clear() {
	h.smu.Lock()
	defer h.smu.Unlock()

	h.hashList = nil
	h.hashList = make(map[string]struct{})
}

type blockTx struct {
	blkTx   map[string]*hashMap // 按block hash 记录交易
	blkTime map[string]int64    // 记录区块的时间，用于清理区块
	smu     sync.RWMutex
}

func (b *blockTx) Lock() {
	b.smu.Lock()
}
func (b *blockTx) UnLock() {
	b.smu.Unlock()
}

func (b *blockTx) GetListTime() map[string]int64 {
	b.smu.Lock()
	defer b.smu.Unlock()

	return b.blkTime

}
func (b *blockTx) Add(bl *block.Block) {
	b.smu.Lock()
	defer b.smu.Unlock()

	blochHash := bl.HashID()

	if _, e := b.blkTx[blochHash]; !e {
		b.blkTx[blochHash] = &hashMap{hashList: make(map[string]struct{})}
	}

	txList := b.blkTx[blochHash]
	for _, tr := range bl.Content {
		txList.Add(tr.TxID())
	}

	slot := consensus_common.Timestamp{Slot: bl.Head.Time}
	b.blkTime[blochHash] = slot.ToUnixSec()

}

func (b *blockTx) Len() int {
	b.smu.Lock()
	defer b.smu.Unlock()

	return len(b.blkTx)
}

func (b *blockTx) Exist(hash string) bool {
	b.smu.Lock()
	defer b.smu.Unlock()

	if _, b := b.blkTx[hash]; b {
		return true
	}

	return false
}

func (b *blockTx) TxList(blockHash string) *hashMap {
	b.smu.Lock()
	defer b.smu.Unlock()

	return b.blkTx[blockHash]
}

func (b *blockTx) Time(hash string) int64 {
	b.smu.Lock()
	defer b.smu.Unlock()

	return b.blkTime[hash]
}

func (b blockTx) Del(hash string) {
	b.smu.Lock()
	defer b.smu.Unlock()

	blk := b.blkTx[hash]
	blk.Clear()
	delete(b.blkTime, hash)
	delete(b.blkTx, hash)
}

type listTx struct {
	list map[string]*tx.Tx
	smu  sync.RWMutex
}

func (l *listTx) Lock() {
	l.smu.Lock()
}
func (l *listTx) UnLock() {
	l.smu.Unlock()
}

func (l *listTx) GetList() map[string]*tx.Tx {
	l.smu.Lock()
	defer l.smu.Unlock()

	return l.list
}

func (l *listTx) Add(Tx *tx.Tx) {
	l.smu.Lock()
	defer l.smu.Unlock()

	if _, ok := l.list[Tx.TxID()]; ok {
		return
	}

	l.list[Tx.TxID()] = &tx.Tx{
		Time:      Tx.Time,
		Nonce:     Tx.Nonce,
		Contract:  Tx.Contract,
		Signs:     Tx.Signs,
		Publisher: Tx.Publisher,
		Recorder:  Tx.Recorder,
	}
}

func (l listTx) Len() int {

	return len(l.list)
}

func (l listTx) Del(hash string) {
	l.smu.Lock()
	defer l.smu.Unlock()

	delete(l.list, hash)

}

func (l listTx) Exist(hash string) bool {
	l.smu.Lock()
	defer l.smu.Unlock()

	if _, b := l.list[hash]; b {
		return true
	}

	return false
}

func (l listTx) Get(hash string) *tx.Tx {

	l.smu.Lock()
	defer l.smu.Unlock()

	return l.list[hash]
}

func (l listTx) Clear() {

	l.smu.Lock()
	defer l.smu.Unlock()

	l.list = make(map[string]*tx.Tx, 0)
}

type blockHashList struct {
	blockList map[string]struct{}
	smu       sync.RWMutex
}

func (b *blockHashList) Add(hash string) {
	b.smu.Lock()
	defer b.smu.Unlock()

	if _, ok := b.blockList[hash]; ok {
		return
	}
	b.blockList[hash] = struct{}{}
}

func (b *blockHashList) Del(hash string) {
	b.smu.Lock()
	defer b.smu.Unlock()

	if _, ok := b.blockList[hash]; ok {
		delete(b.blockList, hash)
	}
}

func (b *blockHashList) Clear() {
	b.smu.Lock()
	defer b.smu.Unlock()

	b.blockList = nil
	b.blockList = make(map[string]struct{})
}

func (b *blockHashList) GetList() map[string]struct{} {
	b.smu.Lock()
	defer b.smu.Unlock()
	return b.blockList
}
