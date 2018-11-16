package event

import (
	"sync"
	"time"

	log "github.com/iost-official/go-iost/ilog"
)

// EventCollectorSize size of post channel in event collector
const EventCollectorSize = 1024

// NewEvent generate new event with topic and data
func NewEvent(topic Event_Topic, data string) *Event {
	now := time.Now().UnixNano()
	return &Event{
		Topic: topic,
		Data:  data,
		Time:  now,
	}
}

// Subscription is a struct used for listening specific topics
type Subscription struct {
	topics []Event_Topic
	postCh chan<- *Event
	readCh <-chan *Event
}

// NewSubscription new subscription with specific topics and size for post channel
func NewSubscription(chSize int, topics []Event_Topic) *Subscription {
	postCh := make(chan *Event, chSize)
	subscribe := &Subscription{
		topics: topics,
		postCh: postCh,
		readCh: postCh,
	}
	return subscribe
}

// ReadChan get channel for reading event
func (s *Subscription) ReadChan() <-chan *Event {
	return s.readCh
}

// nolint
// EventCollector struct for posting event
type EventCollector struct {
	subMap *sync.Map
	postCh chan *Event
	quitCh chan int
}

var ec *EventCollector

// GetEventCollectorInstance return single-instance event collector
func GetEventCollectorInstance() *EventCollector {
	if ec == nil {
		ec = &EventCollector{
			subMap: new(sync.Map),
			postCh: make(chan *Event, EventCollectorSize),
			quitCh: make(chan int, 1),
		}
		ec.start()
	}
	return ec
}

func (ec *EventCollector) start() {
	go ec.deliverLoop()
}

// Stop event collector if no longer in use
func (ec *EventCollector) Stop() {
	ec.quitCh <- 1
}

// Post a event
func (ec *EventCollector) Post(e *Event) {
	select {
	case ec.postCh <- e:
	default:
		log.Debugf("post event failed, aborted. topic = %s", e.Topic)
	}

}

// Subscribe a subscription to event collector
func (ec *EventCollector) Subscribe(sub *Subscription) {
	for _, topic := range sub.topics {
		m, _ := ec.subMap.LoadOrStore(topic, new(sync.Map))
		m.(*sync.Map).Store(sub, true)
		log.Debugf("Subscribe topic = %s, sub = %v", topic.String(), sub)
	}
}

// Unsubscribe a subscription from event collector, won't receive new event, but can still read from this subscription
func (ec *EventCollector) Unsubscribe(sub *Subscription) {
	for _, topic := range sub.topics {
		m, ok := ec.subMap.Load(topic)
		if ok && m != nil {
			m.(*sync.Map).Delete(sub)
			log.Debugf("Unsubscribe topic = %s, sub = %v", topic.String(), sub)
		}
	}
}

func (ec *EventCollector) deliverLoop() {
	for {
		select {
		case <-ec.quitCh:
			return
		case e := <-ec.postCh:
			topic := e.Topic
			v, ok := ec.subMap.Load(topic)
			if ok && v != nil {
				v.(*sync.Map).Range(func(key, value interface{}) bool {
					select {
					case key.(*Subscription).postCh <- e:
					default:
					}
					return true
				})
			}
		}
	}
}
