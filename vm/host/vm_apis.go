package host

import (
	"github.com/iost-official/go-iost/core/contract"
	"github.com/iost-official/go-iost/core/event"
	"github.com/iost-official/go-iost/core/tx"
)

// APIDelegate ...
type APIDelegate struct {
	h  *Host
	ec *event.Collector
}

// NewAPI ...
func NewAPI(h *Host) APIDelegate {
	return APIDelegate{h: h, ec: event.GetCollector()}
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
	h.ec.Post(event.NewEvent(event.ContractReceipt, rec.Content),
		&event.Meta{ContractID: h.h.Context().Value("contract_name").(string)})
}

// Receipt ...
func (h *APIDelegate) Receipt(s string) contract.Cost {
	h.receipt(s)
	return ReceiptCost(len(s))
}
