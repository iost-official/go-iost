package pob3

import (
	"github.com/iost-official/prototype/core/block"
	"github.com/iost-official/prototype/core/message"
	"github.com/iost-official/prototype/core/tx"
	"github.com/iost-official/prototype/network"
)

type Net struct {
	network.Router
	chTx, chBlock chan message.Message
	exit          chan bool
}

func NewNet(router network.Router) Net {
	return Net{
		Router: router,
	}
}

func (n Net) BroadcastBlock(b *block.Block) {

}
func (n Net) BroadcastTx(tx2 *tx.Tx) {

}
func (n Net) Run() {

}
func (n Net) Stop() {

}
func (n Net) txLoop() {
	for {
		select {
		case <-n.exit:
			return
		case msgTx := <-n.chTx:
			_ = msgTx
		}
	}
}
