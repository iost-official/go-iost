package rpc

import (
	"errors"
	"github.com/iost-official/Go-IOS-Protocol/core/event"
	"testing"
	"time"
)

type Mock_Apis_SubscribeServer struct {
	Apis_SubscribeServer
	count int
}

func (s *Mock_Apis_SubscribeServer) Send(req *SubscribeRes) error {
	s.count++
	if req.Ev.Topic != event.Event_TransactionResult || req.Ev.Data != "test1" {
		return errors.New("unexpected event topic or data. ev = " + req.Ev.String())
	}
	return nil
}

func TestRpcServer_Subscribe(t *testing.T) {
	s := newRPCServer()
	ec := event.GetEventCollectorInstance()
	req := &SubscribeReq{Topics: []event.Event_Topic{event.Event_TransactionResult}}
	res := Mock_Apis_SubscribeServer{
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
