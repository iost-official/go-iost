package host

import (
	"github.com/iost-official/go-iost/core/contract"
	"github.com/iost-official/go-iost/core/event"
	"github.com/iost-official/go-iost/core/tx"
)

// APIDelegate ...
type APIDelegate struct {
	h  *Host
	ec *event.EventCollector
}

// NewAPI ...
func NewAPI(h *Host) APIDelegate {
	return APIDelegate{h: h, ec: event.GetEventCollectorInstance()}
}

func (h *APIDelegate) receipt(t tx.ReceiptType, s string) {
	rec := tx.Receipt{
		Type:    t,
		Content: s,
	}

	rs := h.h.ctx.GValue("receipts").([]tx.Receipt)
	h.h.ctx.GSet("receipts", append(rs, rec))

	topic := event.Event_ContractSystemEvent
	if t == tx.UserDefined {
		topic = event.Event_ContractUserEvent
	}
	h.ec.Post(event.NewEvent(topic, rec.Content))

}

// RequireAuth ...
func (h *APIDelegate) RequireAuth(pubkey string) (ok bool, cost *contract.Cost) {
	authList := h.h.ctx.Value("auth_list")
	i, ok := authList.(map[string]int)[pubkey]
	return ok && i > 0, RequireAuthCost
}

// Receipt ...
func (h *APIDelegate) Receipt(s string) *contract.Cost {
	h.receipt(tx.UserDefined, s)
	return ReceiptCost(len(s))
}
