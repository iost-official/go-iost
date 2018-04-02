package protocol

import (
	"reflect"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/iost-official/PrototypeWorks/iosbase"
	"github.com/smartystreets/goconvey/convey"
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

//func TestNetwork_Router(t *testing.T) {
//	fmt.Errorf("no")
//
//	mockCtrl := NewController(t)
//	defer mockCtrl.Finish()
//
//	view := NewMockView(mockCtrl)
//
//	mockRuntimeData := RuntimeData{
//		Member:     Member{ID: "test", Pubkey: make([]byte, 32), Seckey: make([]byte, 32)},
//		ExitSignal: make(chan bool),
//	}
//
//	err := mockRuntimeData.SetView(view)
//	if err != nil {
//		t.Fatal(err)
//	}
//
//	netBase := NewMockNetwork(mockCtrl)
//	reqChan := make(chan iosbase.Request)
//	resChan := make(chan iosbase.Response)
//	netBase.EXPECT().Listen(Any()).AnyTimes().Return(reqChan, resChan, nil)
//
//	recorder := NewMockRecorder(mockCtrl)
//
//	replica := NewMockReplica(mockCtrl)
//	holder := NewMockDataHolder(mockCtrl)
//
//	var wg sync.WaitGroup
//
//	wg.Add(1)
//	wg.Done()
//
//	var nf NetworkFilter
//
//	t.Log("Test of network router: ")
//	t.Log("-- Init")
//	err = nf.Init(&mockRuntimeData, netBase, Port)
//	So(err, ShouldBeNil)
//
//	nf.Init(&mockRuntimeData, netBase, Port)
//	go nf.Router(replica, recorder, holder)
//
//	emptyTx := Tx{
//		Time: time.Now().Unix(),
//	}
//	req := iosbase.Request{
//		ReqType: int(ReqPublishTx),
//		Body:    emptyTx.Encode(),
//	}
//	reqChan <- req
//	<-resChan
//	isOK := false
//	recorder.EXPECT().RecordTx(Any()).Do(func(tx Tx) {
//		isOK = tx.Time == emptyTx.Time
//	})
//
//	isOK = false
//	req = iosbase.Request{
//		ReqType: int(ReqCommit),
//		Body:    []byte{127},
//	}
//	reqChan <- req
//	<-resChan
//	view.EXPECT().IsPrimary(Any()).AnyTimes().Return(true)
//	view.EXPECT().IsBackup(Any()).AnyTimes().Return(false)
//	replica.EXPECT().Oniosbase.Request(Any()).AnyTimes().Do(func(iosbase.request iosbase.Request) {
//		isOK = iosbase.request.Body[0] == 127
//	})
//
//	time.Sleep(100 * time.Millisecond)
//}

func TestNetworkFilter_ReplicaFilter(t *testing.T) {
	mockctl := NewController(t)
	defer mockctl.Finish()
	replica := NewMockReplica(mockctl)
	res := make(chan iosbase.Response, 1)

	mockView := NewMockView(mockctl)

	mockRuntimeData := RuntimeData{
		Member:     iosbase.Member{ID: "test", Pubkey: make([]byte, 32), Seckey: make([]byte, 32)},
		ExitSignal: make(chan bool),
		phase:      CommitPhase,
		view:       mockView,
	}

	nf := NetworkFilter{
		RuntimeData: &mockRuntimeData,
	}
	reqCheck := byte(0)

	req := iosbase.Request{
		ReqType: int(ReqCommit),
		Body:    []byte{127},
	}

	Convey("Input should pass through:", t, func() {

		mockView.EXPECT().IsPrimary(Any()).Return(true)
		mockView.EXPECT().IsBackup(Any()).AnyTimes().Return(false)

		replica.EXPECT().OnRequest(Any()).AnyTimes().Do(func(request iosbase.Request) {
			reqCheck = request.Body[0]
		})

		nf.replicaFilter(replica, res, req)

		resp := <-res

		So(resp.Code, ShouldEqual, int(Accepted))

	})

	Convey("Input should successfully passed:", t, func() {
		So(reqCheck, ShouldEqual, byte(127))
	})

	Convey("Input should be rejected for authority:", t, func() {

		//reqCheck := byte(0)
		mockView.EXPECT().IsPrimary(Any()).Return(false)

		nf.replicaFilter(replica, res, req)

		resp := <-res
		So(resp.Description, ShouldEqual, "Error: Authority error")
	})
	Convey("Input should be rejected for phase error:", t, func() {

		mockView.EXPECT().IsPrimary(Any()).Return(true)
		nf.RuntimeData.phase = PrePreparePhase

		nf.replicaFilter(replica, res, req)

		resp := <-res
		So(resp.Description, ShouldEqual, "Error: Invalid phase")

	})
}

func TestNetworkFilter_Router(t *testing.T) {
	mockctl := NewController(t)
	defer mockctl.Finish()
	rep := NewMockReplica(mockctl)
	rec := NewMockRecorder(mockctl)
	hol := NewMockDataHolder(mockctl)

	mockRuntimeData := RuntimeData{
		ExitSignal: make(chan bool),
	}

	reqChan := make(chan iosbase.Request)
	resChan := make(chan iosbase.Response)

	nf := NetworkFilter{
		RuntimeData: &mockRuntimeData,
		resChan:     resChan,
		reqChan:     reqChan,
	}

	Convey("Should send req correctly", t, func() {

	})

	Convey("Should stop normally:", t, func() {
		isEnd := false
		go func() {
			nf.Router(rep, rec, hol)
			isEnd = true
		}()

		mockRuntimeData.ExitSignal <- true

		So(isEnd, ShouldBeTrue)

	})
}
