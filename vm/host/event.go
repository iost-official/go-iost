package host

import (
	"github.com/iost-official/go-iost/core/contract"
	"github.com/iost-official/go-iost/core/event"
)

// EventPoster the event handler in host
type EventPoster struct {
}

// Post post the event
func (p *EventPoster) PostEvent(data string) contract.Cost {
	e := event.NewEvent(event.Event_ContractEvent, data)
	event.GetEventCollectorInstance().Post(e)
	return EventCost(len(data))
}


