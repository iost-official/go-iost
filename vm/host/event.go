package host

import (
	"github.com/iost-official/go-iost/core/contract"
	"github.com/iost-official/go-iost/core/event"
)

// EventPoster the event handler in host
type EventPoster struct {
	h *Host
}

// NewEventPoster returns a new EventPoster instance.
func NewEventPoster(h *Host) EventPoster {
	return EventPoster{h: h}
}

// PostEvent post the event
func (p *EventPoster) PostEvent(data string) contract.Cost {
	e := event.NewEvent(event.ContractEvent, data)
	event.GetCollector().Post(e,
		&event.Meta{ContractID: p.h.Context().Value("contract_name").(string)})
	return EventCost(len(data))
}
