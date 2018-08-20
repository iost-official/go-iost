package pob

import (
	"os/exec"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/iost-official/Go-IOS-Protocol/account"
	"github.com/iost-official/Go-IOS-Protocol/common"
	"github.com/iost-official/Go-IOS-Protocol/consensus/synchronizer"
	"github.com/iost-official/Go-IOS-Protocol/core/global"
	"github.com/iost-official/Go-IOS-Protocol/core/new_block"
	"github.com/iost-official/Go-IOS-Protocol/core/new_blockcache"
	"github.com/iost-official/Go-IOS-Protocol/core/new_tx"
	"github.com/iost-official/Go-IOS-Protocol/core/new_txpool"
	"github.com/iost-official/Go-IOS-Protocol/p2p"
	"github.com/iost-official/Go-IOS-Protocol/p2p/mocks"
)

func NewBlock(n int64, parentHash []byte, id2Seckey map[string][]byte) *block.Block {
	blk := block.Block{
		Head: block.BlockHead{
			Version:    0,
			ParentHash: parentHash,
			Number:     n,
			Witness:    witnessOfSec(time.Now().Unix()),
			Time:       time.Now().Unix() / common.SlotLength,
		},
		Txs:      make([]*tx.Tx, 0),
		Receipts: make([]*tx.TxReceipt, 0),
	}
	blk.Head.TxsHash = blk.CalculateTxsHash()
	blk.Head.MerkleHash = blk.CalculateMerkleHash()
	headInfo := generateHeadInfo(blk.Head)
	sig := common.Sign(common.Secp256k1, headInfo, id2Seckey[witnessOfSec(time.Now().Unix())])
	blk.Head.Signature, _ = sig.Encode()
	blk.CalculateHeadHash()
	return &blk
}

func TestBlockLoop(t *testing.T) {
	account1, _ := account.NewAccount(nil)
	account2, _ := account.NewAccount(nil)
	account3, _ := account.NewAccount(nil)
	id2Seckey := make(map[string][]byte)
	id2Seckey[account1.ID] = account1.Seckey
	id2Seckey[account2.ID] = account2.Seckey
	id2Seckey[account3.ID] = account3.Seckey
	baseVariable := global.FakeNew()
	genesisBlock := &block.Block{
		Head: block.BlockHead{
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
	blockCache.Add(genesisBlock)
	mockController := gomock.NewController(t)
	mockP2PService := p2p_mock.NewMockService(mockController)
	channel := make(chan p2p.IncomingMessage, 1024)
	mockP2PService.EXPECT().Register(gomock.Any(), gomock.Any()).Return(channel).AnyTimes()
	txPool, _ := txpool.NewTxPoolImpl(baseVariable, blockCache, mockP2PService)               //mock
	synchronizer, _ := synchronizer.NewSynchronizer(baseVariable, blockCache, mockP2PService) //mock
	witnessList := []string{account1.ID, account2.ID, account3.ID}
	pob := NewPoB(account1, baseVariable, blockCache, txPool, mockP2PService, synchronizer, witnessList)
	nextBlock := NewBlock(1, genesisBlock.HeadHash(), id2Seckey)
	nextBlockByte, _ := nextBlock.Encode()
	incomingMessage := p2p.NewIncomingMessage("id2", nextBlockByte, p2p.NewBlock)
	channel <- *incomingMessage
	pob.blockLoop()
	for {
		continue
	}
	exec.Command("rm", "-r", "./BlockChainDB").Run()
	exec.Command("rm", "-r", "./StateDB").Run()
	exec.Command("rm", "-r", "./txDB").Run()
	exec.Command("rm", "", "priv.key").Run()
}
