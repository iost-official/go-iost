package v8

/*
#include "v8/vm.h"
*/
import "C"

type Module struct {
	id   string
	code string
}

func NewModule(id, code string) *Module {
	return &Module{
		id:   id,
		code: code,
	}
}

type Modules map[string]*Module

func NewModules() Modules {
	return make(Modules)
}

func (ms Modules) Set(m *Module) {
	ms[m.id] = m
}

func (ms Modules) Get(id string) *Module {
	return ms[id]
}

func (ms Modules) Del(id string) {
	delete(ms, id)
}

//export requireModule
func requireModule(cSbx C.SandboxPtr, moduleId *C.char) *C.char {
	id := C.GoString(moduleId)

	sbx, ok := GetSandbox(cSbx)
	if !ok {

	}

	m := sbx.modules.Get(id)
	return C.CString(m.code)
}
