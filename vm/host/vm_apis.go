package host

import (
	"github.com/iost-official/Go-IOS-Protocol/core/contract"
	"github.com/iost-official/Go-IOS-Protocol/core/tx"
)

// APIDelegate ...
type APIDelegate struct {
	ctx *Context
}

// NewAPI ...
func NewAPI(ctx *Context) APIDelegate {
	return APIDelegate{ctx: ctx}
}

func (h *APIDelegate) receipt(t tx.ReceiptType, s string) {
	rec := tx.Receipt{
		Type:    t,
		Content: s,
	}

	rs := h.ctx.GValue("receipts").([]tx.Receipt)
	h.ctx.GSet("receipts", append(rs, rec))
}

// RequireAuth ...
func (h *APIDelegate) RequireAuth(pubkey string) (ok bool, cost *contract.Cost) {
	authList := h.ctx.Value("auth_list")
	i, ok := authList.(map[string]int)[pubkey]
	return ok && i > 0, contract.NewCost(1, 1, 1)
}

// Receipt ...
func (h *APIDelegate) Receipt(s string) *contract.Cost {
	h.receipt(tx.UserDefined, s)
	return contract.NewCost(1, 1, 1)
}
