package tx

import (
	"github.com/iost-official/prototype/account"
	"github.com/iost-official/prototype/core/state"
	"github.com/iost-official/prototype/vm"
)

type Holder struct {
	self  account.Account
	pool  state.Pool
	Spool *ServiPool
}

func (h *Holder) Self() account.Account {
	return h.self
}

func (h *Holder) AddServi(tx []Tx) {

	for _, t := range tx {
		if len(t.Recorder.Pubkey) == 0 {
			continue
		}
		servi, err := h.Spool.User(vm.PubkeyToIOSTAccount(t.Recorder.Pubkey))
		if err != nil {
			servi.IncrBehavior(1)
			// balance
			val0, err := state.StdPool.GetHM("iost", state.Key(vm.PubkeyToIOSTAccount(t.Recorder.Pubkey)))
			if err != nil {
				continue
			}
			val, ok := val0.(*state.VFloat)
			if !ok {
				continue
			}

			servi.SetBalance(val.ToFloat64())
		}
	}
	h.Spool.Flush()

}

func (h *Holder) ClearServi(acc string) {

	servi, err := h.Spool.User(vm.IOSTAccount(acc))
	if err != nil {
		servi.Clear()
	}
}

func NewHolder(acc account.Account, pool state.Pool, spool *ServiPool) *Holder {
	return &Holder{acc, pool, spool}
}

type Watcher struct {
	hp *Holder
}

var Data *Holder
