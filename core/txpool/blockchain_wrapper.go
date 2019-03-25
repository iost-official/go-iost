package txpool

import (
	"errors"
	"github.com/iost-official/go-iost/core/block"
	"github.com/iost-official/go-iost/core/blockcache"
	"github.com/iost-official/go-iost/core/tx"
	"sync"
	"time"
)

// BlockchainWrapper caches recent blocks(within last 90s) in memory. Is can be used to check whether a tx hash is in recent blocks
// This class should be integrated into blockchain/blockcache class later
type BlockchainWrapper struct {
	blockList  *sync.Map // map[string]*blockTx
	filterTime int64
	head       *block.Block
	blockCache blockcache.BlockCache
}

// NewBlockchainWrapper ...
func NewBlockchainWrapper(bc blockcache.BlockCache) *BlockchainWrapper {
	return &BlockchainWrapper{
		blockList:  new(sync.Map),
		filterTime: filterTime,
		head:       bc.Head().Block,
		blockCache: bc,
	}
}

func (b *BlockchainWrapper) init(chain block.Chain) {
	filterLimit := time.Now().UnixNano() - b.filterTime
	for i := chain.Length() - 1; i > 0; i-- {
		blk, err := chain.GetBlockByNumber(i)
		if err != nil {
			break
		}
		if blk.Head.Time < filterLimit {
			break
		}
		b.addBlock(blk)
	}
}

func (b *BlockchainWrapper) addBlock(blk *block.Block) error {
	if blk == nil {
		return errors.New("failed to linkedBlock")
	}
	b.blockList.LoadOrStore(string(blk.HeadHash()), newBlockTx(blk))
	if b.head == nil || b.head.Head.Number < blk.Head.Number {
		b.head = blk
	}
	return nil
}

func (b *BlockchainWrapper) parentHash(hash []byte) ([]byte, bool) {
	v, ok := b.findBlock(hash)
	if !ok {
		return nil, false
	}
	return v.ParentHash, true
}

func (b *BlockchainWrapper) findBlock(hash []byte) (*blockTx, bool) {
	if v, ok := b.blockList.Load(string(hash)); ok {
		return v.(*blockTx), true
	}
	return nil, false
}

func (b *BlockchainWrapper) getTxAndReceiptInChain(txHash []byte, blk *block.Block) (*tx.Tx, *tx.TxReceipt) {
	if blk == nil {
		blk = b.head
	}
	blkHash := blk.HeadHash()
	filterLimit := blk.Head.Time - b.filterTime
	var ok bool
	for {
		t, tr := b.getTxAndReceiptInBlock(txHash, blkHash)
		if t != nil {
			return t, tr
		}
		blkHash, ok = b.parentHash(blkHash)
		if !ok {
			return nil, nil
		}
		if b, ok := b.findBlock(blkHash); ok {
			if b.time < filterLimit {
				return nil, nil
			}
		}
	}
}

func (b *BlockchainWrapper) setHead(blk *block.Block) {
	b.head = blk
}

func (b *BlockchainWrapper) existTxInChain(txHash []byte, blk *block.Block) bool {
	t, _ := b.getTxAndReceiptInChain(txHash, blk)
	return t != nil
}

func (b *BlockchainWrapper) getTxAndReceiptInBlock(txHash []byte, blockHash []byte) (*tx.Tx, *tx.TxReceipt) {
	blk, ok := b.blockList.Load(string(blockHash))
	if !ok {
		return nil, nil
	}
	return blk.(*blockTx).getTxAndReceipt(txHash)
}

func (b *BlockchainWrapper) clearBlock() {
	head := b.blockCache.LinkedRoot().Block
	headTime := head.Head.Time
	filterLimit := headTime - b.filterTime
	b.blockList.Range(func(key, value interface{}) bool {
		if value.(*blockTx).time < filterLimit {
			b.blockList.Delete(key)
		}
		return true
	})
}

// getFromChain gets transaction from longest chain.
func (b *BlockchainWrapper) getFromChain(hash []byte) (*tx.Tx, *tx.TxReceipt, error) {
	t, tr := b.getTxAndReceiptInChain(hash, nil)
	if t == nil {
		return nil, nil, ErrTxNotFound
	}
	return t, tr, nil
}
