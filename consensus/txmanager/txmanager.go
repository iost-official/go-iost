package txmanager

import (
	"sync"
	"time"

	"github.com/Jeffail/tunny"
	"github.com/iost-official/go-iost/core/tx"
	"github.com/iost-official/go-iost/core/txpool"
	"github.com/iost-official/go-iost/ilog"
	"github.com/iost-official/go-iost/p2p"
)

var (
	txHandlerPoolSize = 2
	timeout           = 2 * time.Second
)

// TxManager will maintain the tx received from p2p.
type TxManager struct {
	p      p2p.Service
	txPool txpool.TxPool
	pool   *tunny.Pool

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
	t.pool = tunny.NewFunc(txHandlerPoolSize, t.handleTx)

	t.done.Add(1)
	go t.receiveP2PTxController()

	return t
}

// Close will close the TxManager.
func (t *TxManager) Close() {
	close(t.quitCh)
	t.done.Wait()
	ilog.Infof("Stopped tx manager.")
}

func (t *TxManager) handleTx(payload interface{}) interface{} {
	msg, ok := payload.(*p2p.IncomingMessage)
	if !ok {
		ilog.Warnf("Assert payload to IncomingMessage failed")
		return nil
	}

	transaction := &tx.Tx{}
	err := transaction.Decode(msg.Data())
	if err != nil {
		ilog.Errorf("decode tx error. err=%v", err)
		return nil
	}
	if err := t.txPool.AddTx(transaction, "p2p"); err != nil {
		ilog.Debugf("Add tx failed: %v", err)
		return nil
	}
	return nil
}

func (t *TxManager) handle(msg *p2p.IncomingMessage) {
	_, err := t.pool.ProcessTimed(msg, timeout)
	if err == tunny.ErrJobTimedOut {
		ilog.Warnf("The message %v from %v timed out", msg.Type(), msg.From().Pretty())
	}
}

func (t *TxManager) receiveP2PTxController() {
	for {
		select {
		case msg := <-t.msgCh:
			go t.handle(&msg)
		case <-t.quitCh:
			t.done.Done()
			return
		}
	}
}
