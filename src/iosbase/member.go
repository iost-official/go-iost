package iosbase

import (
	"fmt"
)

type Member struct {
	ID     string
	Pubkey []byte
	Seckey []byte
}

type Replica struct {
	TxCache         []Transaction
	BlockChain      BlockChain
	StatePool       StatePool
	CachedStatePool StatePool
}

func (r *Replica) OnReceiveTxs(txs []Transaction) error {
	for _, tx := range txs {
		// TODO : check if tx existed
		if ok, err := tx.Verify(&r.CachedStatePool); err == nil && ok {
			r.CachedStatePool.Transact(tx)
			r.TxCache = append(r.TxCache, tx)
		} else if err != nil {
			return err
		} else {
			return fmt.Errorf("rejected")
		}
	}
	return nil
}

func (r *Replica) OnPullRequest(blkID int) (Block, error) {
	return nil, nil
}


