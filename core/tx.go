package core

import (
	"github.com/iost-official/prototype/vm"
	"github.com/iost-official/prototype/common"
)

type Tx struct {
	Time      int64
	Contract  vm.Contract
	Signs     []common.Signature
	Publisher []common.Signature
}

func (t *Tx) Encode() []byte {
	return nil
}
func (t *Tx) Decode([]byte) error {
	return nil
}
func (t *Tx) Hash() []byte {
	return nil
}
