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

func (h *APIDelegate) receipt(s string) {
	fn := h.h.Context().Value("contract_name").(string) + "/" + h.h.Context().Value("abi_name").(string)
	rec := &tx.Receipt{
		FuncName: fn,
		Content:  s,
	}

	rs := h.h.ctx.GValue("receipts").([]*tx.Receipt)
	h.h.ctx.GSet("receipts", append(rs, rec))

	// post event for receipt
	topic := event.Event_ContractReceipt
	h.ec.Post(event.NewEvent(topic, rec.Content))
}

// Receipt ...
func (h *APIDelegate) Receipt(s string) *contract.Cost {
	h.receipt(s)
	return ReceiptCost(len(s))
}
