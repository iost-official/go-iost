package host

import (
	"github.com/iost-official/Go-IOS-Protocol/core/contract"
	"github.com/iost-official/Go-IOS-Protocol/core/new_tx"
)

type APIDelegate struct {
	ctx *Context
}

func NewAPI(ctx *Context) APIDelegate {
	return APIDelegate{ctx: ctx}
}

func (h *APIDelegate) receipt(t tx.ReceiptType, s string) {
	rec := tx.Receipt{
		Type:    tx.UserDefined,
		Content: s,
	}

	rs := h.ctx.GValue("receipts").([]tx.Receipt)
	h.ctx.GSet("receipts", append(rs, rec))
}

func (h *APIDelegate) RequireAuth(pubkey string) (ok bool, cost *contract.Cost) {
	authList := h.ctx.Value("auth_list")
	i, ok := authList.(map[string]int)[pubkey]
	return ok && i > 0, contract.NewCost(1, 1, 1)
}

func (h *APIDelegate) Receipt(s string) *contract.Cost {
	h.receipt(tx.UserDefined, s)
	return contract.NewCost(1, 1, 1)
}
