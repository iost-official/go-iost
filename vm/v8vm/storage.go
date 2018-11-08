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
func goPut(cSbx C.SandboxPtr, key, val, owner *C.char, gasUsed *C.size_t) *C.char {
	sbx, ok := GetSandbox(cSbx)
	if !ok {
		return C.Cstring(ErrGetSandbox.Error())
	}

	k := C.GoString(key)
	v := C.GoString(val)

	var cost contract.Cost

	if owner == nil {
		cost = sbx.host.Put(k, v)
	} else {
		o := C.GoString(owner)
		cost = sbx.host.Put(k, v, o)
	}
	*gasUsed = C.size_t(cost.CPU)

	return C.Cstring(MessageSuccess)
}

//export goHas
func goHas(cSbx C.SandboxPtr, key, owner *C.char, result *C.bool, gasUsed *C.size_t) *C.char {
	sbx, ok := GetSandbox(cSbx)
	if !ok {
		return C.Cstring(ErrGetSandbox.Error())
	}

	k := C.GoString(key)
	var ret bool
	var cost contract.Cost

	if owner == nil {
		ret, cost = sbx.host.Has(k)
	} else {
		o := C.GoString(owner)
		ret, cost = sbx.host.Has(k, o)
	}

	*gasUsed = C.size_t(cost.CPU)
	*result = C.bool(ret)

	return C.CString(MessageSuccess)
}

//export goGet
func goGet(cSbx C.SandboxPtr, key, owner *C.char, result **C.char, gasUsed *C.size_t) *C.char {
	sbx, ok := GetSandbox(cSbx)
	if !ok {
		return C.Cstring(ErrGetSandbox.Error())
	}

	k := C.GoString(key)
	var val interface{}
	var cost contract.Cost

	if owner == nil {
		val, cost = sbx.host.Get(k)
	} else {
		o := C.GoString(owner)
		val, cost = sbx.host.Get(k, o)
	}

	*gasUsed = C.size_t(cost.CPU)
	if val == nil {
		*result = nil
		return C.Cstring(MessageSuccess)
	}

	valStr, err := dbValToString(val)
	if err != nil {
		return C.Cstring(err.Error())
	}
	*result = C.CString(valStr)

	return C.CString(MessageSuccess)
}

//export goDel
func goDel(cSbx C.SandboxPtr, key, owner *C.char, gasUsed *C.size_t) *C.char {
	sbx, ok := GetSandbox(cSbx)
	if !ok {
		return C.Cstring(ErrGetSandbox.Error())
	}

	k := C.GoString(key)
	var cost contract.Cost

	if owner == nil {
		cost = sbx.host.Del(k)
	} else {
		o := C.GoString(owner)
		cost = sbx.host.Del(k, o)
	}
	*gasUsed = C.size_t(cost.CPU)

	return C.Cstring(MessageSuccess)
}

//export goMapPut
func goMapPut(cSbx C.SandboxPtr, key, field, val, owner *C.char, gasUsed *C.size_t) *C.char {
	sbx, ok := GetSandbox(cSbx)
	if !ok {
		return C.Cstring(ErrGetSandbox.Error())
	}

	k := C.GoString(key)
	f := C.GoString(field)
	v := C.GoString(val)

	var cost contract.Cost
	if owner == nil {
		cost = sbx.host.MapPut(k, f, v)
	} else {
		o := C.GoString(owner)
		cost = sbx.host.MapPut(k, f, v, o)
	}
	*gasUsed = C.size_t(cost.CPU)

	return C.Cstring(MessageSuccess)
}

//export goMapHas
func goMapHas(cSbx C.SandboxPtr, key, field, owner *C.char, result *C.bool, gasUsed *C.size_t) *C.char {
	sbx, ok := GetSandbox(cSbx)
	if !ok {
		return C.Cstring(ErrGetSandbox.Error())
	}

	k := C.GoString(key)
	f := C.GoString(field)
	var cost contract.Cost
	var ret bool
	if owner == nil {
		ret, cost = sbx.host.MapHas(k, f)
	} else {
		o := C.GoString(owner)
		ret, cost = sbx.host.MapHas(k, f, o)
	}

	*gasUsed = C.size_t(cost.CPU)
	*result = C.bool(ret)

	return C.Cstring(MessageSuccess)
}

