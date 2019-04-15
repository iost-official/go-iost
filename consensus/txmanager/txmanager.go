package txmanager

import (
	"sync"

	"github.com/iost-official/go-iost/core/tx"
	"github.com/iost-official/go-iost/core/txpool"
	"github.com/iost-official/go-iost/ilog"
	"github.com/iost-official/go-iost/p2p"
)

// TxManager will maintain the tx received from p2p.
type TxManager struct {
	p      p2p.Service
	txPool txpool.TxPool

	msgCh chan p2p.IncomingMessage

	quitCh chan struct{}
	done   *sync.WaitGroup
}

// New will return a TxManager.
func New(p p2p.Service, txPool txpool.TxPool) *TxManager {
	t := &TxManager{
		p:      p,
		txPool: txPool,

		msgCh: p.Register("tx from other nodes", p2p.PublishTx),

		quitCh: make(chan struct{}),
		done:   new(sync.WaitGroup),
	}

	t.done.Add(1)
	go t.receiveP2PTxController()

	return t
}

// Close will close the TxManager.
func (t *TxManager) Close() {
	close(t.quitCh)
	t.done.Wait()
	ilog.Infof("Stopped tx filter.")
}

func (t *TxManager) handleTx(msg *p2p.IncomingMessage) {
	transaction := &tx.Tx{}
	err := transaction.Decode(msg.Data())
	if err != nil {
		ilog.Errorf("decode tx error. err=%v", err)
		return
	}
	if err := t.txPool.AddTx(transaction, "p2p"); err != nil {
		ilog.Debugf("Add tx failed: %v", err)
		return
	}
}

func (t *TxManager) receiveP2PTxController() {
	for {
		select {
		case msg := <-t.msgCh:
			t.handleTx(&msg)
		case <-t.quitCh:
			t.done.Done()
			return
		}
	}
}
