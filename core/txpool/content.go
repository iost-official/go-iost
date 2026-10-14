package txpool

import (
	"bytes"
	"fmt"

	"github.com/iost-official/go-iost/v3/core/block"
	"github.com/iost-official/go-iost/v3/core/tx"
)

// HasContent searches chainBlock and its ancestors back to t.Time.
// It returns an error if required history cannot be read.
func (pool *TxPImpl) HasContent(t *tx.Tx, chainBlock *block.Block) (bool, error) {
	if t == nil {
		return false, fmt.Errorf("missing transaction for content lookup")
	}
	if chainBlock == nil || chainBlock.Head == nil {
		return false, fmt.Errorf("missing parent block for content lookup")
	}
	contentHash := t.ContentHash()
	b, ok := pool.findBlock(chainBlock.HeadHash())
	if !ok {
		b = pool.cacheContentBlock(chainBlock)
	}
	for b.time >= t.Time && b.number > 0 {
		if _, found := b.contentMap[string(contentHash)]; found {
			return true, nil
		}
		parent, err := pool.loadContentBlock(b.ParentHash)
		if err != nil {
			return false, err
		}
		if parent.number != b.number-1 || parent.time >= b.time {
			return false, fmt.Errorf("invalid content history ancestry")
		}
		b = parent
	}
	return false, nil
}

func (pool *TxPImpl) loadContentBlock(hash []byte) (*blockTx, error) {
	if b, ok := pool.findBlock(hash); ok {
		return b, nil
	}
	if pool.blockCache != nil {
		if b, err := pool.blockCache.GetBlockByHash(hash); err == nil {
			if b == nil || b.Head == nil || !bytes.Equal(b.HeadHash(), hash) {
				return nil, fmt.Errorf("invalid cached history block")
			}
			return pool.cacheContentBlock(b), nil
		}
	}
	if pool.bChain == nil {
		return nil, fmt.Errorf("missing history for content lookup")
	}
	b, err := pool.bChain.GetBlockByHash(hash)
	if err != nil {
		return nil, fmt.Errorf("load content history: %w", err)
	}
	if b == nil || b.Head == nil || !bytes.Equal(b.HeadHash(), hash) {
		return nil, fmt.Errorf("invalid persisted history block")
	}
	return pool.cacheContentBlock(b), nil
}

func (pool *TxPImpl) cacheContentBlock(blk *block.Block) *blockTx {
	value, _ := pool.blockList.LoadOrStore(string(blk.HeadHash()), newBlockTx(blk))
	return value.(*blockTx)
}

// removeConfirmed requires pool.mu to be held.
func (pool *TxPImpl) removeConfirmed(t *tx.Tx, isBase bool) {
	if !isBase {
		pool.pendingTx.DelContent(t.ContentHash())
		return
	}
	pool.pendingTx.Del(t.Hash())
}
