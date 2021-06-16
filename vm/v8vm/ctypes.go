package v8

/*
#include "v8/vm.h"
*/
import "C"

func newCStr(str string) C.CStr {
	cstr := C.CStr{}
	cstr.data = C.CString(str)
	cstr.size = C.int(len(str))
	return cstr
}

func (cstr *C.CStr) SetString(str string) {
	cstr.data = C.CString(str)
	cstr.size = C.int(len(str))
}

func (cstr C.CStr) GoString() string {
	return C.GoStringN(cstr.data, cstr.size)
}
