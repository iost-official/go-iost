package pob

import (
	"github.com/iost-official/Go-IOS-Protocol/account"
	"github.com/iost-official/Go-IOS-Protocol/consensus/synchronizer"
	"github.com/iost-official/Go-IOS-Protocol/core/global"
	"github.com/iost-official/Go-IOS-Protocol/core/new_blockcache"
	"github.com/iost-official/Go-IOS-Protocol/core/new_txpool"
	"testing"
	"github.com/iost-official/Go-IOS-Protocol/core/new_block"
	"github.com/iost-official/Go-IOS-Protocol/core/new_tx"
	"github.com/iost-official/Go-IOS-Protocol/p2p/mocks"
	"github.com/golang/mock/gomock"
	"github.com/iost-official/Go-IOS-Protocol/p2p"
	"time"
	"github.com/iost-official/Go-IOS-Protocol/common"
)

func NewBlock(n int64, parentHash []byte) *block.Block {
	return &block.Block{
		Head: block.BlockHead{
			Version: 0,
			ParentHash: parentHash,
			Number:  n,
			Time:    time.Now().Unix()/common.SlotLength,
			Witness: witnessOfSec(time.Now().Unix()),
		},
		Txs:      make([]*tx.Tx, 0),
		Receipts: make([]*tx.TxReceipt, 0),
	}
}

func TestBlockLoop(t *testing.T) {
	account, _ := account.NewAccount(nil)
	baseVariable := global.FakeNew()
	block := &block.Block{
		Head: block.BlockHead{
			Version: 0,
			Number:  0,
			Time:    0,
		},
		Txs:      make([]*tx.Tx, 0),
		Receipts: make([]*tx.TxReceipt, 0),
	}
	block.CalculateHeadHash()
	baseVariable.BlockChain().Push(block)
	blockCache, _ := blockcache.NewBlockCache(baseVariable)
	mockController := gomock.NewController(t)
	mockP2PService := p2p_mock.NewMockService(mockController)
	channel := make(chan p2p.IncomingMessage, 1024)
	mockP2PService.EXPECT().Register(gomock.Any(), gomock.Any()).Return(channel).AnyTimes()
	txPool, _ := txpool.NewTxPoolImpl(baseVariable, blockCache, mockP2PService)               //mock
	synchronizer, _ := synchronizer.NewSynchronizer(baseVariable, blockCache, mockP2PService) //mock
	witnessList := []string{account.ID, "id2", "id3"}
	pob := NewPoB(account, baseVariable, blockCache, txPool, mockP2PService, synchronizer, witnessList)
	go pob.blockLoop()
	nextBlock := NewBlock(1, block.HeadHash())
	nextBlockByte, _ := nextBlock.Encode()
	incomingMessage := p2p.NewIncomingMessage("id2", nextBlockByte, p2p.NewBlock)
	channel <- *incomingMessage
	for {
		continue
	}
	cmd :=
}
