package v8

/*
#include "v8/vm.h"
*/
import "C"
import (
	"errors"
	"strconv"
)

// ErrInvalidDbValType error
var ErrInvalidDbValType = errors.New("invalid db value type")

//export goPut
func goPut(cSbx C.SandboxPtr, key, val *C.char, gasUsed *C.size_t) C.int {
	sbx, ok := GetSandbox(cSbx)
	if !ok {

	}

	k := C.GoString(key)
	v := C.GoString(val)

	cost := sbx.host.Put(k, v)
	*gasUsed = C.size_t(cost.Data)

	return 0
}

//export goGet
func goGet(cSbx C.SandboxPtr, key *C.char, gasUsed *C.size_t) *C.char {
	sbx, ok := GetSandbox(cSbx)
	if !ok {

	}

	k := C.GoString(key)
	val, cost := sbx.host.Get(k)

	*gasUsed = C.size_t(cost.Data)
	valStr, _ := dbValToString(val)

	return C.CString(valStr)
}

//export goDel
func goDel(cSbx C.SandboxPtr, key *C.char, gasUsed *C.size_t) C.int {
	sbx, ok := GetSandbox(cSbx)
	if !ok {

	}

	k := C.GoString(key)

	cost := sbx.host.Del(k)
	*gasUsed = C.size_t(cost.Data)

	return 0
}

//export goGlobalGet
func goGlobalGet(cSbx C.SandboxPtr, contractName, key *C.char, gasUsed *C.size_t) *C.char {
	sbx, ok := GetSandbox(cSbx)
	if !ok {

	}

	c := C.GoString(contractName)
	k := C.GoString(key)

	val, cost := sbx.host.GlobalGet(c, k)
	valStr, _ := dbValToString(val)

	*gasUsed = C.size_t(cost.Data)

	return C.CString(valStr)
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
