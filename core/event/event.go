package event

import (
	log "github.com/iost-official/Go-IOS-Protocol/ilog"
	"sync"
	"time"
)

const EventCollectorSize = 1024

func NewEvent(topic Event_Topic, data string) *Event {
	now := time.Now().UnixNano()
	return &Event{
		Topic: topic,
		Data:  data,
		Time:  now,
	}
}

type Subscription struct {
	topics []Event_Topic
	postCh chan<- *Event
	readCh <-chan *Event
}

func NewSubscription(chSize int, topics []Event_Topic) *Subscription {
	postCh := make(chan *Event, chSize)
	subscribe := &Subscription{
		topics: topics,
		postCh: postCh,
		readCh: postCh,
	}
	return subscribe
}

func (s *Subscription) ReadChan() <-chan *Event {
	return s.readCh
}

type EventCollector struct {
	subMap *sync.Map
	postCh chan *Event
	quitCh chan int
}

var ec *EventCollector

func GetEventCollectorInstance() *EventCollector {
	if ec == nil {
		ec = &EventCollector{
			subMap: new(sync.Map),
			postCh: make(chan *Event, EventCollectorSize),
			quitCh: make(chan int, 1),
		}
	}
	return ec
}

func (ec *EventCollector) Start() {
	go ec.deliverLoop()
}

func (ec *EventCollector) Stop() {
	ec.quitCh <- 1
}

func (ec *EventCollector) Post(e *Event) {
	select {
	case ec.postCh <- e:
	default:
		log.Debug("post event failed, aborted. topic = %s", e.Topic)
	}

}

func (ec *EventCollector) Subscribe(sub *Subscription) {
	for _, topic := range sub.topics {
		m, _ := ec.subMap.LoadOrStore(topic, new(sync.Map))
		m.(*sync.Map).Store(sub, true)
		log.Debug("Subscribe topic = %s, sub = %s", topic.String(), sub)
	}
}

func (ec *EventCollector) Unsubscribe(sub *Subscription) {
	for _, topic := range sub.topics {
		m, ok := ec.subMap.Load(topic)
		if ok && m != nil {
			m.(*sync.Map).Delete(sub)
			log.Debug("Unsubscribe topic = %s, sub = %s", topic.String(), sub)
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
