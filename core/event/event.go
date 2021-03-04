package event

import (
	"strconv"
	"sync"
	"time"

	"github.com/iost-official/go-iost/v3/ilog"
)

// EventChSize is the size of subscriber's read channel.
const EventChSize = 100

// Topic defines different event topics.
type Topic int

// event topics
const (
	ContractReceipt Topic = iota
	ContractEvent
)

func (t Topic) String() string {
	switch t {
	case ContractReceipt:
		return "ContractReceipt"
	case ContractEvent:
		return "ContractEvent"
	default:
		return "unknown_topic:" + strconv.Itoa(int(t))
	}
}

// Event is the struct sent to subscriber.
type Event struct {
	Topic Topic
	Data  string
	Time  int64
}

// NewEvent generate new event with topic and data
func NewEvent(topic Topic, data string) *Event {
	return &Event{
		Topic: topic,
		Data:  data,
		Time:  time.Now().UnixNano(),
	}
}

// Meta is the information abount event.
type Meta struct {
	ContractID string
}

// Match checks whether the given meta argument is matched to self.
func (m *Meta) Match(meta *Meta) bool {
	if meta == nil {
		return true
	}

	if m.ContractID != "" && m.ContractID != meta.ContractID {
		return false
	}
	return true
}

// Subscription is a struct used for listening specific topics
type Subscription struct {
	C      chan<- *Event
	filter *Meta
}

var ec *Collector
var o sync.Once

// Collector is the struct for posting event.
type Collector struct {
	subMap *sync.Map // map[Topic]map[int64]*Subscription
}

// GetCollector returns single-instance event collector.
func GetCollector() *Collector {
	o.Do(func() {
		ec = &Collector{new(sync.Map)}
	})
	return ec
}

// Subscribe registers a subscription in event collector.
func (ec *Collector) Subscribe(id int64, topics []Topic, filter *Meta) <-chan *Event {
	c := make(chan *Event, EventChSize)
	for _, topic := range topics {
		m, _ := ec.subMap.LoadOrStore(topic, new(sync.Map))
		m.(*sync.Map).Store(id, &Subscription{c, filter})
		ilog.Debugf("Subscribe id = %d, topic = %s, filter = %v", id, topic, filter)
	}
	return c
}

// Unsubscribe deregisters a subscription from event collector.
func (ec *Collector) Unsubscribe(id int64, topics []Topic) {
	for _, topic := range topics {
		m, ok := ec.subMap.Load(topic)
		if ok && m != nil {
			m.(*sync.Map).Delete(id)
			ilog.Debugf("Unsubscribe id = %d, topic = %s", id, topic)
		}
	}
}

func (ec *Collector) sendEvent(e *Event, meta *Meta) {
	if m, exist := ec.subMap.Load(e.Topic); exist {
		m.(*sync.Map).Range(func(k, v interface{}) bool {
			sub := v.(*Subscription)
			if sub.filter != nil && !sub.filter.Match(meta) {
				return true
			}
			select {
			case sub.C <- e:
			default:
				ilog.Debugf("sending event failed. id=%d, topic=%s", k.(int64), e.Topic)
			}
			return true
		})
	}
}

// Post a event.
func (ec *Collector) Post(e *Event, meta *Meta) {
	go ec.sendEvent(e, meta)
}
