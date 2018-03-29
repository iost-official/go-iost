package protocol

import (
	"IOS/src/iosbase"
	"fmt"
	"testing"
	"time"
)

type MockNetworkBase struct {
	reqChan chan iosbase.Request
	resChan chan iosbase.Response
}

func NewNetworkBase() MockNetworkBase {
	var req = make(chan iosbase.Request)
	var res = make(chan iosbase.Response)
	return MockNetworkBase{req, res}
}
func (m *MockNetworkBase) Send(req iosbase.Request) chan iosbase.Response {
	fmt.Println("===== send req:")
	fmt.Println(req)
	return m.resChan
}
func (m *MockNetworkBase) Listen(port uint16) (chan iosbase.Request, chan iosbase.Response, error) {

	return m.reqChan, m.resChan, nil
}
func (m *MockNetworkBase) Close(port uint16) error {
	return nil
}
func (m *MockNetworkBase) run() {
	for true {
		res := <-m.resChan
		fmt.Println("----- receive res:", res)
	}
}

type MockView struct {
}

func (m *MockView) GetPrimary() iosbase.Member {
	return iosbase.Member{ID: "test_primary"}
}
func (m *MockView) GetBackup() []iosbase.Member {
	return []iosbase.Member{
		{ID: "test_backup1"},
		{ID: "test_backup2"},
		{ID: "test_backup3"},
	}
}
func (m *MockView) isPrimary(ID string) bool {
	return ID == "test_primary"
}
func (m *MockView) isBackup(ID string) bool {
	return ID == "test_backup1" || ID == "test_backup2" || ID == "test_backup3"
}
func (m *MockView) CommitteeSize() int {
	return 4
}
func (m *MockView) ByzantineTolerance() int {
	return 1
}

func TestNetwork_Route(t *testing.T) {

	mockRuntimeData := RuntimeData{
		Member:    iosbase.Member{ID: "test", Pubkey: make([]byte, 32), Seckey: make([]byte, 32)},
		isRunning: true,
	}
	var nf NetworkFilter
	var netBase = NewNetworkBase()
	go netBase.run()

	err := nf.init(&mockRuntimeData, &netBase)
	if err != nil {
		t.Error(err)
	}

	go nf.router(netBase.reqChan)

	go func() {
		for true {
			select {
			case req1 := <-nf.replicaChan:
				//fmt.Println("ReplicaImpl received:", req1.ReqType)
				if req1.ReqType > 2 {
					t.Fail()
				}
			case req2 := <-nf.recorderChan:
				//fmt.Println("RecorderImpl received:", req2.ReqType)
				if req2.ReqType < 3 || req2.ReqType > 7 {
					t.Fail()
				}
			}
		}
	}()

	for i := 0; i <= 9; i++ {
		req := iosbase.Request{
			From:    "foo",
			To:      "bar",
			ReqType: i,
			Body:    []byte{byte(i)},
		}
		netBase.reqChan <- req
	}

	time.Sleep(1 * time.Second)
}

func TestNetwork_ReplicaFilter(t *testing.T) {

	mockRuntimeData := RuntimeData{
		Member:    iosbase.Member{ID: "test", Pubkey: make([]byte, 32), Seckey: make([]byte, 32)},
		isRunning: true,
		phase:     CommitPhase,
		view:      &MockView{},
	}
	var nf NetworkFilter
	var netBase = NewNetworkBase()
	go netBase.run()

	err := nf.init(&mockRuntimeData, &netBase)
	if err != nil {
		t.Error(err)
	}

	go nf.router(netBase.reqChan)

	replica := ReplicaImpl{
		reqChan:     make(chan iosbase.Request),
		RuntimeData: &mockRuntimeData,
	}

	go func() {
		fmt.Println(<-replica.reqChan)
	}()

	go nf.replicaFilter(&replica, netBase.resChan)

	for i := 0; i <= 2; i++ {
		req := iosbase.Request{
			From:    "test_primary",
			To:      "bar",
			ReqType: i,
			Body:    []byte{byte(i)},
		}
		netBase.reqChan <- req
	}

	time.Sleep(1 * time.Second)

}
