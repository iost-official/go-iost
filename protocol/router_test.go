package protocol

import (
	"testing"
	"github.com/golang/mock/gomock"
	"reflect"
	"github.com/iost-official/PrototypeWorks/iosbase"
	. "github.com/smartystreets/goconvey/convey"
	. "github.com/golang/mock/gomock"
	"sync"
	"time"
)

// MockNetwork is a mock of Network interface
type MockNetwork struct {
	ctrl     *gomock.Controller
	recorder *MockNetworkMockRecorder
}

// MockNetworkMockRecorder is the mock recorder for MockNetwork
type MockNetworkMockRecorder struct {
	mock *MockNetwork
}

// NewMockNetwork creates a new mock instance
func NewMockNetwork(ctrl *gomock.Controller) *MockNetwork {
	mock := &MockNetwork{ctrl: ctrl}
	mock.recorder = &MockNetworkMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockNetwork) EXPECT() *MockNetworkMockRecorder {
	return m.recorder
}

// Send mocks base method
func (m *MockNetwork) Send(req iosbase.Request) chan iosbase.Response {
	ret := m.ctrl.Call(m, "Send", req)
	ret0, _ := ret[0].(chan iosbase.Response)
	return ret0
}

// Send indicates an expected call of Send
func (mr *MockNetworkMockRecorder) Send(req interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Send", reflect.TypeOf((*MockNetwork)(nil).Send), req)
}

// Listen mocks base method
func (m *MockNetwork) Listen(port uint16) (chan iosbase.Request, chan iosbase.Response, error) {
	ret := m.ctrl.Call(m, "Listen", port)
	ret0, _ := ret[0].(chan iosbase.Request)
	ret1, _ := ret[1].(chan iosbase.Response)
	ret2, _ := ret[2].(error)
	return ret0, ret1, ret2
}

// Listen indicates an expected call of Listen
func (mr *MockNetworkMockRecorder) Listen(port interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Listen", reflect.TypeOf((*MockNetwork)(nil).Listen), port)
}

// Close mocks base method
func (m *MockNetwork) Close(port uint16) error {
	ret := m.ctrl.Call(m, "Close", port)
	ret0, _ := ret[0].(error)
	return ret0
}

// Close indicates an expected call of Close
func (mr *MockNetworkMockRecorder) Close(port interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Close", reflect.TypeOf((*MockNetwork)(nil).Close), port)
}

func TestRouter(t *testing.T) {
	Convey("Test of RouterFactory", t, func() {
		router, err := RouterFactory("base")
		So(err, ShouldBeNil)
		So(reflect.TypeOf(router), ShouldEqual, reflect.TypeOf(&RouterImpl{}))
	})

	Convey("Test of RouterImpl", t, func() {
		mockctl := NewController(t)
		defer mockctl.Finish()

		netBase := NewMockNetwork(mockctl)
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
			time.Sleep(10*time.Millisecond)
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
