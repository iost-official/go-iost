package pob

import "BlockChainFramework/src/iosbase"

type Leader struct {
	self *Client

	cachedStatePool iosbase.StatePool

	txs []iosbase.Transaction
}

func (l *Leader) Prepare () {
	//copy(l.cachedStatePool, l.statePool)
}

func (l *Leader) MakeBlock() (blk iosbase.Block) {
	return
}

func (l *Leader) OnReceiveTxs(transaction iosbase.Transaction) error {
	if isOK, err := transaction.Verify(l.self.statePool); err == nil && isOK == true {
		l.cachedStatePool.Transact(transaction)
		return nil
	} else {
		return err
	}
}


