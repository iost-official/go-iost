package event

import (
	"fmt"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

//nolint
func TestEventCollector_Post(t *testing.T) {
	sub1 := NewSubscription(100, []Event_Topic{Event_TransactionResult})
	sub2 := NewSubscription(100, []Event_Topic{Event_ContractEvent})
	sub3 := NewSubscription(100, []Event_Topic{Event_TransactionResult, Event_ContractEvent})

	ec := GetEventCollectorInstance()
	ec.Subscribe(sub1)
	ec.Subscribe(sub2)
	ec.Subscribe(sub3)
	count1 := int32(0)
	count2 := int32(0)
	count3 := int32(0)

	go func(ch <-chan *Event) {
		t.Log("run sub1")
		for {
			select {
			case e := <-ch:
				if e.Topic != Event_TransactionResult {
					t.Fatalf("sub1 expect event topic Event_TransactionResult, got %s", e.Topic.String())
				}
				if e.Data != "test1" {
					t.Fatalf("sub1 expect event data test1, got %s", e.Data)
				}
				atomic.AddInt32(&count1, 1)
			}
		}
	}(sub1.ReadChan())

	go func(ch <-chan *Event) {
		t.Log("run sub2")
		for {
			select {
			case e := <-ch:
				if e.Topic != Event_ContractEvent {
					t.Fatalf("sub2 expect event topic Event_ContractEvent, got %s", e.Topic.String())
				}
				if e.Data != "test2" {
					t.Fatalf("sub2 expect event data test2, got %s", e.Data)
				}
				atomic.AddInt32(&count2, 1)
			}
		}
	}(sub2.ReadChan())

	go func(ch <-chan *Event) {
		t.Log("run sub3")
		for {
			select {
			case e := <-ch:
				if e.Topic != Event_TransactionResult && e.Topic != Event_ContractEvent {
					t.Fatalf("sub3 expect event topic Event_TransactionResult or Event_ContractEvent, got %s", e.Topic.String())
				}
				atomic.AddInt32(&count3, 1)
			}
		}
	}(sub3.ReadChan())

	ec.Post(NewEvent(Event_TransactionResult, "test1"))
	ec.Post(NewEvent(Event_ContractEvent, "test2"))
	ec.Post(NewEvent(Event_ContractEvent, "test2"))

	time.Sleep(time.Millisecond * 100)

	fcount1 := atomic.LoadInt32(&count1)
	fcount2 := atomic.LoadInt32(&count2)
	fcount3 := atomic.LoadInt32(&count3)
	assert.EqualValues(t, 1, fcount1)
	assert.EqualValues(t, 2, fcount2)
	assert.EqualValues(t, 3, fcount3)

	ec.Unsubscribe(sub1)
	ec.Post(NewEvent(Event_TransactionResult, "test1"))
	ec.Post(NewEvent(Event_ContractEvent, "test2"))
	ec.Post(NewEvent(Event_ContractEvent, "test2"))

	time.Sleep(time.Millisecond * 100)

	fcount1 = atomic.LoadInt32(&count1)
	fcount2 = atomic.LoadInt32(&count2)
	fcount3 = atomic.LoadInt32(&count3)
	assert.EqualValues(t, 1, fcount1)
	assert.EqualValues(t, 4, fcount2)
	assert.EqualValues(t, 6, fcount3)
}

//nolint
func TestEventCollector_Full(t *testing.T) {
	sub1 := NewSubscription(1, []Event_Topic{Event_TransactionResult})
	sub2 := NewSubscription(1, []Event_Topic{Event_ContractEvent})
	sub3 := NewSubscription(1, []Event_Topic{Event_TransactionResult, Event_ContractEvent})

	ec := GetEventCollectorInstance()
	ec.Subscribe(sub1)
	ec.Subscribe(sub2)
	ec.Subscribe(sub3)
	count1 := int32(0)
	count2 := int32(0)
	count3 := int32(0)

	ec.Post(NewEvent(Event_TransactionResult, "test1"))
	ec.Post(NewEvent(Event_TransactionResult, "test1"))
	ec.Post(NewEvent(Event_ContractEvent, "test2"))
	ec.Post(NewEvent(Event_ContractEvent, "test2"))
	time.Sleep(time.Millisecond * 100)

	go func(ch <-chan *Event) {
		t.Log("run sub1")
		for {
			select {
			case e := <-ch:
				if e.Topic != Event_TransactionResult {
					t.Fatalf("sub1 expect event topic Event_TransactionResult, got %s", e.Topic.String())
				}
				atomic.AddInt32(&count1, 1)
			}
		}
	}(sub1.ReadChan())

	go func(ch <-chan *Event) {
		t.Log("run sub2")
		for {
			select {
			case e := <-ch:
				if e.Topic != Event_ContractEvent {
					t.Fatalf("sub2 expect event topic Event_ContractEvent, got %s", e.Topic.String())
				}
				atomic.AddInt32(&count2, 1)
			}
		}
	}(sub2.ReadChan())

	go func(ch <-chan *Event) {
		t.Log("run sub3")
		for {
			select {
			case e := <-ch:
				if e.Topic != Event_TransactionResult && e.Topic != Event_ContractEvent {
					t.Fatalf("sub3 expect event topic Event_TransactionResult or Event_ContractEvent, got %s", e.Topic.String())
				}
				atomic.AddInt32(&count3, 1)
			}
		}
	}(sub3.ReadChan())

	time.Sleep(time.Millisecond * 100)
	fcount1 := atomic.LoadInt32(&count1)
	fcount2 := atomic.LoadInt32(&count2)
	fcount3 := atomic.LoadInt32(&count3)
	assert.EqualValues(t, 1, fcount1)
	assert.EqualValues(t, 1, fcount2)
	assert.EqualValues(t, 1, fcount3)

	sub1 = NewSubscription(1000, []Event_Topic{Event_TransactionResult})
	sub2 = NewSubscription(1000, []Event_Topic{Event_ContractEvent})
	sub3 = NewSubscription(1000, []Event_Topic{Event_TransactionResult, Event_ContractEvent})
	ec.Subscribe(sub1)
	ec.Subscribe(sub2)
	ec.Subscribe(sub3)
	data := "test1"
	for i := 0; i < 5; i++ {
		data += data
	}

	t0 := time.Now().Nanosecond()
	// almost 6ms for 10000 post
	for i := 0; i < 1000; i++ {
		ec.Post(NewEvent(Event_TransactionResult, data))
		time.Sleep(time.Microsecond * 50)
	}
	t1 := time.Now().Nanosecond()
	fmt.Println(t1 - t0)
	time.Sleep(time.Millisecond * 100)
	fcount1 = atomic.LoadInt32(&count1)
	fcount2 = atomic.LoadInt32(&count2)
	fcount3 = atomic.LoadInt32(&count3)
	assert.True(t, fcount1 <= 1001, fmt.Sprintf("Expect count1 <= 1001, got count1: %v", fcount1))
	assert.EqualValues(t, 1, fcount2)
	assert.True(t, fcount3 <= 1001, fmt.Sprintf("Expect count3 <= 1001, got count3: %v", fcount3))
}
