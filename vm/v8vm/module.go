package v8

/*
#include "v8/vm.h"
*/
import "C"
import "strings"

// Module JavaScript module.
type Module struct {
	id   string
	code string
}

var moduleReplacer = strings.NewReplacer("\\", "\\\\", "\n", "\\n", "\r", "\\r", "\"", "\\\"")

// NewModule create new Module.
func NewModule(id, code string) *Module {
	return &Module{
		id:   id,
		code: code,
	}
}

// Modules module map.
type Modules map[string]*Module

// NewModules create new Modules.
func NewModules() Modules {
	return make(Modules)
}

// Set set module to modules.
func (ms Modules) Set(m *Module) {
	ms[m.id] = m
}

// Get get module with specified id.
func (ms Modules) Get(id string) *Module {
	return ms[id]
}

// Del del module with specified id.
func (ms Modules) Del(id string) {
	delete(ms, id)
}

// requireModule get module from modules and return to c function.
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
