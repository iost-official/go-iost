package txpool

import (
	"fmt"
	"sort"
	"sync"
	"time"

	"github.com/iost-official/prototype/consensus/common"
	"github.com/iost-official/prototype/core/block"
	"github.com/iost-official/prototype/core/blockcache"
	"github.com/iost-official/prototype/core/message"
	"github.com/iost-official/prototype/core/tx"
	"github.com/iost-official/prototype/log"
	"github.com/iost-official/prototype/network"
	"github.com/prometheus/client_golang/prometheus"
)

var (
	clearInterval = 11 * time.Second
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
	chTx           chan message.Message
	chConfirmBlock chan *block.Block

	chain  blockcache.BlockCache
	router network.Router

	blockTx   blockTx
	listTx    listTx
	pendingTx listTx

	checkIterateBlockHash blockHashList

	filterTime int64
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
			err := tx.Decode(tr.Body)
			if err != nil {
				continue
			}

			if pool.txTimeOut(&tx) {
				continue
			}

			if blockcache.VerifyTxSig(tx) {
				pool.addListTx(&tx)
				receivedTransactionCount.Inc()
			}

		case bl, ok := <-pool.chConfirmBlock:
			if !ok {
				return
			}
			pool.addBlockTx(bl)
			bhl := pool.blockHash(pool.chain.LongestChain())
			pool.updateBlockHash(bhl)
		case <-clearTx.C:
			pool.delTimeOutTx()
			pool.delTimeOutBlockTx()
		}
	}
}

func (pool *TxPoolServer) AddTransaction(tx *message.Message) {
	pool.chTx <- *tx
}

func (pool *TxPoolServer) PendingTransactions(maxCnt int) tx.TransactionsList {

	pool.updatePending(maxCnt)

	pool.mu.RLock()
	defer pool.mu.RUnlock()

	var pendingList tx.TransactionsList
	list := pool.pendingTx.GetList()

	for _, tx := range list {
		pendingList = append(pendingList, tx)
	}
	sort.Sort(pendingList)

	return pendingList
}

func (pool *TxPoolServer) Transaction(hash string) *tx.Tx {
	pool.mu.RLock()
	defer pool.mu.RUnlock()

	return pool.listTx.Get(hash)
}

func (pool *TxPoolServer) ExistTransaction(hash string) bool {
	pool.mu.RLock()
	defer pool.mu.RUnlock()

	return pool.listTx.Exist(hash)
}

func (pool *TxPoolServer) TransactionNum() int {
	pool.mu.RLock()
	defer pool.mu.RUnlock()

	return pool.listTx.Len()
}

func (pool *TxPoolServer) PendingTransactionNum() int {
	pool.mu.RLock()
	defer pool.mu.RUnlock()

	return pool.pendingTx.Len()
}

func (pool *TxPoolServer) BlockTxNum() int {
	pool.mu.RLock()
	defer pool.mu.RUnlock()

	return pool.blockTx.Len()
}

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

func (pool *TxPoolServer) delTimeOutTx() {
	pool.mu.Lock()
	defer pool.mu.Unlock()

	nTime := time.Now().Unix()
	hashList := make([]string, 0)

	list := pool.listTx.GetList()
	for hash, tx := range list {
		txTime := tx.Time / 1e9
		if nTime-txTime > pool.filterTime {
			hashList = append(hashList, hash)
		}
	}
	for _, hash := range hashList {
		pool.listTx.Del(hash)
	}

}

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

	list := pool.blockTx.GetListTime()

	hashList := make([]string, 0)
	for hash, t := range list {

		if t < confirmTime && nTime-pool.filterTime > t {
			hashList = append(hashList, hash)
		}
	}
	for _, hash := range hashList {

		pool.blockTx.Del(hash)
		pool.checkIterateBlockHash.Del(hash)
	}
}

func (pool *TxPoolServer) updatePending(maxCnt int) {
	pool.mu.Lock()
	defer pool.mu.Unlock()

	pool.pendingTx.list = make(map[string]*tx.Tx, 0)

	list := pool.listTx.GetList()
	for hash, tr := range list {
		if !pool.txExistTxPool(hash) {
			pool.pendingTx.Add(tr)
			if pool.pendingTx.Len() >= maxCnt {
				break
			}
		}
	}
}

func (pool *TxPoolServer) txExistTxPool(hash string) bool {
	for blockHash := range pool.checkIterateBlockHash.GetList() {
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
	pool.mu.RLock()
	defer pool.mu.RUnlock()

	bhl := &blockHashList{blockList: make(map[string]struct{}, 0)}
	iter := chain.Iterator()
	for {
		blk := iter.Next()
		if blk == nil {
			break
		}
		bhl.Add(blk.HashID())
	}
	return bhl
}

func (pool *TxPoolServer) updateBlockHash(bhl *blockHashList) {
	pool.mu.RLock()
	defer pool.mu.RUnlock()

	for hash := range bhl.GetList() {
		pool.checkIterateBlockHash.Add(hash)
	}
}

func (pool *TxPoolServer) addBlockTx(bl *block.Block) {
	pool.mu.Lock()
	defer pool.mu.Unlock()

	if !pool.blockTx.Exist(bl.HashID()) {
		pool.blockTx.Add(bl)
	}
}

type hashMap struct {
	hashList map[string]struct{}
}

func (h *hashMap) Add(hash string) {
	h.hashList[hash] = struct{}{}
}

func (h *hashMap) Exist(txHash string) bool {
	if _, b := h.hashList[txHash]; b {
		return true
	}

	return false
}

func (h *hashMap) Del(hash string) {
	delete(h.hashList, hash)
}

func (h *hashMap) Clear() {
	h.hashList = nil
	h.hashList = make(map[string]struct{})
}

type blockTx struct {
	blkTx   map[string]*hashMap
	blkTime map[string]int64
}

func (b *blockTx) GetListTime() map[string]int64 {
	return b.blkTime
}

func (b *blockTx) Add(bl *block.Block) {
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
	return len(b.blkTx)
}

func (b *blockTx) Exist(hash string) bool {
	if _, b := b.blkTx[hash]; b {
		return true
	}
	return false
}

func (b *blockTx) TxList(blockHash string) *hashMap {
	return b.blkTx[blockHash]
}

func (b *blockTx) Time(hash string) int64 {
	return b.blkTime[hash]
}

func (b blockTx) Del(hash string) {
	blk := b.blkTx[hash]
	blk.Clear()
	delete(b.blkTime, hash)
	delete(b.blkTx, hash)
}

type listTx struct {
	list map[string]*tx.Tx
}

func (l *listTx) GetList() map[string]*tx.Tx {
	return l.list
}

func (l *listTx) Add(Tx *tx.Tx) {
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
	delete(l.list, hash)

}

func (l listTx) Exist(hash string) bool {
	if _, b := l.list[hash]; b {
		return true
	}

	return false
}

func (l listTx) Get(hash string) *tx.Tx {
	return l.list[hash]
}

func (l listTx) Clear() {
	l.list = make(map[string]*tx.Tx, 0)
}

type blockHashList struct {
	blockList map[string]struct{}
}

func (b *blockHashList) Add(hash string) {
	if _, ok := b.blockList[hash]; ok {
		return
	}
	b.blockList[hash] = struct{}{}
}

func (b *blockHashList) Del(hash string) {
	if _, ok := b.blockList[hash]; ok {
		delete(b.blockList, hash)
	}
}

func (b *blockHashList) Clear() {
	b.blockList = nil
	b.blockList = make(map[string]struct{})
}

func (b *blockHashList) GetList() map[string]struct{} {
	return b.blockList
}
