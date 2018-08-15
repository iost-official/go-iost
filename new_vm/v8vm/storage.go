package v8

/*
#include "v8/vm.h"
*/
import "C"
import (
	"strconv"
	"errors"
)

var ErrInvalidDbValType = errors.New("invalid db value type")

//export goPut
func goPut(cSbx C.SandboxPtr, key, val *C.char, gasUsed *C.size_t) C.int {
	sbx, ok := GetSandbox(cSbx)
	if !ok {

	}

	k := C.GoString(key)
	v := C.GoString(val)

	sbx.host.Put(k, v)

	return 0
}

//export goGet
func goGet(cSbx C.SandboxPtr, key *C.char, gasUsed *C.size_t) *C.char {
	sbx, ok := GetSandbox(cSbx)
	if !ok {

	}

	k := C.GoString(key)
	val, _ := sbx.host.Get(k)
	valStr, _ := dbValToString(val)

	return C.CString(valStr)
}

//export goDel
func goDel(cSbx C.SandboxPtr, key *C.char, gasUsed *C.size_t) C.int {
	sbx, ok := GetSandbox(cSbx)
	if !ok {

	}

	k := C.GoString(key)

	sbx.host.Del(k)

	return 0
}

func dbValToString(val interface{}) (string, error) {
	switch v := val.(type) {
	case int64:
		return strconv.FormatInt(v, 10), nil
	case string:
		return v, nil
	case nil:
		return "nil", nil
	case bool:
		return strconv.FormatBool(v), nil
	case []byte:
		return string(v), nil
	default:
		return "", ErrInvalidDbValType
	}
}