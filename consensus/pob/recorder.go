package pob

import (
	"errors"
	"sync"

	common "github.com/iost-official/prototype/consensus/common"
	"github.com/iost-official/prototype/core/tx"
)

var ErrIllegalTx = errors.New("illegal tx")

type Recorder interface {
	Record(tx2 tx.Tx) error
	Pop() tx.Tx
	Listen()
	Close()
}

func NewRecorder() Recorder {
	return &RecorderImpl{
		TxHeap: NewTxHeap(),
	}
}

type RecorderImpl struct {
	TxHeap
	isClose     bool
	isListening sync.Mutex
}

func (p *RecorderImpl) Record(tx2 tx.Tx) error {
	p.isListening.Lock()
	defer p.isListening.Unlock()
	if common.VerifyTxSig(tx2) {
		p.Push(tx2)
		return nil
	} else {
		return ErrIllegalTx
	}
}

func (p *RecorderImpl) Listen() {
	if p.isClose == true {
		p.isClose = false
	} else {
		return
	}
	p.isListening.Unlock()
}

func (p *RecorderImpl) Close() {
	if p.isClose == false {
		p.isClose = true
	} else {
		return
	}
	p.isListening.Lock()
}
