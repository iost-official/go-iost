package pob

import (
	"os/exec"
	"testing"
	"time"

	"fmt"

	"github.com/golang/mock/gomock"
	"github.com/iost-official/go-iost/account"
	"github.com/iost-official/go-iost/common"
	"github.com/iost-official/go-iost/core/block"
	"github.com/iost-official/go-iost/core/blockcache"
	"github.com/iost-official/go-iost/core/global"
	"github.com/iost-official/go-iost/core/tx"
	"github.com/iost-official/go-iost/core/txpool"
	"github.com/iost-official/go-iost/crypto"
	"github.com/iost-official/go-iost/ilog"
	"github.com/iost-official/go-iost/p2p"
	"github.com/iost-official/go-iost/p2p/mocks"
)

func testRun(t *testing.T) {
	exec.Command("rm", "-r", "./BlockChainDB").Run()
	exec.Command("rm", "-r", "./StateDB").Run()
	exec.Command("rm", "-r", "./TXDB").Run()
	exec.Command("rm", "", "priv.key").Run()
	account1, _ := account.NewKeyPair(nil, crypto.Secp256k1)
	account2, _ := account.NewKeyPair(nil, crypto.Secp256k1)
	account3, _ := account.NewKeyPair(nil, crypto.Secp256k1)
	id2Seckey := make(map[string][]byte)
	id2Seckey[account1.ReadablePubkey()] = account1.Seckey
	id2Seckey[account2.ReadablePubkey()] = account2.Seckey
	id2Seckey[account3.ReadablePubkey()] = account3.Seckey
	baseVariable, _ := global.New(&common.Config{
		DB: &common.DBConfig{
			LdbPath: "Fakedb/",
		},
	})
	genesisBlock := &block.Block{
		Head: &block.BlockHead{
			Version: 0,
			Number:  0,
			Time:    0,
		},
		Txs:      make([]*tx.Tx, 0),
		Receipts: make([]*tx.TxReceipt, 0),
	}
	genesisBlock.CalculateHeadHash()
	baseVariable.BlockChain().Push(genesisBlock)
	blockCache, _ := blockcache.NewBlockCache(baseVariable)
	baseVariable.StateDB().Tag(string(genesisBlock.HeadHash()))
	mockController := gomock.NewController(t)
	mockP2PService := p2p_mock.NewMockService(mockController)
	channel := make(chan p2p.IncomingMessage, 1024)
	mockP2PService.EXPECT().Register(gomock.Any(), gomock.Any()).Return(channel).AnyTimes()
	txPool, _ := txpool.NewTxPoolImpl(baseVariable, blockCache, mockP2PService) //mock
	pob := New(account1, baseVariable, blockCache, txPool, mockP2PService)
	pob.Start()
	fmt.Println(time.Now().Second())
	fmt.Println(time.Now().Nanosecond())
	fw := ilog.NewFileWriter("pob/")
	ilog.AddWriter(fw)
	select {}
}
