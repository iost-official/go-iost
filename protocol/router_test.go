package protocol

import (
	"testing"
	"github.com/golang/mock/gomock"
	"reflect"
	"github.com/iost-official/PrototypeWorks/iosbase"
	. "github.com/smartystreets/goconvey/convey"
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



func TestRouter (t *testing.T) {
	Convey("Test of RouterFactory", t, func() {
		router, err := RouterFactory("base")
		So(err, ShouldBeNil)
		So(reflect.TypeOf(router), ShouldEqual, reflect.TypeOf(&RouterImpl{}))
	})

	//Convey("Test of RouterImpl", t, func() {
	//	error
	//})
}
