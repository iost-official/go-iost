package synchronizer

import (
	"os"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/iost-official/Go-IOS-Protocol/core/blockcache"
	"github.com/iost-official/Go-IOS-Protocol/core/global"
	"github.com/iost-official/Go-IOS-Protocol/p2p"
	"github.com/iost-official/Go-IOS-Protocol/p2p/mocks"
	. "github.com/smartystreets/goconvey/convey"
)

func TestDownloadController(t *testing.T) {
	Convey("Test DownloadController", t, func() {
		var dHash string
		var dPID p2p.PeerID
		dc, err := NewDownloadController(func(hash string, peerID p2p.PeerID) {
			dHash = hash
			dPID = peerID
		})
		dc.Start()
		So(err, ShouldBeNil)
		Convey("Check OnRecvHash", func() {
			dc.OnRecvHash("111", "aaa")
			time.Sleep(100 * time.Millisecond)
			dc.OnRecvHash("222", "bbb")
			time.Sleep(100 * time.Millisecond)
			dc.OnRecvHash("222", "ccc")
			//dc.OnRecvBlock("123", "abc")
			time.Sleep(300 * time.Millisecond)
			So(dHash, ShouldEqual, "222")
			So(dPID, ShouldEqual, p2p.PeerID("bbb"))
		})
		Convey("Stop DownloadLoop", func() {
			dc.Stop()
		})
	})
}

func TestSynchronizer(t *testing.T) {
	Convey("Test Synchronizer", t, func() {
		baseVariable, err := global.FakeNew()
		So(err, ShouldBeNil)
		So(baseVariable, ShouldNotBeNil)
		defer func() {
			os.RemoveAll("Fakedb")
		}()

		blockCache, err := blockcache.NewBlockCache(baseVariable)
		So(err, ShouldBeNil)
		mockController := gomock.NewController(t)
		mockP2PService := p2p_mock.NewMockService(mockController)
		channel := make(chan p2p.IncomingMessage, 1024)
		mockP2PService.EXPECT().Register(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(channel).AnyTimes()
		mockP2PService.EXPECT().Register(gomock.Any(), gomock.Any()).Return(channel)
		mockP2PService.EXPECT().Broadcast(gomock.Any(), gomock.Any(), gomock.Any()).Do(func(a interface{}, b interface{}, c interface{}) {
			channel <- *p2p.NewIncomingMessage("abc", a.([]byte), b.(p2p.MessageType))
		}).AnyTimes()
		sy, err := NewSynchronizer(baseVariable, blockCache, mockP2PService) //mock
		sy.Start()
		So(err, ShouldBeNil)
		err = sy.syncBlocks(1, 15)
		So(err, ShouldBeNil)
		time.Sleep(200 * time.Millisecond)
	})
}
