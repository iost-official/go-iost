package host

import (
	"github.com/iost-official/Go-IOS-Protocol/core/contract"
	"github.com/iost-official/Go-IOS-Protocol/core/tx"
)

// APIDelegate ...
type APIDelegate struct {
	h *Host
}

// NewAPI ...
func NewAPI(h *Host) APIDelegate {
	return APIDelegate{h: h}
}

func (h *APIDelegate) receipt(t tx.ReceiptType, s string) {
	rec := tx.Receipt{
		Type:    t,
		Content: s,
	}

	rs := h.h.ctx.GValue("receipts").([]tx.Receipt)
	h.h.ctx.GSet("receipts", append(rs, rec))
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
