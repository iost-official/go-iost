package synchronizer

import (
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/iost-official/Go-IOS-Protocol/core/global"
	"github.com/iost-official/Go-IOS-Protocol/core/new_block"
	"github.com/iost-official/Go-IOS-Protocol/core/new_blockcache"
	"github.com/iost-official/Go-IOS-Protocol/core/new_tx"
	"github.com/iost-official/Go-IOS-Protocol/p2p"
	"github.com/iost-official/Go-IOS-Protocol/p2p/mocks"
	. "github.com/smartystreets/goconvey/convey"
)

func TestDownloadController(t *testing.T) {
	Convey("Test DownloadController", t, func() {
		dc, err := NewDownloadController()
		So(err, ShouldBeNil)
		var dHash string
		var dPID p2p.PeerID
		go dc.DownloadLoop(func(hash string, peerID p2p.PeerID) {
			dHash = hash
			dPID = peerID
		})
		Convey("Check OnRecvHash", func() {
			dc.OnRecvHash("111", "aaa")
			time.Sleep(time.Second)
			dc.OnRecvHash("222", "bbb")
			time.Sleep(time.Second)
			dc.OnRecvHash("222", "ccc")
			//dc.OnRecvBlock("123", "abc")
			time.Sleep(7 * time.Second)
			So(dHash, ShouldEqual, "222")
			So(dPID, ShouldEqual, "bbb")
		})
		Convey("Stop DownloadLoop", func() {
			dc.Stop()
		})
	})
}

func TestSynchronizer(t *testing.T) {
	Convey("Test Synchronizer", t, func() {
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
		baseVariable.StateDB().Tag(string(genesisBlock.HeadHash()))
		mockController := gomock.NewController(t)
		mockP2PService := p2p_mock.NewMockService(mockController)
		channel := make(chan p2p.IncomingMessage, 1024)
		mockP2PService.EXPECT().Register(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(channel).AnyTimes()
		mockP2PService.EXPECT().Broadcast(gomock.Any(), gomock.Any(), gomock.Any()).Do(func(a interface{}, b interface{}, c interface{}) {
			channel <- *p2p.NewIncomingMessage("abc", a.([]byte), b.(p2p.MessageType))
		}).AnyTimes()
		sy, err := NewSynchronizer(baseVariable, blockCache, mockP2PService) //mock
		sy.Start()
		So(err, ShouldBeNil)
		err = sy.SyncBlocks(1, 15)
		So(err, ShouldBeNil)
		time.Sleep(2 * time.Second)
	})
}
