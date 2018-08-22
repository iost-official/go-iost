package v8

/*
#include "v8/vm.h"
*/
import "C"
import "strings"

// Module module struct
type Module struct {
	id   string
	code string
}

var moduleReplacer = strings.NewReplacer("\\", "\\\\", "\n", "\\n", "\r", "\\r", "\"", "\\\"")

// NewModule return module with id and code
func NewModule(id, code string) *Module {
	return &Module{
		id:   id,
		code: code,
	}
}

// Modules module map
type Modules map[string]*Module

// NewModules return module map
func NewModules() Modules {
	return make(Modules)
}

// Set set module by id
func (ms Modules) Set(m *Module) {
	ms[m.id] = m
}

// Get get module by id
func (ms Modules) Get(id string) *Module {
	return ms[id]
}

// Del delete module by id
func (ms Modules) Del(id string) {
	delete(ms, id)
}

//export requireModule
func requireModule(cSbx C.SandboxPtr, moduleID *C.char) *C.char {
	id := C.GoString(moduleID)

	sbx, ok := GetSandbox(cSbx)
	if !ok {

	}

	m := sbx.modules.Get(id)
	if m == nil {
		return nil
	}

	m.code = moduleReplacer.Replace(m.code)

	return C.CString(m.code)
}
