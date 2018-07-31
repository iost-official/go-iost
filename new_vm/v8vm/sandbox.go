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
}

func NewSandbox(e *VM) *Sandbox {
	s := &Sandbox{
		isolate: e.isolate,
		context: C.newSandbox(e.isolate),
	}

	return s
}

func (sbx *Sandbox) Release() {
	if sbx.context != nil {
		C.releaseSandbox(sbx.context)
	}
	sbx.context = nil
}

func (sbx *Sandbox) Prepare(code, function string, args []string) string {
	return fmt.Sprintf(`var exports = {};
var module = {};

var wrapper = (function (exports, module) {
%s
});

wrapper.call(wrapper, exports, module)

var obj = new module.exports();
obj["%s"].apply(obj, %v);
`, code, function, args)
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
