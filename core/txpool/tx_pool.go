package txpool

import (
	"errors"
	"fmt"
	"runtime"
	"sync"
	"time"

	"github.com/iost-official/go-iost/common"
	"github.com/iost-official/go-iost/core/block"
	"github.com/iost-official/go-iost/core/blockcache"
	"github.com/iost-official/go-iost/core/global"
	"github.com/iost-official/go-iost/core/tx"
	"github.com/iost-official/go-iost/ilog"
	"github.com/iost-official/go-iost/p2p"
)

var errDelaytxNotFound = errors.New("delay tx not found")

// TxPImpl defines all the API of txpool package.
type TxPImpl struct {
	global            global.BaseVariable
	txMigrater        *TxMigrater
	blockchainWrapper *BlockchainWrapper
	p2pService        p2p.Service
	pendingTx         *SortedTxMap
	mu                sync.RWMutex
	chP2PTx           chan p2p.IncomingMessage
	deferServer       *DeferServer
	quitGenerateMode  chan struct{}
	quitCh            chan struct{}
}

// NewTxPoolImpl returns a default TxPImpl instance.
func NewTxPoolImpl(global global.BaseVariable, blockCache blockcache.BlockCache, p2pService p2p.Service) (*TxPImpl, error) {
	p := &TxPImpl{
		global:            global,
		blockchainWrapper: NewBlockchainWrapper(blockCache),
		p2pService:        p2pService,
		pendingTx:         NewSortedTxMap(),
		chP2PTx:           p2pService.Register("txpool message", p2p.PublishTx),
		quitGenerateMode:  make(chan struct{}),
		quitCh:            make(chan struct{}),
	}
	p.txMigrater = NewTxMigrater(blockCache, p.blockchainWrapper)
	deferServer, err := NewDeferServer(p)
	if err != nil {
		return nil, err
	}
	p.deferServer = deferServer
	close(p.quitGenerateMode)
	return p, nil
}

// Start starts the jobs.
func (pool *TxPImpl) Start() error {
	go pool.deferServer.Start()
	go pool.loop()
	return nil
}

// Stop stops all the jobs.
func (pool *TxPImpl) Stop() {
	pool.deferServer.Stop()
	close(pool.quitCh)
}

// AddDefertx adds defer transaction.
func (pool *TxPImpl) AddDefertx(txHash []byte) error {
	if pool.pendingTx.Size() > maxCacheTxs {
		return ErrCacheFull
	}
	referredTx, err := pool.global.BlockChain().GetTx(txHash)
	if err != nil {
		referredTx, _, err = pool.blockchainWrapper.getFromChain(txHash)
		if err != nil {
			return errDelaytxNotFound
		}
	}
	deferTx := referredTx.DeferTx()
	err = pool.verifyDuplicate(deferTx)
	if err != nil {
		return err
	}
	err = deferTx.VerifySelf()
	if err != nil {
		return err
	}
	pool.pendingTx.Add(deferTx)
	return nil
}

func (pool *TxPImpl) loop() {
	for {
		if pool.global.Mode() != global.ModeInit {
			break
		}
		time.Sleep(time.Second)
	}
	pool.initBlockTx()
	workerCnt := (runtime.NumCPU() + 1) / 2
	if workerCnt == 0 {
		workerCnt = 1
	}
	for i := 0; i < workerCnt; i++ {
		go pool.verifyWorkers()
	}
	clearTx := time.NewTicker(clearInterval)
	defer clearTx.Stop()
	for {
		select {
		case <-clearTx.C:
			pool.mu.Lock()
			pool.blockchainWrapper.clearBlock()
			pool.clearTimeoutTx()
			pool.mu.Unlock()
			metricsTxPoolSize.Set(float64(pool.pendingTx.Size()), nil)
		case <-pool.quitCh:
			return
		}
	}
}

// Lock lock the txpool
func (pool *TxPImpl) Lock() {
	pool.mu.Lock()
	pool.quitGenerateMode = make(chan struct{})
}

// PendingTx is return pendingTx
func (pool *TxPImpl) PendingTx() (*SortedTxMap, *blockcache.BlockCacheNode) {
	return pool.pendingTx, nil // nil will be removed later
}

// Release release the txpool
func (pool *TxPImpl) Release() {
	close(pool.quitGenerateMode)
	pool.mu.Unlock()
}

func (pool *TxPImpl) verifyWorkers() {
	for v := range pool.chP2PTx {
		select {
		case <-pool.quitGenerateMode:
		}
		var t tx.Tx
		err := t.Decode(v.Data())
		if err != nil {
			ilog.Errorf("decode tx error. err=%v", err)
			continue
		}
		pool.mu.Lock()
		ret := pool.verifyDuplicate(&t)
		if ret != nil {
			pool.mu.Unlock()
			continue
		}
		ret = pool.verifyTx(&t)
		if ret != nil {
			pool.mu.Unlock()
			continue
		}
		pool.pendingTx.Add(&t)
		pool.mu.Unlock()
		metricsReceivedTxCount.Add(1, map[string]string{"from": "p2p"})
		pool.p2pService.Broadcast(v.Data(), p2p.PublishTx, p2p.NormalMessage)
	}
}

