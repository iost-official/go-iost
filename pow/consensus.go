package pow

import (
	"github.com/iost-official/prototype/core"
	"encoding/binary"
)

type Pow struct {
	Recorder
	Holder
	Router

	ExitSignal chan bool

	chTx    chan core.Request
	chBlock chan core.Request
}

const (
	Port = 23333
)

func (p *Pow) Init(bc core.BlockChain, sp core.UTXOPool, network core.Network) error {
	var err error
	p.Holder = NewHolder(bc, sp)
	p.Recorder = NewRecorder()
	p.Router, err = RouterFactory("base")
	if err != nil {
		return err
	}
	p.Router.Init(network, 23333)

	p.chTx, err = p.Router.FilteredChan(Filter{
		AcceptType: []ReqType{ReqPublishTx},
	})

	p.chBlock, err = p.Router.FilteredChan(Filter{
		AcceptType: []ReqType{ReqNewBlock},
	})

	return err

}
func (p *Pow) Run() {
	go p.Router.Run()

}
func (p *Pow) Stop()                                                    {}
func (p *Pow) PublishTx(tx core.Tx) error                               {}
func (p *Pow) CheckTx(tx core.Tx) (core.TxStatus, error)                {}
func (p *Pow) GetStatus() (core.BlockChain, core.UTXOPool, error)       {}
func (p *Pow) GetCachedStatus() (core.BlockChain, core.UTXOPool, error) {}

func parseInfo(head core.BlockHead) (difficulty, nonce uint64) {
	difficulty = binary.BigEndian.Uint64(head.Info[0:8])
	nonce = binary.BigEndian.Uint64(head.Info[8:16])
	return
}

func makeInfo(dif, nonce uint64) []byte {
	b1 := make([]byte, 8)
	binary.BigEndian.PutUint64(b1, dif)
	b2 := make([]byte, 8)
	binary.BigEndian.PutUint64(b2, nonce)

	return append(b1, b2...)
}

func (p *Pow) txListenLoop() {
	for {
		req, ok := <-p.chTx
		if !ok {
			return
		}
		var tx core.Tx
		tx.Decode(req.Body)
		p.PublishTx(tx)
	}
}

func (p *Pow) blockLoop() {
	for {
		req, ok := <-p.chBlock
		if !ok {
			return
		}
		var blk core.Block
		blk.Decode(req.Body)
		p.Holder.Add(&blk)

	}
}

func (p *Pow) mineLoop() {
	for {
		select {
		case <-p.ExitSignal:
			return
		default:
			p.Holder.CacheTop()
		}
	}
}
