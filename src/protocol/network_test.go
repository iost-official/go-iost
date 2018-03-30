package protocol

import (
	"IOS/src/iosbase"
	"testing"
	//. "github.com/bouk/monkey"
	//"reflect"
	. "github.com/smartystreets/goconvey/convey"
	. "github.com/golang/mock/gomock"

	"time"
	"sync"
)

func TestNetwork_Router(t *testing.T) {
	mockRuntimeData := RuntimeData{
		Member:    iosbase.Member{ID: "test", Pubkey: make([]byte, 32), Seckey: make([]byte, 32)},
		isRunning: true,
	}

	mockCtrl := NewController(t)
	defer mockCtrl.Finish()

	netBase := iosbase.NewMockNetwork(mockCtrl)
	reqChan := make(chan iosbase.Request)
	resChan := make(chan iosbase.Response)
	netBase.EXPECT().Listen(Any()).Return(reqChan, resChan, nil)

	recorder := NewMockRecorder(mockCtrl)
	recorder.EXPECT().PublishTx(Any())

	replica := NewMockReplica(mockCtrl)
	holder := NewMockDataHolder(mockCtrl)

	var wg sync.WaitGroup

	var nf NetworkFilter

	Convey("Test of network router: ", t, func() {
		Convey("Should Init success and return nil", func() {
			err := nf.Init(&mockRuntimeData, netBase, Port)
			So(err, ShouldBeNil)
		})


		go func() {
			wg.Add(1)
			defer wg.Done()
			nf.Router(replica, recorder, holder)
		}()

		Convey("Shouled Send to Recorder", func() {
			emptyTx := iosbase.Tx{}
			reqChan <- iosbase.Request{
				ReqType: int(ReqPublishTx),
				Body: emptyTx.Encode(),
			}
			So(true, ShouldBeTrue)
		})

	})

	time.Sleep(time.Second)
	mockRuntimeData.isRunning = false
	//wg.Wait()
}

//func TestNetwork_Filter(t *testing.T) {
//
//	mockRuntimeData := RuntimeData{
//		Member:    iosbase.Member{ID: "test", Pubkey: make([]byte, 32), Seckey: make([]byte, 32)},
//		isRunning: true,
//		phase:     CommitPhase,
//		view:      &MockView{},
//	}
//	var nf NetworkFilter
//	var netBase = NewNetworkBase()
//	go netBase.run()
//
//	err := nf.Init(&mockRuntimeData, &netBase)
//	if err != nil {
//		t.Error(err)
//	}
//
//	go nf.Router(netBase.reqChan)
//
//	replica := ReplicaImpl{
//		reqChan:     make(chan iosbase.Request),
//		RuntimeData: &mockRuntimeData,
//	}
//
//	go func() {
//		for true {
//			fmt.Println(<-replica.reqChan)
//		}
//	}()
//
//	recorder := MockRecorder{}
//
//	go nf.replicaFilter(&replica, netBase.resChan)
//
//	go nf.recorderFilter(&recorder, netBase.resChan)
//
//	for i := 0; i <= 9; i++ {
//		req := iosbase.Request{
//			From:    "test_primary",
//			To:      "bar",
//			ReqType: i,
//			Body:    []byte{byte(i)},
//		}
//		netBase.reqChan <- req
//	}
//
//	time.Sleep(1 * time.Second)
//
//}
