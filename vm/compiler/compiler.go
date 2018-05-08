package compiler

import "github.com/iost-official/prototype/vm"

type Compiler interface {
	Compile(code string) vm.Contract
}