//export goMapGet
func goMapGet(cSbx C.SandboxPtr, key, field, owner *C.char, result **C.char, gasUsed *C.size_t) *C.char {
	sbx, ok := GetSandbox(cSbx)
	if !ok {
		return C.Cstring(ErrGetSandbox.Error())
	}

	k := C.GoString(key)
	f := C.GoString(field)
	var cost contract.Cost
	var val interface{}
	if owner == nil {
		val, cost = sbx.host.MapGet(k, f)
	} else {
		o := C.GoString(owner)
		val, cost = sbx.host.MapGet(k, f, o)
	}

	*gasUsed = C.size_t(cost.CPU)

	if val == nil {
		*result = nil
		return C.Cstring(MessageSuccess)
	}
	valStr, _ := dbValToString(val)
	*result = C.Cstring(valStr)

	return C.CString(MessageSuccess)
}

//export goMapDel
func goMapDel(cSbx C.SandboxPtr, key, field, owner *C.char, gasUsed *C.size_t) *C.char {
	sbx, ok := GetSandbox(cSbx)
	if !ok {
		return C.Cstring(ErrGetSandbox.Error())
	}

	k := C.GoString(key)
	f := C.GoString(field)

	var cost contract.Cost
	if owner == nil {
		cost = sbx.host.MapDel(k, f)
	} else {
		o := C.GoString(owner)
		cost = sbx.host.MapDel(k, f, o)
	}

	*gasUsed = C.size_t(cost.CPU)

	return C.Cstring(MessageSuccess)
}

//export goMapKeys
func goMapKeys(cSbx C.SandboxPtr, key, owner *C.char, result **C.char, gasUsed *C.size_t) *C.char {
	sbx, ok := GetSandbox(cSbx)
	if !ok {
		return C.Cstring(ErrGetSandbox.Error())
	}

	k := C.GoString(key)

	var cost contract.Cost
	var fstr []string
	if owner == nil {
		fstr, cost = sbx.host.MapKeys(k)
	} else {
		o := C.GoString(owner)
		fstr, cost = sbx.host.MapKeys(k, o)
	}
	j, err := json.Marshal(fstr)
	if err != nil {
		return C.Cstring(err.Error())
	}
	//fmt.Println("storage145", fstr)
	*gasUsed = C.size_t(cost.CPU)
	*result = C.Cstring(string(j))

	return C.CString(MessageSuccess)
}

//export goMapLen
func goMapLen(cSbx C.SandboxPtr, key, owner *C.char, result *C.size_t, gasUsed *C.size_t) *C.char {
	sbx, ok := GetSandbox(cSbx)
	if !ok {
		return C.Cstring(ErrGetSandbox.Error())
	}

	k := C.GoString(key)

	var cost contract.Cost
	var len int
	if owner == nil {
		len, cost = sbx.host.MapLen(k)
	} else {
		o := C.GoString(owner)
		len, cost = sbx.host.MapLen(k, o)
	}
	*gasUsed = C.size_t(cost.CPU)
	*result = C.size_t(len)

	return C.CString(MessageSuccess)
}

//export goGlobalHas
func goGlobalHas(cSbx C.SandboxPtr, contractName, key, owner *C.char, result *C.bool, gasUsed *C.size_t) *C.char {
	sbx, ok := GetSandbox(cSbx)
	if !ok {
		return C.Cstring(ErrGetSandbox.Error())
	}

	c := C.GoString(contractName)
	k := C.GoString(key)
	var ret bool
	var cost contract.Cost

	if owner == nil {
		ret, cost = sbx.host.GlobalHas(c, k)
	} else {
		o := C.GoString(owner)
		ret, cost = sbx.host.GlobalHas(c, k, o)
	}

	*gasUsed = C.size_t(cost.CPU)
	*result = C.bool(ret)

	return C.CString(MessageSuccess)
}

