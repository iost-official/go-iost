package new_txpool

import (
	"fmt"
	"sort"
	"sync"
	"time"

	"github.com/iost-official/Go-IOS-Protocol/consensus/common"
	"github.com/iost-official/Go-IOS-Protocol/core/block"
	"github.com/iost-official/Go-IOS-Protocol/core/message"
	"github.com/iost-official/Go-IOS-Protocol/core/new_blockcache"
	"github.com/iost-official/Go-IOS-Protocol/core/new_tx"
	"github.com/iost-official/Go-IOS-Protocol/log"
	"github.com/iost-official/Go-IOS-Protocol/network"
	"github.com/prometheus/client_golang/prometheus"
)

var (
	clearInterval = 11 * time.Second
	filterTime    = 60
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

type FRet uint

const (
	NotFound FRet = iota
	FoundPending
	FoundChain
)

type RecNode struct {
	LinkedNode *blockcache.BlockCacheNode
	HeadNode   *blockcache.BlockCacheNode
}

type TxPoolImpl struct {
	chTx         chan message.Message
	chLinkedNode chan *RecNode

	chain  blockcache.BlockCache
	router network.Router

	head      *blockcache.BlockCacheNode
	block     *sync.Map
	pendingTx *sync.Map

	longestChainHash *sync.Map

	filterTime int64
	mu         sync.RWMutex
}

func NewTxPoolImpl(chain blockcache.BlockCache, router network.Router) (TxPool, error) {

	p := &TxPoolImpl{
		chain:            chain,
		chLinkedNode:     make(chan *RecNode, 100),
		block:            new(sync.Map),
		pendingTx:        new(sync.Map),
		longestChainHash: new(sync.Map),
		filterTime:       int64(filterTime),
	}
	p.router = router
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

	return p, nil
}

func (pool *TxPoolImpl) Start() {
	log.Log.I("TxPoolImpl Start")
	go pool.loop()
}

func (pool *TxPoolImpl) Stop() {
	log.Log.I("TxPoolImpl Stop")
	close(pool.chTx)
	close(pool.chLinkedNode)
}

func (pool *TxPoolImpl) loop() {

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

func (pool *TxPoolImpl) AddLinkedNode(linkedNode *blockcache.BlockCacheNode, headNode *blockcache.BlockCacheNode) error {

	return nil
}

func (pool *TxPoolImpl) AddTx(tx message.Message) error {
	pool.chTx <- tx
	return nil
}

func (pool *TxPoolImpl) PendingTxs(maxCnt int) (tx.TransactionsList, error) {

	pool.updatePending(maxCnt)

	pool.mu.RLock()
	defer pool.mu.RUnlock()

	var pendingList tx.TransactionsList
	list := pool.pendingTx.GetList()

	for _, tx := range list {
		pendingList = append(pendingList, tx)
	}
	sort.Sort(pendingList)

	return pendingList, nil
}

func (pool *TxPoolImpl) ExistTxs(hash string, chainNode *blockcache.BlockCacheNode) (FRet, error) {
	pool.mu.RLock()
	defer pool.mu.RUnlock()

	return pool.listTx.Exist(hash), nil
}

func (pool *TxPoolImpl) initBlockTx() {
	chain := pool.chain.BlockChain()
	timeNow := time.Now().Unix()

	for i := chain.Length() - 1; i > 0; i-- {
		blk := chain.GetBlockByNumber(i)
		if blk == nil {
			return
		}

		t := pool.slotToSec(blk.Head.Time)
		if timeNow-t >= pool.filterTime {
			pool.block.Add(blk)
			pool.longestChainHash.Add(blk.HashID())
		}
	}

}

func (pool *TxPoolImpl) slotToSec(t int64) int64 {
	slot := consensus_common.Timestamp{Slot: t}
	return slot.ToUnixSec()
}

func (pool *TxPoolImpl) addBlock(linkedNode *blockcache.BlockCacheNode) error {

	if _, ok := pool.block.Load(linkedNode.Block.HeadHash()); ok {
		return nil
	}

	b := new(blockTx)

	b.setTime(linkedNode.Block.Head.Time)
	b.addBlock(linkedNode.Block)

	pool.block.Store(linkedNode.Block.HeadHash(), b)

	return nil
}

func (pool *TxPoolImpl) existTxInLongestChain(txHash []byte) bool {

	h := pool.head.Block.Head.Hash()

	ret := pool.existTxInBlock(txHash, h)
	if ret {
		return ret
	}

	h = h.ParentHash

	return ret
}

func (pool *TxPoolImpl) existTxInBlock(txHash []byte, blockHash []byte) bool {

	b, ok := pool.block.Load(blockHash)
	if !ok {
		return false
	}

	return b.(*blockTx).existTx(txHash)
}

func (pool *TxPoolImpl) clearBlock() {
	ft := pool.chain.LinkedTree.Block.Head.Time - (pool.filterTime + pool.filterTime/2)

	pool.block.Range(func(key, value interface{}) bool {
		if value.(*blockTx).time() < ft {
			pool.block.Delete(key)
		}

		return true
	})

}

func (pool *TxPoolImpl) addListTx(tx *tx.Tx) {
	pool.mu.Lock()
	defer pool.mu.Unlock()

	if !pool.listTx.Exist(tx.TxID()) {
		pool.listTx.Add(tx)
	}

}

func (pool *TxPoolImpl) txTimeOut(tx *tx.Tx) bool {

	nTime := time.Now().Unix()
	txTime := tx.Time / 1e9

	if nTime-txTime > pool.filterTime {
		return true
	}
	return false
}

func (pool *TxPoolImpl) delTimeOutTx() {
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

func (pool *TxPoolImpl) delTimeOutBlockTx() {
	pool.mu.Lock()
	defer pool.mu.Unlock()

	nTime := time.Now().Unix()
	chain := pool.chain.BlockChain()
	blk := chain.Top()

	var confirmTime int64
	if blk != nil {
		confirmTime = pool.slotToSec(blk.Head.Time)
	}

	list := pool.block.GetListTime()

	hashList := make([]string, 0)
	for hash, t := range list {

		if t < confirmTime && nTime-pool.filterTime > t {
			hashList = append(hashList, hash)
		}
	}
	for _, hash := range hashList {

		pool.block.Del(hash)
		pool.longestChainHash.Del(hash)
	}
}

func (pool *TxPoolImpl) updatePending(maxCnt int) {
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

func (pool *TxPoolImpl) txExistTxPool(hash string) bool {
	for blockHash := range pool.longestChainHash.GetList() {
		txList := pool.block.TxList(string(blockHash))
		if txList != nil {
			if b := txList.Exist(hash); b {
				return true
			}
		}
	}
	return false
}

func (pool *TxPoolImpl) blockHash(chain block.Chain) *blockHashList {
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

func (pool *TxPoolImpl) updateBlockHash(bhl *blockHashList) {
	pool.mu.RLock()
	defer pool.mu.RUnlock()

	for hash := range bhl.GetList() {
		pool.longestChainHash.Add(hash)
	}
}

type hashMap struct {
	hashList map[string]struct{}
	time     int64
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
	txMap sync.Map
	cTime int64
}

func (b *blockTx) time() int64 {
	return b.cTime
}

func (b *blockTx) setTime(t int64) {
	b.cTime = t
}

func (b *blockTx) addBlock(ib *block.Block) {

	for _, v := range ib.Content {

		b.txMap.Store(v.Hash(), nil)
	}
}

func (b *blockTx) existTx(hash []byte) bool {

	_, r := b.txMap.Load(hash)

	return r
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
