package txpool

import (
	"errors"
	"github.com/iost-official/go-iost/core/blockcache"
	"github.com/iost-official/go-iost/core/tx"
	"time"
)

// TxMigrater when blockchain changes, TxMigrater can be used to indicate how to change txpool
type TxMigrater struct {
	blockCache        blockcache.BlockCache
	forkChain         *forkChain
	recentBlockBuffer *BlockchainWrapper
	filterTime        int64
}

// NewTxMigrater ...
func NewTxMigrater(blockCache blockcache.BlockCache, recentBlockBuffer *BlockchainWrapper) *TxMigrater {
	tm := &TxMigrater{
		blockCache:        blockCache,
		forkChain:         new(forkChain),
		recentBlockBuffer: recentBlockBuffer,
		filterTime:        filterTime,
	}
	tm.forkChain.SetNewHead(blockCache.Head())
	return tm
}

func (x *TxMigrater) getUpdateTxs(linkedNode *blockcache.BlockCacheNode) (txsToAdd []*tx.Tx, txsToDel []*tx.Tx, err error) {
	var newHead *blockcache.BlockCacheNode
	h := x.blockCache.Head()
	if linkedNode.Head.Number > h.Head.Number {
		newHead = linkedNode
	} else {
		newHead = h
	}

	typeOfFork := x.updateForkChain(newHead)
	switch typeOfFork {
	case forkBCN:
		txsToAdd, txsToDel = x.doChainChangeByForkBCN()
		err = nil
		return
	case noForkBCN:
		txsToAdd, txsToDel = x.doChainChangeByTimeout()
		err = nil
		return
	case sameHead:
		txsToAdd, txsToDel = make([]*tx.Tx, 0), make([]*tx.Tx, 0)
		err = nil
		return
	default:
		txsToAdd, txsToDel = make([]*tx.Tx, 0), make([]*tx.Tx, 0)
		err = errors.New("failed to get update txs")
		return
	}
}

func (x *TxMigrater) updateForkChain(newHead *blockcache.BlockCacheNode) tFork {
	if x.forkChain.GetNewHead() == newHead {
		return sameHead
	}
	x.forkChain.SetOldHead(x.forkChain.GetNewHead())
	x.forkChain.SetNewHead(newHead)
	bcn, ok := x.findForkBCN(x.forkChain.GetNewHead(), x.forkChain.GetOldHead())
	if ok {
		x.forkChain.SetForkBCN(bcn)
		return forkBCN
	}
	x.forkChain.SetForkBCN(nil)
	return noForkBCN
}

func (x *TxMigrater) findForkBCN(newHead *blockcache.BlockCacheNode, oldHead *blockcache.BlockCacheNode) (*blockcache.BlockCacheNode, bool) {
	for {
		for oldHead != nil && oldHead.Head.Number > newHead.Head.Number {
			oldHead = oldHead.GetParent()
		}
		if oldHead == nil {
			return nil, false
		}
		if oldHead == newHead {
			return oldHead, true
		}
		newHead = newHead.GetParent()
		if newHead == nil {
			return nil, false
		}
	}
}

func (x *TxMigrater) doChainChangeByForkBCN() (txsToAdd []*tx.Tx, txsToDel []*tx.Tx) {
	newHead := x.forkChain.GetNewHead()
	oldHead := x.forkChain.GetOldHead()
	forkBCN := x.forkChain.GetForkBCN()
	//add txs
	filterLimit := time.Now().UnixNano() - x.filterTime
	txsToAdd = make([]*tx.Tx, 0)
	txsToDel = make([]*tx.Tx, 0)
	for {
		if oldHead == nil || oldHead == forkBCN || oldHead.Block.Head.Time < filterLimit {
			break
		}
		txsToAdd = append(txsToAdd, oldHead.Block.Txs...)
		oldHead = oldHead.GetParent()
	}

	//del txs
	for {
		if newHead == nil || newHead == forkBCN || newHead.Block.Head.Time < filterLimit {
			break
		}
		txsToDel = append(txsToDel, newHead.Block.Txs...)
		newHead = newHead.GetParent()
	}
	return
}

func (x *TxMigrater) doChainChangeByTimeout() (txsToAdd []*tx.Tx, txsToDel []*tx.Tx) {
	newHead := x.forkChain.GetNewHead()
	oldHead := x.forkChain.GetOldHead()
	filterLimit := time.Now().UnixNano() - x.filterTime
	txsToAdd = make([]*tx.Tx, 0)
	txsToDel = make([]*tx.Tx, 0)
	ob, ok := x.recentBlockBuffer.findBlock(oldHead.Block.HeadHash())
	if ok {
		for {
			if ob.time < filterLimit {
				break
			}
			ob.txMap.Range(func(k, v interface{}) bool {
				txsToAdd = append(txsToAdd, v.(*tx.Tx))
				return true
			})
			ob, ok = x.recentBlockBuffer.findBlock(ob.ParentHash)
			if !ok {
				break
			}
		}
	}
	nb, ok := x.recentBlockBuffer.findBlock(newHead.Block.HeadHash())
	if ok {
		for {
			if nb.time < filterLimit {
				break
			}
			nb.txMap.Range(func(k, v interface{}) bool {
				txsToDel = append(txsToDel, v.(*tx.Tx))
				return true
			})
			nb, ok = x.recentBlockBuffer.findBlock(nb.ParentHash)
			if !ok {
				break
			}
		}
	}
	return
}
