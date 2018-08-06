package v8

/*
#include <stdlib.h>
#include "v8/vm.h"
*/
import "C"
import (
	"errors"
	"fmt"
	"unsafe"
)

// A Sandbox is an execution environment that allows separate, unrelated, JavaScript
// code to run in a single instance of IVM.
type Sandbox struct {
	id      int
	isolate C.IsolatePtr
	context C.SandboxPtr
	modules Modules
}

var sbxMap = make(map[C.SandboxPtr]*Sandbox)

func GetSandbox(cSbx C.SandboxPtr) (*Sandbox, bool) {
	sbx, ok := sbxMap[cSbx]
	return sbx, ok
}

func NewSandbox(e *VM) *Sandbox {
	cSbx := C.newSandbox(e.isolate)
	s := &Sandbox{
		isolate: e.isolate,
		context: cSbx,
		modules: NewModules(),
	}
	sbxMap[cSbx] = s

	return s
}

func (sbx *Sandbox) Release() {
	if sbx.context != nil {
		delete(sbxMap, sbx.context)
		C.releaseSandbox(sbx.context)
	}
	sbx.context = nil
}

func (sbx *Sandbox) Init() {
	// init require
}

func (sbx *Sandbox) SetModule(name, code string) {
	if name == "" || code == "" {
		return
	}
	m := NewModule(name, code)
	sbx.modules.Set(m)
}

func (sbx *Sandbox) Prepare(code, function string, args []string) string {
	sbx.SetModule("_native_main", code)
	return fmt.Sprintf(`
var _native_main = NativeModule.require('_native_main');
var obj = new _native_main();
obj['%s'].apply(obj, %v);
`, function, args)
}

func (sbx *Sandbox) Execute(preparedCode string) (string, error) {
	cCode := C.CString(preparedCode)
	defer C.free(unsafe.Pointer(cCode))

	rs := C.Execute(sbx.context, cCode)

	result := C.GoString(rs.Value)

	var err error
	if rs.Err != nil {
		err = errors.New(C.GoString(rs.Err))
	}

	C.free(unsafe.Pointer(rs.Value))
	C.free(unsafe.Pointer(rs.Err))

	return result, err
}
