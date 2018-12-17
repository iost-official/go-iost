package event_test

import (
	"sync/atomic"
	"testing"
	"time"

	"github.com/iost-official/go-iost/core/event"
	"github.com/iost-official/go-iost/ilog"
	"github.com/stretchr/testify/assert"
)

func TestEventCollectorPost(t *testing.T) {
	ilog.Stop()
	ec := event.GetCollector()

	ch1 := ec.Subscribe(1, []event.Topic{event.ContractEvent}, nil)
	ch2 := ec.Subscribe(2, []event.Topic{event.ContractEvent}, &event.Meta{ContractID: "base.iost"})
	ch3 := ec.Subscribe(3, []event.Topic{event.ContractReceipt, event.ContractEvent}, nil)

	count1 := int32(0)
	count2 := int32(0)
	count3 := int32(0)

	go func(ch <-chan *event.Event) {
		t.Log("select ch1")
		for {
			select {
			case e := <-ch:
				assert.Equal(t, event.ContractEvent, e.Topic)
				if e.Data != "test1" && e.Data != "test2" {
					t.Fatalf("sub1 expect event data test1/2, got %s", e.Data)
				}
				atomic.AddInt32(&count1, 1)
			}
		}
	}(ch1)

	go func(ch <-chan *event.Event) {
		t.Log("select ch2")
		for {
			select {
			case e := <-ch:
				assert.Equal(t, event.ContractEvent, e.Topic)
				assert.Equal(t, "test2", e.Data)
				atomic.AddInt32(&count2, 1)
			}
		}
	}(ch2)

	go func(ch <-chan *event.Event) {
		t.Log("select ch3")
		for {
			select {
			case e := <-ch:
				if e.Topic != event.ContractEvent && e.Topic != event.ContractReceipt {
					t.Fatalf("sub3 expect event topic ContractEvent or ContractReceipt, got %s", e.Topic)
				}
				atomic.AddInt32(&count3, 1)
			}
		}
	}(ch3)

	ec.Post(event.NewEvent(event.ContractEvent, "test1"), &event.Meta{ContractID: "token.iost"})
	ec.Post(event.NewEvent(event.ContractEvent, "test2"), &event.Meta{ContractID: "base.iost"})
	ec.Post(event.NewEvent(event.ContractReceipt, "test3"), &event.Meta{ContractID: "base.iost"})

	time.Sleep(time.Millisecond * 100)

	assert.EqualValues(t, 2, atomic.LoadInt32(&count1))
	assert.EqualValues(t, 1, atomic.LoadInt32(&count2))
	assert.EqualValues(t, 3, atomic.LoadInt32(&count3))

	ec.Unsubscribe(1, []event.Topic{event.ContractEvent})
	ec.Post(event.NewEvent(event.ContractEvent, "test2"), &event.Meta{ContractID: "base.iost"})

	time.Sleep(time.Millisecond * 100)

	assert.EqualValues(t, 2, atomic.LoadInt32(&count1))
	assert.EqualValues(t, 2, atomic.LoadInt32(&count2))
	assert.EqualValues(t, 4, atomic.LoadInt32(&count3))
}

func TestEventCollectorFullPost(t *testing.T) {
	ilog.Stop()

	ec := event.GetCollector()
	ch := ec.Subscribe(1, []event.Topic{event.ContractEvent}, nil)

	count := int32(0)

	for i := 0; i < event.EventChSize+100; i++ {
		ec.Post(event.NewEvent(event.ContractEvent, "test1"), &event.Meta{ContractID: "token.iost"})
	}

	time.Sleep(time.Millisecond * 100)

	go func(ch <-chan *event.Event) {
		t.Log("select ch")
		for {
			select {
			case e := <-ch:
				assert.Equal(t, event.ContractEvent, e.Topic)
				atomic.AddInt32(&count, 1)
			}
		}
	}(ch)

	time.Sleep(time.Millisecond * 100)

	assert.EqualValues(t, event.EventChSize, atomic.LoadInt32(&count))
}