func (pool *TxPImpl) processDelaytx(blk *block.Block) {
	for i, t := range blk.Txs {
		if t.Delay > 0 && blk.Receipts[i].Status.Code == tx.Success {
			pool.deferServer.StoreDeferTx(t)
		}
		if t.IsDefer() {
			pool.deferServer.DelDeferTx(t)
		}
		canceledDelayHashes := blk.Receipts[i].ParseCancelDelaytx()
		for _, canceledHash := range canceledDelayHashes {
			pool.deferServer.DelDeferTxByHash(canceledHash)
		}
	}
}

// AddLinkedNode add the findBlock
func (pool *TxPImpl) AddLinkedNode(linkedNode *blockcache.BlockCacheNode) error {
	pool.mu.Lock()
	defer pool.mu.Unlock()
	pool.processDelaytx(linkedNode.Block)
	err := pool.blockchainWrapper.addBlock(linkedNode.Block)
	if err != nil {
		return fmt.Errorf("failed to add findBlock: %v", err)
	}

	txsToAdd, txsToDel, err := pool.txMigrater.getUpdateTxs(linkedNode)
	if err != nil {
		return nil
	}
	for _, t := range txsToAdd {
		pool.pendingTx.Add(t)
	}
	for _, t := range txsToDel {
		pool.pendingTx.Del(t.Hash())
	}

	return nil
}

// AddTx add the transaction
func (pool *TxPImpl) AddTx(t *tx.Tx) error {
	err := pool.verifyDuplicate(t)
	if err != nil {
		return err
	}
	err = pool.verifyTx(t)
	if err != nil {
		return err
	}
	pool.pendingTx.Add(t)
	ilog.Debugf(
		"Added %v to pendingTx, now size is %v.",
		common.Base58Encode(t.Hash()),
		pool.pendingTx.Size(),
	)

	pool.p2pService.Broadcast(t.Encode(), p2p.PublishTx, p2p.NormalMessage)
	metricsReceivedTxCount.Add(1, map[string]string{"from": "rpc"})
	return nil
}

// DelTx del the transaction
func (pool *TxPImpl) DelTx(hash []byte) error {
	pool.pendingTx.Del(hash)
	return nil
}

// DelTxList deletes the tx list in txpool.
func (pool *TxPImpl) DelTxList(delList []*tx.Tx) {
	for _, t := range delList {
		pool.pendingTx.Del(t.Hash())
	}
}

// ExistTxs determine if the transaction exists
func (pool *TxPImpl) ExistTxs(hash []byte, chainBlock *block.Block) FRet {
	var r FRet
	switch {
	case pool.existTxInPending(hash):
		r = FoundPending
	case pool.blockchainWrapper.existTxInChain(hash, chainBlock):
		r = FoundChain
	default:
		r = NotFound
	}
	return r
}

func (pool *TxPImpl) initBlockTx() {
	pool.blockchainWrapper.init(pool.global.BlockChain())
}

func (pool *TxPImpl) verifyTx(t *tx.Tx) error {
	if pool.pendingTx.Size() > maxCacheTxs {
		return ErrCacheFull
	}
	if t.IsDefer() {
		return errors.New("reject defertx")
	}
	// Add one second delay for tx created time check
	if !t.IsCreatedBefore(time.Now().UnixNano()+maxTxTimeGap) || t.IsExpired(time.Now().UnixNano()) {
		return fmt.Errorf("TimeError")
	}
	if err := t.VerifySelf(); err != nil {
		return fmt.Errorf("VerifyError %v", err)
	}

	return nil
}

func (pool *TxPImpl) verifyDuplicate(t *tx.Tx) error {
	if pool.existTxInPending(t.Hash()) {
		return ErrDupPendingTx
	}
	if pool.blockchainWrapper.existTxInChain(t.Hash(), nil) {
		return ErrDupChainTx
	}
	return nil
}

func (pool *TxPImpl) existTxInPending(hash []byte) bool {
	return pool.pendingTx.Get(hash) != nil
}

func (pool *TxPImpl) clearTimeoutTx() {
	iter := pool.pendingTx.Iter()
	t, ok := iter.Next()
	for ok {
		if t.IsExpired(time.Now().UnixNano()) && !t.IsDefer() {
			pool.pendingTx.Del(t.Hash())
		}
		t, ok = iter.Next()
	}
}

// GetFromPending gets transaction from pending list.
func (pool *TxPImpl) GetFromPending(hash []byte) (*tx.Tx, error) {
	tx := pool.pendingTx.Get(hash)
	if tx == nil {
		return nil, ErrTxNotFound
	}
	return tx, nil
}

// GetFromChain gets transaction from longest chain.
func (pool *TxPImpl) GetFromChain(hash []byte) (*tx.Tx, *tx.TxReceipt, error) {
	return pool.blockchainWrapper.getFromChain(hash)
}
