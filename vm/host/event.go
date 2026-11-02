package host

import (
	"github.com/iost-official/go-iost/v3/core/contract"
	"github.com/iost-official/go-iost/v3/core/event"
)

// EventPoster the event handler in host
type EventPoster struct {
	h *Host
}

type pendingEvent struct {
	event *event.Event
	meta  *event.Meta
}

// NewEventPoster returns a new EventPoster instance.
func NewEventPoster(h *Host) EventPoster {
	return EventPoster{h: h}
}

// PostEvent post the event
func (p *EventPoster) PostEvent(data string) contract.Cost {
	e := event.NewEvent(event.ContractEvent, data)
	p.post(e,
		&event.Meta{ContractID: p.h.Context().Value("contract_name").(string)})
	return EventCost(len(data))
}

func (p *EventPoster) post(e *event.Event, meta *event.Meta) {
	if p.h.Rules != nil && p.h.IsFork4_10_0 {
		if p.h.CallError() == nil {
			p.h.events = append(p.h.events, pendingEvent{event: e, meta: meta})
		}
		return
	}
	event.GetCollector().Post(e, meta)
}

// FlushEvents publishes events after the transaction commits.
func (p *EventPoster) FlushEvents() {
	for _, e := range p.h.events {
		event.GetCollector().Post(e.event, e.meta)
	}
	p.ClearEvents()
}

// ClearEvents discards uncommitted events.
func (p *EventPoster) ClearEvents() {
	p.h.events = nil
}
