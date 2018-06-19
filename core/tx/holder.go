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
		servi := h.Spool.User(vm.PubkeyToIOSTAccount(t.Recorder.Pubkey))
		if servi != nil {
			servi.IncrBehavior(1)
		}
	}
}

func (h *Holder) ClearServi(tx []*Tx) {
	for _, t := range tx {
		if len(t.Recorder.Pubkey) == 0 {
			continue
		}
		servi := h.Spool.User(vm.PubkeyToIOSTAccount(t.Recorder.Pubkey))
		if servi != nil {
			servi.Clear()
		}
	}
}

func NewHolder(acc account.Account, pool state.Pool, spool *ServiPool) *Holder {
	return &Holder{acc, pool, spool}
}

type Watcher struct {
	hp *Holder
}

var Data *Holder
