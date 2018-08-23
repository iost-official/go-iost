package host

import (
	"github.com/iost-official/Go-IOS-Protocol/core/contract"
	"github.com/iost-official/Go-IOS-Protocol/core/event"
)

type EventPoster struct {
}

func Post(topic event.Event_Topic, data string) *contract.Cost {
	e := event.NewEvent(topic, data)
	event.GetEventCollectorInstance().Post(e)
	return EventCost(len(data))
}
