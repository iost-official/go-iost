package txpool

import (
	"bytes"
	"fmt"
	"math"
	"sync"
	"time"

	"github.com/iost-official/go-iost/v3/core/block"
	"github.com/iost-official/go-iost/v3/core/tx"
	"github.com/iost-official/go-iost/v3/ilog"

	"github.com/emirpasic/gods/trees/redblacktree"
	"go.uber.org/atomic"
)

var (
	minTickerTime = time.Second
)

func compareDeferTx(a, b any) int {
	txa := a.(*tx.Tx)
	txb := b.(*tx.Tx)
	if txa.Time == txb.Time {
		return bytes.Compare(txa.Hash(), txb.Hash())
	}
	return int(txa.Time - txb.Time)
}

// DeferServer manages defer transaction index and sends them to txpool on time.
type DeferServer struct {
	pool             *redblacktree.Tree
	idxMap           map[string]*tx.Tx
	rw               *sync.RWMutex
	nextScheduleTime atomic.Int64

	bChain block.Chain
	txpool TxPool

	quitCh chan struct{}
}

// NewDeferServer returns a new DeferServer instance.
func NewDeferServer(txpool TxPool, bChain block.Chain) (*DeferServer, error) {
	deferServer := &DeferServer{
		pool:   redblacktree.NewWith(compareDeferTx),
		idxMap: make(map[string]*tx.Tx),
		rw:     new(sync.RWMutex),
		bChain: bChain,
		txpool: txpool,
		quitCh: make(chan struct{}),
	}
	err := deferServer.buildIndex()
	if err != nil {
		return nil, fmt.Errorf("build defertx index error, %v", err)
	}

	return deferServer, nil
}

func (d *DeferServer) buildIndex() error {
	txs, err := d.bChain.AllDelaytx()
	ilog.Info("defer index num: ", len(txs))
	if err != nil {
		return err
	}
	for _, t := range txs {
		idx := d.toIndex(t)
		d.pool.Put(idx, true)
		d.idxMap[string(idx.ReferredTx)] = idx
	}
	return nil
}

func (d *DeferServer) toIndex(delayTx *tx.Tx) *tx.Tx {
	return &tx.Tx{
		ReferredTx: delayTx.Hash(),
		Time:       delayTx.Time + delayTx.Delay,
	}
}

func (d *DeferServer) delDeferIndex(idx *tx.Tx) {
	d.rw.Lock()
	d.pool.Remove(idx)
	delete(d.idxMap, string(idx.ReferredTx))
	d.rw.Unlock()
}

// DelDeferTx deletes a tx in defer server.
func (d *DeferServer) DelDeferTx(deferTx *tx.Tx) error {
	idx := &tx.Tx{
		ReferredTx: deferTx.ReferredTx,
		Time:       deferTx.Time,
	}
	d.delDeferIndex(idx)
	return nil
}

// DelDeferTxByHash deletes a tx in defer server by referredTx hash.
func (d *DeferServer) DelDeferTxByHash(txHash []byte) {
	hashString := string(txHash)

	d.rw.Lock()
	defer d.rw.Unlock()
	idx := d.idxMap[hashString]
	if idx != nil {
		d.pool.Remove(idx)
		delete(d.idxMap, hashString)
	}
}

// StoreDeferTx stores a tx in defer server.
func (d *DeferServer) StoreDeferTx(delayTx *tx.Tx) {
	idx := d.toIndex(delayTx)
	d.rw.Lock()
	d.pool.Put(idx, true)
	d.idxMap[string(idx.ReferredTx)] = idx
	d.rw.Unlock()
	if idx.Time < d.nextScheduleTime.Load() {
		d.nextScheduleTime.Store(idx.Time)
		d.restartDeferTicker()
	}
}

// DumpDeferTx dumps all defer transactions for debug.
func (d *DeferServer) DumpDeferTx() []*tx.Tx {
	ret := make([]*tx.Tx, 0)
	iter := d.pool.Iterator()
	d.rw.RLock()
	ok := iter.Next()
	for ok {
		deferTx := iter.Key().(*tx.Tx)
		ret = append(ret, deferTx)
		ok = iter.Next()
	}
	d.rw.RUnlock()
	return ret
}

// Start starts the defer server.
func (d *DeferServer) Start() error {
	go d.deferTicker()
	return nil
}

// Stop stops the defer server.
func (d *DeferServer) Stop() {
	d.stopDeferTicker()
}

func (d *DeferServer) stopDeferTicker() {
	d.quitCh <- struct{}{}
	<-d.quitCh
}

func (d *DeferServer) restartDeferTicker() {
	d.stopDeferTicker()
	go d.deferTicker()
}

func (d *DeferServer) deferTicker() {
	for {
		scheduled := time.Duration(d.nextScheduleTime.Load() - time.Now().UnixNano())
		if scheduled < minTickerTime {
			scheduled = minTickerTime
		}
		ilog.Info("next defer schedule: ", scheduled)
		select {
		case <-d.quitCh:
			d.quitCh <- struct{}{}
			return
		case <-time.After(scheduled):
			iter := d.pool.Iterator()
			d.rw.RLock()
			ok := iter.Next()
			d.rw.RUnlock()
			for ok {
				idx := iter.Key().(*tx.Tx)
				if idx.Time > time.Now().UnixNano() {
					d.nextScheduleTime.Store(idx.Time)
					break
				}
				/*
					err := d.txpool.AddDefertx(idx.ReferredTx)
					if err == ErrCacheFull {
						d.nextScheduleTime.Store(idx.Time)
						ilog.Infof("Adding defertx failed, txpool is full, retry after one second.")
						break
					}
					if err == errDelaytxNotFound {
						ilog.Error("Adding defertx failed, delaytx not found, delete defer index")
						d.delDeferIndex(idx)
					}
					if err == nil || err == ErrDupChainTx || err == ErrDupPendingTx {
						d.delDeferIndex(idx)
					}
				*/
				d.rw.RLock()
				ok = iter.Next()
				d.rw.RUnlock()
			}
			if !ok {
				d.nextScheduleTime.Store(math.MaxInt64)
			}
		}
	}
}

// ProcessDelaytx will process the delay tx.
func (d *DeferServer) ProcessDelaytx(blk *block.Block) {
	for i, t := range blk.Txs {
		if t.Delay > 0 && blk.Receipts[i].Status.Code == tx.Success {
			d.StoreDeferTx(t)
		}
		if t.IsDefer() {
			d.DelDeferTx(t)
		}
		canceledDelayHashes := blk.Receipts[i].ParseCancelDelaytx()
		for _, canceledHash := range canceledDelayHashes {
			d.DelDeferTxByHash(canceledHash)
		}
	}
}
