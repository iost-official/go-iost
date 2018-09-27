package rpc

import (
	"errors"

	"github.com/iost-official/go-iost/core/event"

	"testing"
	"time"

	"github.com/iost-official/go-iost/core/blockcache"
	"github.com/iost-official/go-iost/core/global"
	"github.com/iost-official/go-iost/core/txpool"

	"github.com/bouk/monkey"
	"github.com/iost-official/go-iost/p2p"
)

type MockApisSubscribeServer struct {
	Apis_SubscribeServer
	count int
}

func (s *MockApisSubscribeServer) Send(req *SubscribeRes) error {
	s.count++
	if req.Ev.Topic != event.Event_TransactionResult || req.Ev.Data != "test1" {
		return errors.New("unexpected event topic or data. ev = " + req.Ev.String())
	}
	return nil
}

func TestRpcServer_Subscribe(t *testing.T) {
	monkey.Patch(NewRPCServer, func(tp txpool.TxPool, bcache blockcache.BlockCache, _global global.BaseVariable, p2pService p2p.Service) *GRPCServer {
		return &GRPCServer{}
	})

	s := NewRPCServer(nil, nil, nil, nil)
	ec := event.GetEventCollectorInstance()
	req := &SubscribeReq{Topics: []event.Event_Topic{event.Event_TransactionResult}}
	res := MockApisSubscribeServer{
		count: 0,
	}
	go func() {
		err := s.Subscribe(req, &res)
		if err != nil {
			t.Fatalf(err.Error())
		}
		if res.count > 5000 {
			t.Fatalf("should send <= 5000 events. got %d", res.count)
		} else {
			t.Logf("send %d events.", res.count)
		}
	}()

	for i := 0; i < 5000; i++ {
		ec.Post(event.NewEvent(event.Event_TransactionResult, "test1"))
		time.Sleep(time.Microsecond * 50)
	}
}
