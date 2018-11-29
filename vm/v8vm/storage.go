package v8

/*
#include "v8/vm.h"
*/
import "C"
import (
	"errors"
	"strconv"

	"encoding/json"

	"github.com/iost-official/go-iost/core/contract"
)

// ErrInvalidDbValType error
var ErrInvalidDbValType = errors.New("invalid db value type")

//export goPut
func goPut(cSbx C.SandboxPtr, key, val, ramPayer *C.char, gasUsed *C.size_t) *C.char {
	sbx, ok := GetSandbox(cSbx)
	if !ok {
		return C.CString(ErrGetSandbox.Error())
	}

	k := C.GoString(key)
	v := C.GoString(val)

	var cost contract.Cost

	var err error
	if ramPayer == nil || C.GoString(ramPayer) == "" {
		cost, err = sbx.host.Put(k, v)
	} else {
		o := C.GoString(ramPayer)
		cost, err = sbx.host.Put(k, v, o)
	}
	*gasUsed = C.size_t(cost.CPU)
	if err != nil {
		return C.CString(err.Error())
	}

	return nil
}

//export goHas
func goHas(cSbx C.SandboxPtr, key, ramPayer *C.char, result *C.bool, gasUsed *C.size_t) *C.char {
	sbx, ok := GetSandbox(cSbx)
	if !ok {
		return C.CString(ErrGetSandbox.Error())
	}

	k := C.GoString(key)
	var ret bool
	var cost contract.Cost

	ret, cost = sbx.host.Has(k)

	*gasUsed = C.size_t(cost.CPU)
	*result = C.bool(ret)

	return nil
}

//export goGet
func goGet(cSbx C.SandboxPtr, key, ramPayer *C.char, result **C.char, gasUsed *C.size_t) *C.char {
	sbx, ok := GetSandbox(cSbx)
	if !ok {
		return C.CString(ErrGetSandbox.Error())
	}

	k := C.GoString(key)
	var val interface{}
	var cost contract.Cost

	val, cost = sbx.host.Get(k)

	*gasUsed = C.size_t(cost.CPU)
	if val == nil {
		*result = nil
		return nil
	}

	valStr, err := dbValToString(val)
	if err != nil {
		return C.CString(err.Error())
	}
	*result = C.CString(valStr)

	return nil
}

//export goDel
func goDel(cSbx C.SandboxPtr, key, ramPayer *C.char, gasUsed *C.size_t) *C.char {
	sbx, ok := GetSandbox(cSbx)
	if !ok {
		return C.CString(ErrGetSandbox.Error())
	}

	k := C.GoString(key)
	var cost contract.Cost

	cost, err := sbx.host.Del(k)
	*gasUsed = C.size_t(cost.CPU)

	if err != nil {
		return C.CString(err.Error())
	}

	return nil
}

//export goMapPut
func goMapPut(cSbx C.SandboxPtr, key, field, val, ramPayer *C.char, gasUsed *C.size_t) *C.char {
	sbx, ok := GetSandbox(cSbx)
	if !ok {
		return C.CString(ErrGetSandbox.Error())
	}

	k := C.GoString(key)
	f := C.GoString(field)
	v := C.GoString(val)

	var cost contract.Cost
	var err error
	if ramPayer == nil || C.GoString(ramPayer) == "" {
		cost, err = sbx.host.MapPut(k, f, v)
	} else {
		o := C.GoString(ramPayer)
		cost, err = sbx.host.MapPut(k, f, v, o)
	}
	*gasUsed = C.size_t(cost.CPU)
	if err != nil {
		return C.CString(err.Error())
	}

	return nil
}

//export goMapHas
func goMapHas(cSbx C.SandboxPtr, key, field, ramPayer *C.char, result *C.bool, gasUsed *C.size_t) *C.char {
	sbx, ok := GetSandbox(cSbx)
	if !ok {
		return C.CString(ErrGetSandbox.Error())
	}

	k := C.GoString(key)
	f := C.GoString(field)
	var cost contract.Cost
	var ret bool
	ret, cost = sbx.host.MapHas(k, f)

	*gasUsed = C.size_t(cost.CPU)
	*result = C.bool(ret)

	return nil
}

//export goMapGet
func goMapGet(cSbx C.SandboxPtr, key, field, ramPayer *C.char, result **C.char, gasUsed *C.size_t) *C.char {
	sbx, ok := GetSandbox(cSbx)
	if !ok {
		return C.CString(ErrGetSandbox.Error())
	}

	k := C.GoString(key)
	f := C.GoString(field)
	var cost contract.Cost
	var val interface{}
	val, cost = sbx.host.MapGet(k, f)

	*gasUsed = C.size_t(cost.CPU)

	if val == nil {
		*result = nil
		return nil
	}
	valStr, _ := dbValToString(val)
	*result = C.CString(valStr)

	return nil
}

//export goMapDel
func goMapDel(cSbx C.SandboxPtr, key, field, ramPayer *C.char, gasUsed *C.size_t) *C.char {
	sbx, ok := GetSandbox(cSbx)
	if !ok {
		return C.CString(ErrGetSandbox.Error())
	}

	k := C.GoString(key)
	f := C.GoString(field)

	var cost contract.Cost
	cost, err := sbx.host.MapDel(k, f)

	*gasUsed = C.size_t(cost.CPU)
	if err != nil {
		return C.CString(err.Error())
	}

	return nil
}

