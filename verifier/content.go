package verifier

import (
	"fmt"

	"github.com/iost-official/go-iost/v3/core/block"
	"github.com/iost-official/go-iost/v3/core/tx"
)

func verifyTxTime(t *tx.Tx, timestamp int64) error {
	if !t.IsCreatedBefore(timestamp) {
		return ErrNotArrivedTx
	}
	if t.IsExpired(timestamp) {
		return ErrExpiredTx
	}
	return nil
}

// VerifyBlockContent checks user transaction times and content hashes.
func VerifyBlockContent(blk, parent *block.Block, hasContent func(*tx.Tx, *block.Block) (bool, error)) error {
	if hasContent == nil {
		return fmt.Errorf("missing transaction content checker")
	}
	seen := make(map[string]struct{}, len(blk.Txs))
	for i, t := range blk.Txs {
		if i == 0 {
			continue
		}
		if err := verifyTxTime(t, blk.Head.Time); err != nil {
			return err
		}
		key := string(t.ContentHash())
		if _, ok := seen[key]; ok {
			return ErrDuplicateContent
		}
		seen[key] = struct{}{}
		found, err := hasContent(t, parent)
		if err != nil {
			return err
		}
		if found {
			return ErrDuplicateContent
		}
	}
	return nil
}
