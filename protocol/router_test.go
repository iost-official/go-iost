package protocol

import (
	"reflect"
	"sync"
	"testing"
	"time"

	"github.com/iost-official/Go-IOS-Protocol/iosbase"
	"github.com/iost-official/Go-IOS-Protocol/iosbase/mocks"

	. "github.com/golang/mock/gomock"
	. "github.com/smartystreets/goconvey/convey"
)

func TestRouter(t *testing.T) {
	Convey("Test of RouterFactory", t, func() {
		router, err := RouterFactory("base")
		So(err, ShouldBeNil)
		So(reflect.TypeOf(router), ShouldEqual, reflect.TypeOf(&RouterImpl{}))
	})

	Convey("Test of RouterImpl", t, func() {
		mockctl := NewController(t)
		defer mockctl.Finish()

		netBase := iosbase_mock.NewMockNetwork(mockctl)
		chRes := make(chan iosbase.Response, 10)
		chReq := make(chan iosbase.Request, 10)
		netBase.EXPECT().Listen(Any()).AnyTimes().Return(chReq, chRes, nil)
		router, err := RouterFactory("base")

		Convey("Init:", func() {
			err = router.Init(netBase, Port)
			So(err, ShouldBeNil)
		})

		Convey("Run and stop:", func() {
			router.Init(netBase, Port)

			var wg sync.WaitGroup
			wg.Add(1)
			go func() {
				router.Run()
				wg.Done()
			}()
			time.Sleep(10 * time.Millisecond)
			router.Stop()
			wg.Wait()
			So(true, ShouldBeTrue)
		})

		Convey("Type Filter white list:", func() {
			router.Init(netBase, Port)

			typeFilter, _, err := router.FilteredChan(Filter{
				AcceptType: []ReqType{ReqPublishTx},
			})

			chReq <- iosbase.Request{
				ReqType: int(ReqPublishTx),
			}
			go router.Run()
			So(err, ShouldBeNil)
			req := <-typeFilter
			So(req.ReqType, ShouldEqual, int(ReqPublishTx))
		})

		Convey("Type Filter black list:", func() {
			router.Init(netBase, Port)

			typeFilter, _, err := router.FilteredChan(Filter{
				RejectType: []ReqType{ReqPublishTx},
			})

			chReq <- iosbase.Request{
				ReqType: int(ReqPublishTx),
			}
			chReq <- iosbase.Request{
				ReqType: int(ReqCommit),
			}
			go router.Run()
			So(err, ShouldBeNil)
			req := <-typeFilter
			So(req.ReqType, ShouldEqual, int(ReqCommit))
		})
	})
}
