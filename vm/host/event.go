package host

import (
	"github.com/iost-official/go-iost/core/contract"
	"github.com/iost-official/go-iost/core/event"
)

// EventPoster the event handler in host
type EventPoster struct {
}

// PostEvent post the event
func (p *EventPoster) PostEvent(data string) contract.Cost {
	e := event.NewEvent(event.ContractEvent, data)
	event.GetCollector().Post(e, nil)
	return EventCost(len(data))
}
