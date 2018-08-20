package pob

import (
	"github.com/iost-official/Go-IOS-Protocol/account"
	"github.com/iost-official/Go-IOS-Protocol/consensus/synchronizer"
	"github.com/iost-official/Go-IOS-Protocol/core/global"
	"github.com/iost-official/Go-IOS-Protocol/core/new_blockcache"
	"github.com/iost-official/Go-IOS-Protocol/core/new_txpool"
	"github.com/iost-official/Go-IOS-Protocol/p2p"
	"testing"
)

func TestNewPoB(t *testing.T) {
	account := account.Account{}
	baseVariable := global.FakeNew()
	blockCache, _ := blockcache.NewBlockCache(baseVariable)
	p2pService, _ := p2p.NewDefault()
	txPool, _ := txpool.NewTxPoolImpl(baseVariable, blockCache, p2pService)
	synchronizer, _ := synchronizer.NewSynchronizer(baseVariable, blockCache, p2pService)
	witnessList := []string{"id1", "id2", "id3"}
	NewPoB(account, baseVariable, blockCache, txPool, p2pService, synchronizer, witnessList)
}
