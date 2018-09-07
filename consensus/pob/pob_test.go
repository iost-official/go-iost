package pob

import (
	"os/exec"
	"testing"
	"time"

	"fmt"

	"github.com/golang/mock/gomock"
	"github.com/iost-official/Go-IOS-Protocol/account"
	"github.com/iost-official/Go-IOS-Protocol/consensus/synchronizer"
	"github.com/iost-official/Go-IOS-Protocol/core/block"
	"github.com/iost-official/Go-IOS-Protocol/core/blockcache"
	"github.com/iost-official/Go-IOS-Protocol/core/global"
	"github.com/iost-official/Go-IOS-Protocol/core/tx"
	"github.com/iost-official/Go-IOS-Protocol/core/txpool"
	"github.com/iost-official/Go-IOS-Protocol/crypto"
	"github.com/iost-official/Go-IOS-Protocol/ilog"
	"github.com/iost-official/Go-IOS-Protocol/p2p"
	"github.com/iost-official/Go-IOS-Protocol/p2p/mocks"
)

func testRun(t *testing.T) {
	exec.Command("rm", "-r", "./BlockChainDB").Run()
	exec.Command("rm", "-r", "./StateDB").Run()
	exec.Command("rm", "-r", "./TXDB").Run()
	exec.Command("rm", "", "priv.key").Run()
	account1, _ := account.NewAccount(nil, crypto.Secp256k1)
	account2, _ := account.NewAccount(nil, crypto.Secp256k1)
	account3, _ := account.NewAccount(nil, crypto.Secp256k1)
	id2Seckey := make(map[string][]byte)
	id2Seckey[account1.ID] = account1.Seckey
	id2Seckey[account2.ID] = account2.Seckey
	id2Seckey[account3.ID] = account3.Seckey
	baseVariable := global.FakeNew()
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
	txPool, _ := txpool.NewTxPoolImpl(baseVariable, blockCache, mockP2PService)               //mock
	synchronizer, _ := synchronizer.NewSynchronizer(baseVariable, blockCache, mockP2PService) //mock
	witnessList := []string{account1.ID, account2.ID, account3.ID}
	pob := NewPoB(account1, baseVariable, blockCache, txPool, mockP2PService, synchronizer, witnessList)
	pob.Start()
	fmt.Println(time.Now().Second())
	fmt.Println(time.Now().Nanosecond())
	fw := ilog.NewFileWriter("pob/")
	ilog.AddWriter(fw)
	select {}
}
