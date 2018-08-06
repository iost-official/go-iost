package consensus_common

import (
	"fmt"
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"
)

func TestDownloadController(t *testing.T) {
	Convey("Test DownloadController", t, func() {
		dc, ok := NewDownloadController()
		So(ok, ShouldBeNil)
		var dHash, dPID string
		go dc.DownloadLoop(func(hash, peerID string) {
			dHash = hash
			dPID = peerID
			fmt.Println("CallBack", hash, peerID)
		})
		Convey("Check OnRecvHash", func() {
			dc.OnRecvHash("123", "abc")
			time.Sleep(1 * time.Second)
			//dc.OnRecvBlock("123", "abc")
			time.Sleep(10 * time.Second)
			So(dHash, ShouldEqual, "123")
			So(dPID, ShouldEqual, "abc")
		})
		/*
			Convey("Stop DownloadLoop", func() {
				dc.Stop()
			})
		*/
	})
}