//export goMapKeys
func goMapKeys(cSbx C.SandboxPtr, key, ramPayer *C.char, result **C.char, gasUsed *C.size_t) *C.char {
	sbx, ok := GetSandbox(cSbx)
	if !ok {
		return C.CString(ErrGetSandbox.Error())
	}

	k := C.GoString(key)

	var cost contract.Cost
	var fstr []string
	fstr, cost = sbx.host.MapKeys(k)
	j, err := json.Marshal(fstr)
	if err != nil {
		return C.CString(err.Error())
	}
	*gasUsed = C.size_t(cost.CPU)
	*result = C.CString(string(j))

	return nil
}

//export goMapLen
func goMapLen(cSbx C.SandboxPtr, key, ramPayer *C.char, result *C.size_t, gasUsed *C.size_t) *C.char {
	sbx, ok := GetSandbox(cSbx)
	if !ok {
		return C.CString(ErrGetSandbox.Error())
	}

	k := C.GoString(key)

	var cost contract.Cost
	var len int
	len, cost = sbx.host.MapLen(k)
	*gasUsed = C.size_t(cost.CPU)
	*result = C.size_t(len)

	return nil
}

//export goGlobalHas
func goGlobalHas(cSbx C.SandboxPtr, contractName, key, ramPayer *C.char, result *C.bool, gasUsed *C.size_t) *C.char {
	sbx, ok := GetSandbox(cSbx)
	if !ok {
		return C.CString(ErrGetSandbox.Error())
	}

	c := C.GoString(contractName)
	k := C.GoString(key)
	var ret bool
	var cost contract.Cost

	ret, cost = sbx.host.GlobalHas(c, k)

	*gasUsed = C.size_t(cost.CPU)
	*result = C.bool(ret)

	return nil
}

//export goGlobalGet
func goGlobalGet(cSbx C.SandboxPtr, contractName, key, ramPayer *C.char, result **C.char, gasUsed *C.size_t) *C.char {
	sbx, ok := GetSandbox(cSbx)
	if !ok {
		return C.CString(ErrGetSandbox.Error())
	}

	c := C.GoString(contractName)
	k := C.GoString(key)

	var cost contract.Cost
	var val interface{}
	val, cost = sbx.host.GlobalGet(c, k)

	*gasUsed = C.size_t(cost.CPU)

	if val == nil {
		*result = nil
		return nil
	}
	valStr, _ := dbValToString(val)
	*result = C.CString(valStr)

	return nil
}

//export goGlobalMapHas
func goGlobalMapHas(cSbx C.SandboxPtr, contractName, key, field, ramPayer *C.char, result *C.bool, gasUsed *C.size_t) *C.char {
	sbx, ok := GetSandbox(cSbx)
	if !ok {
		return C.CString(ErrGetSandbox.Error())
	}

	c := C.GoString(contractName)
	k := C.GoString(key)
	f := C.GoString(field)
	var cost contract.Cost
	var ret bool
	ret, cost = sbx.host.GlobalMapHas(c, k, f)

	*gasUsed = C.size_t(cost.CPU)
	*result = C.bool(ret)

	return nil
}

//export goGlobalMapGet
func goGlobalMapGet(cSbx C.SandboxPtr, contractName, key, field, ramPayer *C.char, result **C.char, gasUsed *C.size_t) *C.char {
	sbx, ok := GetSandbox(cSbx)
	if !ok {
		return C.CString(ErrGetSandbox.Error())
	}

	c := C.GoString(contractName)
	k := C.GoString(key)
	f := C.GoString(field)
	var cost contract.Cost
	var val interface{}
	val, cost = sbx.host.GlobalMapGet(c, k, f)

	*gasUsed = C.size_t(cost.CPU)

	if val == nil {
		*result = nil
		return nil
	}
	valStr, _ := dbValToString(val)
	*result = C.CString(valStr)

	return nil
}

//export goGlobalMapKeys
func goGlobalMapKeys(cSbx C.SandboxPtr, contractName, key, ramPayer *C.char, result **C.char, gasUsed *C.size_t) *C.char {
	sbx, ok := GetSandbox(cSbx)
	if !ok {
		return C.CString(ErrGetSandbox.Error())
	}

	c := C.GoString(contractName)
	k := C.GoString(key)

	var cost contract.Cost
	var fstr []string
	fstr, cost = sbx.host.GlobalMapKeys(c, k)
	j, err := json.Marshal(fstr)
	if err != nil {
		return C.CString(err.Error())
	}
	*gasUsed = C.size_t(cost.CPU)
	*result = C.CString(string(j))

	return nil
}

//export goGlobalMapLen
func goGlobalMapLen(cSbx C.SandboxPtr, contractName, key, ramPayer *C.char, result *C.size_t, gasUsed *C.size_t) *C.char {
	sbx, ok := GetSandbox(cSbx)
	if !ok {
		return C.CString(ErrGetSandbox.Error())
	}

	c := C.GoString(contractName)
	k := C.GoString(key)

	var cost contract.Cost
	var len int
	len, cost = sbx.host.GlobalMapLen(c, k)
	*gasUsed = C.size_t(cost.CPU)
	*result = C.size_t(len)

	return nil
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