//export goGlobalGet
func goGlobalGet(cSbx C.SandboxPtr, contractName, key, owner *C.char, result **C.char, gasUsed *C.size_t) *C.char {
	sbx, ok := GetSandbox(cSbx)
	if !ok {
		return C.Cstring(ErrGetSandbox.Error())
	}

	c := C.GoString(contractName)
	k := C.GoString(key)

	var cost contract.Cost
	var val interface{}
	if owner == nil {
		val, cost = sbx.host.GlobalGet(c, k)
	} else {
		o := C.GoString(owner)
		val, cost = sbx.host.GlobalGet(c, k, o)
	}

	*gasUsed = C.size_t(cost.CPU)

	if val == nil {
		*result = nil
		return C.Cstring(MessageSuccess)
	}
	valStr, _ := dbValToString(val)
	*result = C.Cstring(valStr)

	return C.CString(MessageSuccess)
}

//export goGlobalMapHas
func goGlobalMapHas(cSbx C.SandboxPtr, contractName, key, field, owner *C.char, result *C.bool, gasUsed *C.size_t) *C.char {
	sbx, ok := GetSandbox(cSbx)
	if !ok {
		return C.Cstring(ErrGetSandbox.Error())
	}

	c := C.GoString(contractName)
	k := C.GoString(key)
	f := C.GoString(field)
	var cost contract.Cost
	var ret bool
	if owner == nil {
		ret, cost = sbx.host.GlobalMapHas(c, k, f)
	} else {
		o := C.GoString(owner)
		ret, cost = sbx.host.GlobalMapHas(c, k, f, o)
	}

	*gasUsed = C.size_t(cost.CPU)
	*result = C.bool(ret)

	return C.Cstring(MessageSuccess)
}

//export goGlobalMapGet
func goGlobalMapGet(cSbx C.SandboxPtr, contractName, key, field, owner *C.char, result **C.char, gasUsed *C.size_t) *C.char {
	sbx, ok := GetSandbox(cSbx)
	if !ok {
		return C.Cstring(ErrGetSandbox.Error())
	}

	c := C.GoString(contractName)
	k := C.GoString(key)
	f := C.GoString(field)
	var cost contract.Cost
	var val interface{}
	if owner == nil {
		val, cost = sbx.host.GlobalMapGet(c, k, f)
	} else {
		o := C.GoString(owner)
		val, cost = sbx.host.GlobalMapGet(c, k, f, o)
	}

	*gasUsed = C.size_t(cost.CPU)

	if val == nil {
		*result = nil
		return C.Cstring(MessageSuccess)
	}
	valStr, _ := dbValToString(val)
	*result = C.Cstring(valStr)

	return C.CString(MessageSuccess)
}

//export goGlobalMapKeys
func goGlobalMapKeys(cSbx C.SandboxPtr, contractName, key, owner *C.char, result **C.char, gasUsed *C.size_t) *C.char {
	sbx, ok := GetSandbox(cSbx)
	if !ok {
		return C.Cstring(ErrGetSandbox.Error())
	}

	c := C.GoString(contractName)
	k := C.GoString(key)

	var cost contract.Cost
	var fstr []string
	if owner == nil {
		fstr, cost = sbx.host.GlobalMapKeys(c, k)
	} else {
		o := C.GoString(owner)
		fstr, cost = sbx.host.GlobalMapKeys(c, k, o)
	}
	j, err := json.Marshal(fstr)
	if err != nil {
		return C.Cstring(err.Error())
	}
	*gasUsed = C.size_t(cost.CPU)
	*result = C.Cstring(string(j))

	return C.CString(MessageSuccess)
}

//export goGlobalMapLen
func goGlobalMapLen(cSbx C.SandboxPtr, contractName, key, owner *C.char, result *C.size_t, gasUsed *C.size_t) *C.char {
	sbx, ok := GetSandbox(cSbx)
	if !ok {
		return C.Cstring(ErrGetSandbox.Error())
	}

	c := C.GoString(contractName)
	k := C.GoString(key)

	var cost contract.Cost
	var len int
	if owner == nil {
		len, cost = sbx.host.GlobalMapLen(c, k)
	} else {
		o := C.GoString(owner)
		len, cost = sbx.host.GlobalMapLen(c, k, o)
	}
	*gasUsed = C.size_t(cost.CPU)
	*result = C.size_t(len)

	return C.CString(MessageSuccess)
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
