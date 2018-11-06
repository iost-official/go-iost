package v8

/*
#include "v8/vm.h"
*/
import "C"
import (
	"errors"
	"strconv"

	"encoding/json"
)

// ErrInvalidDbValType error
var ErrInvalidDbValType = errors.New("invalid db value type")

//export goPut
func goPut(cSbx C.SandboxPtr, key, val *C.char, gasUsed *C.size_t) C.int {
	sbx, ok := GetSandbox(cSbx)
	if !ok {
		panic("get sandbox failed.")
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
		panic("get sandbox failed.")
	}

	k := C.GoString(key)
	val, cost := sbx.host.Get(k)

	*gasUsed = C.size_t(cost.Data)

	if val == nil {
		return nil
	}

	valStr, _ := dbValToString(val)
	return C.CString(valStr)
}

//export goDel
func goDel(cSbx C.SandboxPtr, key *C.char, gasUsed *C.size_t) C.int {
	sbx, ok := GetSandbox(cSbx)
	if !ok {
		panic("get sandbox failed.")
	}

	k := C.GoString(key)

	cost := sbx.host.Del(k)
	*gasUsed = C.size_t(cost.Data)

	return 0
}

//export goMapPut
func goMapPut(cSbx C.SandboxPtr, key, field, val *C.char, gasUsed *C.size_t) C.int {
	sbx, ok := GetSandbox(cSbx)
	if !ok {
		panic("get sandbox failed.")
	}

	k := C.GoString(key)
	f := C.GoString(field)
	v := C.GoString(val)

	cost := sbx.host.MapPut(k, f, v)
	*gasUsed = C.size_t(cost.Data)

	return 0
}

//export goMapHas
func goMapHas(cSbx C.SandboxPtr, key, field *C.char, gasUsed *C.size_t) C.bool {
	sbx, ok := GetSandbox(cSbx)
	if !ok {
		panic("get sandbox failed.")
	}

	k := C.GoString(key)
	f := C.GoString(field)
	ret, cost := sbx.host.MapHas(k, f)

	*gasUsed = C.size_t(cost.Data)

	return C.bool(ret)
}

//export goMapGet
func goMapGet(cSbx C.SandboxPtr, key, field *C.char, gasUsed *C.size_t) *C.char {
	sbx, ok := GetSandbox(cSbx)
	if !ok {
		panic("get sandbox failed.")
	}

	k := C.GoString(key)
	f := C.GoString(field)
	val, cost := sbx.host.MapGet(k, f)

	*gasUsed = C.size_t(cost.Data)

	if val == nil {
		return nil
	}

	valStr, _ := dbValToString(val)
	return C.CString(valStr)
}

//export goMapDel
func goMapDel(cSbx C.SandboxPtr, key, field *C.char, gasUsed *C.size_t) C.int {
	sbx, ok := GetSandbox(cSbx)
	if !ok {
		panic("get sandbox failed.")
	}

	k := C.GoString(key)
	f := C.GoString(field)

	cost := sbx.host.MapDel(k, f)
	*gasUsed = C.size_t(cost.Data)

	return 0
}

//export goMapKeys
func goMapKeys(cSbx C.SandboxPtr, key *C.char, gasUsed *C.size_t) *C.char {
	sbx, ok := GetSandbox(cSbx)
	if !ok {
		panic("get sandbox failed.")
	}

	k := C.GoString(key)

	fstr, cost := sbx.host.MapKeys(k)
	j, err := json.Marshal(fstr)
	if err != nil {
		panic(err)
	}
	//fmt.Println("storage145", fstr)
	*gasUsed = C.size_t(cost.Data)

	return C.CString(string(j))
}

//export goGlobalGet
func goGlobalGet(cSbx C.SandboxPtr, contractName, key *C.char, gasUsed *C.size_t) *C.char {
	sbx, ok := GetSandbox(cSbx)
	if !ok {

	}

	c := C.GoString(contractName)
	k := C.GoString(key)
	val, cost := sbx.host.GlobalGet(c, k)

	*gasUsed = C.size_t(cost.Data)

	if val == nil {
		return nil
	}

	valStr, _ := dbValToString(val)
	return C.CString(valStr)
}

func dbValToString(val interface{}) (string, error) {
	switch v := val.(type) {
	case int64:
		return strconv.FormatInt(v, 10), nil
	case string:
		return v, nil
	case bool:
		return strconv.FormatBool(v), nil
	case []byte:
		return string(v), nil
	default:
		return "", ErrInvalidDbValType
	}
}
