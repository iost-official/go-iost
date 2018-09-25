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
		dHash := make(chan string, 10)
		dPID := make(chan p2p.PeerID, 10)
		dc, err := NewDownloadController(func(hash string, peerID p2p.PeerID) {
			dHash <- hash
			dPID <- peerID
		})
		dc.Start()
		So(err, ShouldBeNil)
		Convey("Check OnRecvHash", func() {
			dc.OnRecvHash("111", "aaa")
			dc.OnRecvHash("222", "bbb")
			dc.OnRecvHash("222", "ccc")
			//dc.OnRecvBlock("123", "abc")
			var hash string
			var PID p2p.PeerID
			hash = <-dHash
			PID = <-dPID
			So(hash, ShouldEqual, "111")
			So(PID, ShouldEqual, p2p.PeerID("aaa"))

			hash = <-dHash
			PID = <-dPID
			So(hash, ShouldEqual, "222")
			So(PID, ShouldEqual, p2p.PeerID("bbb"))
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
