package host

import (
	"github.com/iost-official/go-iost/core/contract"
	"github.com/iost-official/go-iost/core/event"
)

// EventPoster the event handler in host
type EventPoster struct {
}

// Post post the event
func Post(topic event.Event_Topic, data string) contract.Cost {
	e := event.NewEvent(topic, data)
	event.GetEventCollectorInstance().Post(e)
	return EventCost(len(data))
}
