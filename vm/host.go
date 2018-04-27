package vm

import (
	"github.com/iost-official/prototype/state"
)

type Bridge struct {
}

func (b *Bridge) Args(n uint) []state.Value {
	return nil
}

func (b *Bridge) Return(rtn []state.Value) {

}

func SetPublic(name string, f func(b *Bridge)) error {
	return nil
}

func SetProtected(name string, f func(b *Bridge)) error {
	return nil
}

func SetPrivate(name string, f func(b *Bridge)) error {
	return nil
}
