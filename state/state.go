package state

import "github.com/iost-official/prototype/vm"

type State struct {
	key       Key
	value     Value
	lazyEvals vm.Method
}
