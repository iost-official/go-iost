package new_vm

import (
	"github.com/iost-official/Go-IOS-Protocol/core/new_tx"
)

type Engine interface {
	Exec(tx0 tx.Tx)
}
